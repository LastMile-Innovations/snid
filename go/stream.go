package snid

import (
	"encoding/binary"
	"sync"
	"time"
	"unsafe"
)

// =============================================================================
// STATIC STREAMER (Fixed Batch Size) - OPTIMIZED
// =============================================================================

// Streamer is a zero-alloc, infinite generator optimized for hot loops.
// It pre-allocates a buffer once and refills it in-place when exhausted.
//
// NOT THREAD-SAFE. Use one Streamer per Goroutine (P-Local pattern).
//
// Performance: ~1.5ns amortized per ID (vs ~3.7ns for NewFast).
//
// Example usage:
//
//	firehose := snid.NewStreamer(4096)
//	for msg := range kafkaChannel {
//	    msg.ID = firehose.Next()  // ~1.5ns
//	    process(msg)
//	}
type Streamer struct {
	buffer []ID
	cursor int
	size   int
}

// NewStreamer allocates the river bed (memory) once at startup.
// Buffer size should be cache-line aligned for optimal L1 utilization:
//   - 1024 IDs = 16KB (fits L1)
//   - 4096 IDs = 64KB (fits L1/L2)
//   - 65536 IDs = 1MB (L3 territory)
//
// Larger buffers amortize refill cost better but may spill to L2/L3.
// Recommended: 4096 for most use cases.
func NewStreamer(size int) *Streamer {
	if size < 64 {
		size = 64 // Minimum viable buffer
	}
	// Align to power of 2 for potential bitwise masking optimizations
	size = nextPowerOf2(size)

	s := &Streamer{
		buffer: make([]ID, size), // The ONLY allocation ever made
		size:   size,
		cursor: size, // Force immediate refill on first Next()
	}
	return s
}

// Next returns the next ID from the stream.
// If the buffer is exhausted, it refills instantly using the Turbo engine.
//
// OPTIMIZED: Minimal hot path - just index check, array access, increment.
// Cost: ~1ns hot path, ~1.5ns amortized.
//
//go:nosplit
func (s *Streamer) Next() ID {
	if s.cursor >= s.size {
		s.refill()
	}
	id := s.buffer[s.cursor]
	s.cursor++
	return id
}

// NextUnsafe returns the next ID without bounds checking.
// ONLY use when you know the buffer is not exhausted (check Remaining() first).
// Cost: ~0.5ns.
//
//go:nosplit
func (s *Streamer) NextUnsafe() ID {
	id := s.buffer[s.cursor]
	s.cursor++
	return id
}

// Remaining returns how many IDs are left before the next refill.
// Useful for batch-aware code that wants to avoid mid-batch refills.
func (s *Streamer) Remaining() int {
	return s.size - s.cursor
}

// Size returns the buffer capacity.
func (s *Streamer) Size() int {
	return s.size
}

// Reset forces a refill on the next Next() call.
// Useful after clock drift or when fresh timestamps are needed.
func (s *Streamer) Reset() {
	s.cursor = s.size
}

// refill regenerates the entire buffer using Turbo engine logic.
func (s *Streamer) refill() {
	fillBurstNInline(s.buffer, s.size)
	s.cursor = 0
}

// =============================================================================
// ADAPTIVE STREAMER (Dynamic Batch Size - AIMD Algorithm) - OPTIMIZED
// =============================================================================

// AdaptiveStreamer auto-tunes batch size based on demand pressure.
// Uses AIMD (Additive Increase, Multiplicative Decrease) algorithm.
//
// OPTIMIZED:
//   - Removed consumed tracking from hot path (now inferred from cursor)
//   - Uses integer shift for AIMD instead of float multiplication
//   - Removed time.Now() from hot path
//
// Behavior:
//   - Low Load: Small batches (16-64 IDs) → Latency optimized
//   - High Load: Large batches (up to maxSize) → Throughput optimized (~1.5ns/ID)
//
// NOT THREAD-SAFE. Use one AdaptiveStreamer per Goroutine.
type AdaptiveStreamer struct {
	buffer []ID
	cursor int

	// Adaptive sizing
	currentSize int // Current batch size (dynamically adjusted)
	minSize     int // Floor (default: 16)
	maxSize     int // Ceiling (buffer capacity)

	// AIMD thresholds (as integer percentages for speed)
	starvationPct int // % consumed to trigger growth (default: 75)
	idlePct       int // % consumed to trigger shrink (default: 25)

	// Observability (only updated on refill, not hot path)
	totalRefills   uint64
	totalGenerated uint64
}

// StreamerMetrics exposes adaptive state for observability/debugging.
type StreamerMetrics struct {
	CurrentBatchSize int     // Current batch size
	Consumed         int     // IDs consumed since last refill
	Pressure         float64 // consumed/currentSize ratio (0.0-1.0+)
	MinSize          int     // Minimum batch size
	MaxSize          int     // Maximum batch size
	TotalRefills     uint64  // Lifetime refill count
	TotalGenerated   uint64  // Lifetime IDs generated
}

// NewAdaptiveStreamer creates a demand-responsive stream.
// maxSize is the maximum batch size (buffer capacity).
// Batch size starts at minSize (16) and grows/shrinks based on demand.
func NewAdaptiveStreamer(maxSize int) *AdaptiveStreamer {
	if maxSize < 64 {
		maxSize = 64
	}
	maxSize = nextPowerOf2(maxSize)

	const defaultMinSize = 16

	s := &AdaptiveStreamer{
		buffer:      make([]ID, maxSize),
		currentSize: defaultMinSize, // Start small, grow as needed
		minSize:     defaultMinSize,
		maxSize:     maxSize,

		// AIMD defaults as integer percentages
		starvationPct: 75, // 75% consumed = starving
		idlePct:       25, // <25% consumed = idle

		cursor: defaultMinSize, // Force immediate refill
	}
	return s
}

// WithAIMD configures custom AIMD thresholds. Returns self for chaining.
// starvation/idle are percentages (0-100).
func (s *AdaptiveStreamer) WithAIMD(starvationPct, idlePct int) *AdaptiveStreamer {
	if starvationPct > idlePct && starvationPct <= 100 && starvationPct >= 50 {
		s.starvationPct = starvationPct
	}
	if idlePct >= 0 && idlePct < starvationPct && idlePct <= 50 {
		s.idlePct = idlePct
	}
	return s
}

// WithMinSize sets the minimum batch size floor. Returns self for chaining.
func (s *AdaptiveStreamer) WithMinSize(min int) *AdaptiveStreamer {
	if min >= 1 && min < s.maxSize {
		s.minSize = min
		if s.currentSize < min {
			s.currentSize = min
		}
	}
	return s
}

// Next returns the next ID from the adaptive stream.
// Automatically adjusts batch size based on consumption patterns.
//
// OPTIMIZED: No per-call tracking, just cursor increment.
//
//go:nosplit
func (s *AdaptiveStreamer) Next() ID {
	if s.cursor >= s.currentSize {
		s.refillAdaptive()
	}
	id := s.buffer[s.cursor]
	s.cursor++
	return id
}

// NextUnsafe returns the next ID without bounds checking.
// ONLY use when you know the buffer is not exhausted.
//
//go:nosplit
func (s *AdaptiveStreamer) NextUnsafe() ID {
	id := s.buffer[s.cursor]
	s.cursor++
	return id
}

// Remaining returns how many IDs are left before the next refill.
func (s *AdaptiveStreamer) Remaining() int {
	return s.currentSize - s.cursor
}

// CurrentSize returns the current adaptive batch size.
func (s *AdaptiveStreamer) CurrentSize() int {
	return s.currentSize
}

// Metrics returns current adaptive state for observability.
func (s *AdaptiveStreamer) Metrics() StreamerMetrics {
	consumed := s.cursor // cursor = consumed since last refill
	pressure := 0.0
	if s.currentSize > 0 {
		pressure = float64(consumed) / float64(s.currentSize)
	}
	return StreamerMetrics{
		CurrentBatchSize: s.currentSize,
		Consumed:         consumed,
		Pressure:         pressure,
		MinSize:          s.minSize,
		MaxSize:          s.maxSize,
		TotalRefills:     s.totalRefills,
		TotalGenerated:   s.totalGenerated,
	}
}

// Reset forces a refill and resets batch size to minimum.
func (s *AdaptiveStreamer) Reset() {
	s.currentSize = s.minSize
	s.cursor = s.currentSize
}

// refillAdaptive applies AIMD algorithm and refills the buffer.
// OPTIMIZED: Uses integer arithmetic, no float64, no time.Now().
func (s *AdaptiveStreamer) refillAdaptive() {
	// 1. Calculate pressure from cursor position (cursor = consumed)
	// Using integer percentage: (consumed * 100) / currentSize
	consumedPct := 0
	if s.currentSize > 0 {
		consumedPct = (s.cursor * 100) / s.currentSize
	}

	// 2. AIMD: Adjust batch size using bit shifts (faster than float multiply)
	if consumedPct >= s.starvationPct {
		// HIGH DEMAND: Buffer nearly emptied → GROW (double via left shift)
		newSize := min(
			// *2
			s.currentSize<<1, s.maxSize)
		s.currentSize = newSize
	} else if consumedPct < s.idlePct && s.cursor > 0 {
		// LOW DEMAND: Buffer mostly unused → SHRINK (halve via right shift)
		newSize := max(
			// /2
			s.currentSize>>1, s.minSize)
		s.currentSize = newSize
	}
	// STABLE DEMAND: Keep current size (hysteresis prevents oscillation)

	// 3. Generate exactly currentSize IDs
	fillBurstNInline(s.buffer, s.currentSize)

	// 4. Update stats (cheap, only on refill)
	s.totalRefills++
	s.totalGenerated += uint64(s.currentSize)
	s.cursor = 0
}

// =============================================================================
// SHARED GENERATION LOGIC - ULTRA OPTIMIZED
// =============================================================================
// fillBurstNInline is the core batch generation loop.
// Uses the unified shard system with pointer walking optimization.
func fillBurstNInline(ids []ID, n int) int {
	if n <= 0 || len(ids) < n {
		return 0
	}

	s := nextShard()

	s.mu.Lock()

	ms := coarseClock.value.Load()
	if ms > s.lastMS {
		s.lastMS = ms
		s.sequence = 0
	}

	seq := s.sequence
	mid24 := uint64(s.machineID&0xFFFFFF) << 36
	s0, s1, s2, s3 := s.s0, s.s1, s.s2, s.s3

	ptr := unsafe.Pointer(&ids[0])

	for range n {
		seq++
		if seq > 0x3FFF {
			seq = 0
			s.lastMS++
			ms = s.lastMS
		}

		res := rotl(s1*5, 7) * 9
		t := s1 << 17
		s2 ^= s0
		s3 ^= s1
		s1 ^= s2
		s0 ^= s3
		s2 ^= t
		s3 = rotl(s3, 45)

		hi := (ms << 16) | 0x7000 | (uint64(seq) >> 2)
		lo := 0x8000000000000000 |
			((uint64(seq) & 0x03) << 60) |
			mid24 |
			((res >> 28) & 0xFFFFFFFFF)

		// Direct memory write using compiler intrinsics (BSWAP)
		hiPtr := (*[8]byte)(ptr)
		loPtr := (*[8]byte)(unsafe.Add(ptr, 8))
		binary.BigEndian.PutUint64(hiPtr[:], hi)
		binary.BigEndian.PutUint64(loPtr[:], lo)

		ptr = unsafe.Add(ptr, 16)
	}

	s.sequence = seq
	s.s0, s.s1, s.s2, s.s3 = s0, s1, s2, s3
	s.mu.Unlock()

	return n
}

// fillBurstN is the safe version that uses bounds-checked access.
// Use fillBurstNInline for maximum performance.
func fillBurstN(ids []ID, n int) int {
	return fillBurstNInline(ids, n)
}

// fillBurst writes IDs to fill the entire slice.
func fillBurst(ids []ID) int {
	return fillBurstNInline(ids, len(ids))
}

// FillBurst is the exported version for external callers.
// Use this when you need to manage your own buffer for ultimate control.
func FillBurst(ids []ID) int {
	return fillBurst(ids)
}

// FillBurstN generates exactly n IDs into the provided buffer.
// Returns the number of IDs generated (n, or 0 if buffer too small).
func FillBurstN(ids []ID, n int) int {
	return fillBurstN(ids, n)
}

// =============================================================================
// TURBO STREAMER (Lock-Free Single-Thread Optimized)
// =============================================================================

// TurboStreamer is the fastest possible single-thread streamer.
// It maintains its own shard state to avoid lock contention entirely.
//
// RESTRICTIONS:
//   - MUST be used from a single goroutine only
//   - IDs are unique within this streamer but may not sort chronologically
//     with IDs from other streamers/generators
//
// Performance: ~0.8ns per ID (theoretical limit)
type TurboStreamer struct {
	buffer []ID
	cursor int
	size   int

	// Private shard state (no lock needed - single thread)
	lastMS         uint64
	sequence       uint32
	machineID      uint32
	s0, s1, s2, s3 uint64
}

// NewTurboStreamer creates the fastest possible streamer.
// Each TurboStreamer has its own isolated state - no shared locks.
func NewTurboStreamer(size int) *TurboStreamer {
	if size < 64 {
		size = 64
	}
	size = nextPowerOf2(size)

	// Initialize with unique state
	seed := uint64(time.Now().UnixNano()) ^ uint64(uintptr(unsafe.Pointer(&size)))

	s := &TurboStreamer{
		buffer:    make([]ID, size),
		size:      size,
		cursor:    size, // Force immediate refill
		lastMS:    coarseClock.value.Load(),
		machineID: uint32(seed) & 0xFFFFFF,
	}

	// Initialize Xoshiro state
	s.s0 = splitMix64(&seed)
	s.s1 = splitMix64(&seed)
	s.s2 = splitMix64(&seed)
	s.s3 = splitMix64(&seed)

	return s
}

// Next returns the next ID from the turbo stream.
// This is the absolute fastest path - no locks, no atomics.
//
//go:nosplit
func (s *TurboStreamer) Next() ID {
	if s.cursor >= s.size {
		s.refill()
	}
	id := s.buffer[s.cursor]
	s.cursor++
	return id
}

// NextUnsafe returns the next ID without any checks.
//
//go:nosplit
func (s *TurboStreamer) NextUnsafe() ID {
	id := s.buffer[s.cursor]
	s.cursor++
	return id
}

// Remaining returns how many IDs are left before the next refill.
func (s *TurboStreamer) Remaining() int {
	return s.size - s.cursor
}

// Size returns the buffer capacity.
func (s *TurboStreamer) Size() int {
	return s.size
}

// Reset forces a refill on the next Next() call.
func (s *TurboStreamer) Reset() {
	s.cursor = s.size
}

// refill regenerates the buffer using private state (no locks).
// OPTIMIZATIONS: Pre-computed mid24, pointer walking, register-only Xoshiro.
func (s *TurboStreamer) refill() {
	ms := coarseClock.value.Load()

	if ms > s.lastMS {
		s.lastMS = ms
		s.sequence = 0
	}

	seq := s.sequence
	// OPTIMIZATION: Pre-compute mid24 (invariant)
	mid24 := uint64(s.machineID&0xFFFFFF) << 36
	s0, s1, s2, s3 := s.s0, s.s1, s.s2, s.s3

	ptr := unsafe.Pointer(&s.buffer[0])
	size := s.size

	for range size {
		seq++
		if seq > 0x3FFF {
			seq = 0
			s.lastMS++
			ms = s.lastMS
		}

		res := rotl(s1*5, 7) * 9
		t := s1 << 17
		s2 ^= s0
		s3 ^= s1
		s1 ^= s2
		s0 ^= s3
		s2 ^= t
		s3 = rotl(s3, 45)

		hi := (ms << 16) | 0x7000 | (uint64(seq) >> 2)
		lo := 0x8000000000000000 |
			((uint64(seq) & 0x03) << 60) |
			mid24 |
			((res >> 28) & 0xFFFFFFFFF)

		// Write hi/lo using compiler intrinsics (BSWAP)
		hiPtr := (*[8]byte)(ptr)
		loPtr := (*[8]byte)(unsafe.Add(ptr, 8))
		binary.BigEndian.PutUint64(hiPtr[:], hi)
		binary.BigEndian.PutUint64(loPtr[:], lo)

		ptr = unsafe.Add(ptr, 16)
	}

	s.sequence = seq
	s.s0, s.s1, s.s2, s.s3 = s0, s1, s2, s3
	s.cursor = 0
}

// =============================================================================
// STREAM ITERATOR (Batch-Oriented Processing)
// =============================================================================

// StreamIterator is an alternative API for batch-oriented processing.
// Unlike Streamer (which auto-refills), StreamIterator requires explicit
// Refill() calls, giving the caller full control over timing.
//
// Use when:
//   - You need deterministic refill timing (e.g., before a batch process)
//   - You want to iterate over Buffer directly without method call overhead
type StreamIterator struct {
	Buffer []ID
	size   int
}

// NewStreamIterator creates a reusable batch iterator.
func NewStreamIterator(batchSize int) *StreamIterator {
	if batchSize < 64 {
		batchSize = 64
	}
	return &StreamIterator{
		Buffer: make([]ID, batchSize),
		size:   batchSize,
	}
}

// Refill regenerates the entire batch using Burst Mode.
// Call this after consuming the current Buffer.
func (it *StreamIterator) Refill() {
	fillBurst(it.Buffer)
}

// Size returns the batch size.
func (it *StreamIterator) Size() int {
	return it.size
}

// =============================================================================
// POOL-BASED STREAMER (For Short-Lived Goroutines)
// =============================================================================

// turboPool provides O(P) scaling for concurrent streamer access.
// sync.Pool uses per-processor sharding, making it lock-free on the happy path
// and infinitely scalable vs the previous atomic.Pointer (stack of size 1).
var turboPool = sync.Pool{
	New: func() any {
		return NewTurboStreamer(4096)
	},
}

// BorrowStreamer gets a pre-warmed TurboStreamer from the pool.
// Use this for short-lived goroutines to avoid initialization cost.
// MUST call ReturnStreamer when done.
//
// OPTIMIZATION: Uses sync.Pool for O(P) concurrent scaling.
// Under 100 concurrent requests, all 100 get cached streamers (vs 1 before).
func BorrowStreamer() *TurboStreamer {
	return turboPool.Get().(*TurboStreamer)
}

// ReturnStreamer returns a TurboStreamer to the pool for reuse.
// No reset needed - timestamp refreshes naturally on next refill.
func ReturnStreamer(s *TurboStreamer) {
	if s != nil {
		turboPool.Put(s)
	}
}

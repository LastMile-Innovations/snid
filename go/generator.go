package snid

import (
	"encoding/binary"
	"hash/maphash"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// =============================================================================
// SNID v5.3 FINAL - SIMPLIFIED HIGH-PERFORMANCE GENERATOR
// =============================================================================
//
// Performance Targets:
//   - NewFast():            ~3.7ns  (single ID, thread-safe)
//   - TurboStreamer.Next(): ~1.7ns  (hot loop, single-thread)
//   - NewBurst(1000):       ~2μs    (batch mode)
//
// Architecture:
//   - Lock-free per-P state for NewFast()
//   - Fixed shared shards for projected/batch generation
//   - Elastic pool for streamers and batch operations
//   - Virtual time advancement (no spin-waiting)
//   - Adaptive coarse clock (10ms-500μs based on CPU)
//
// =============================================================================

// shard is a thread-safe ID generator with minimal state.
type shard struct {
	_ [64]byte

	mu             sync.Mutex
	lastMS         uint64
	sequence       uint32
	machineID      uint32
	s0, s1, s2, s3 uint64 // Xoshiro256** state

	_ [64]byte
}

// fastPShard is the lock-free per-P state used by NewFast.
// Access is protected by runtime proc pinning, so no mutex is required.
type fastPShard struct {
	_ [64]byte

	lastMS         uint64
	sequence       uint32
	machineID      uint32
	s0, s1, s2, s3 uint64

	_ [64]byte
}

// Package-level state
var (
	shards        []*shard  // Fixed array for NewFast()
	shardMask     uint64    // Mask for round-robin
	perPCursors   []pCursor // Per-P round-robin cursors (no contention)
	perPCursorLen int       // Cached len(perPCursors) for hot path
	fastPShards   []fastPShard
	fastPShardLen int
	noPinRR       atomic.Uint64
)

type shardCursor struct {
	index uint64
}

type pCursor struct {
	index uint64
	_     [56]byte
}

func nextShard() *shard {
	s, _ := nextShardWithIndex()
	return s
}

func nextShardWithIndex() (*shard, uint64) {
	if !runtimePinEnabled {
		idx := noPinRR.Add(1) & shardMask
		return shards[idx], idx
	}

	p := procPin()
	if p >= 0 && p < perPCursorLen {
		cur := &perPCursors[p]
		idx := cur.index & shardMask
		cur.index++
		procUnpin()
		return shards[idx], idx
	}
	procUnpin()

	// Fallback when GOMAXPROCS changes after init.
	idx := uint64(p) & shardMask
	return shards[idx], idx
}

// init implements the corresponding operation.
func init() {
	initShards()
}

// initShards implements the corresponding operation.
func initShards() {
	// Scale shards: 4 per core, min 8, max 256
	numCPU := runtime.GOMAXPROCS(0)
	maxCPU := runtime.NumCPU()
	pSlots := max(max(numCPU, maxCPU), 1)
	n := min(max(nextPowerOf2(numCPU*4), 8), 256)

	shards = make([]*shard, n)
	shardMask = uint64(n - 1)
	perPCursorLen = pSlots
	perPCursors = make([]pCursor, perPCursorLen)
	fastPShardLen = pSlots
	fastPShards = make([]fastPShard, fastPShardLen)

	// Seed each shard from process-unique entropy via maphash.
	pid := uint64(os.Getpid())
	startNanos := uint64(time.Now().UnixNano())
	cursorSeed := startNanos ^ pid
	noPinRR.Store(cursorSeed)
	for i := 0; i < perPCursorLen; i++ {
		perPCursors[i].index = splitMix64(&cursorSeed)
	}

	processSeed := maphash.MakeSeed()
	var hasher maphash.Hash
	hasher.SetSeed(processSeed)
	var entropy [24]byte
	binary.LittleEndian.PutUint64(entropy[0:8], startNanos)
	binary.LittleEndian.PutUint64(entropy[8:16], pid)
	nowMS := coarseClock.value.Load()

	for i := 0; i < fastPShardLen; i++ {
		hasher.Reset()
		// High bit segregates NewFast per-P stream from shared shard stream.
		binary.LittleEndian.PutUint64(entropy[16:24], (uint64(i) | (1 << 63)))
		_, _ = hasher.Write(entropy[:])
		seed := hasher.Sum64()
		if seed == 0 {
			seed = uint64(i + 1)
		}
		s := &fastPShards[i]
		s.lastMS = nowMS
		s.machineID = (uint32(splitMix64(&seed)) ^ uint32(pid) ^ uint32(i)) & 0xFFFFFF
		s.s0 = splitMix64(&seed)
		s.s1 = splitMix64(&seed)
		s.s2 = splitMix64(&seed)
		s.s3 = splitMix64(&seed)
	}

	for i := 0; i < n; i++ {
		hasher.Reset()
		binary.LittleEndian.PutUint64(entropy[16:24], uint64(i))
		_, _ = hasher.Write(entropy[:])
		seed := hasher.Sum64()
		if seed == 0 {
			seed = uint64(i + 1)
		}

		s := &shard{
			lastMS:    nowMS,
			machineID: (uint32(splitMix64(&seed)) ^ uint32(pid)) & 0xFFFFFF,
		}
		s.s0 = splitMix64(&seed)
		s.s1 = splitMix64(&seed)
		s.s2 = splitMix64(&seed)
		s.s3 = splitMix64(&seed)
		shards[i] = s
	}
}

// NewFast generates a UUID v7 with ~3.7ns latency.
// Thread-safe for concurrent use.
func NewFast() ID {
	if !runtimePinEnabled {
		return newFastShared()
	}

	p := procPin()
	if p >= 0 && p < fastPShardLen {
		s := &fastPShards[p]
		ms := coarseClock.value.Load()
		lastMS := s.lastMS
		seq := s.sequence
		mid := s.machineID
		s0, s1, s2, s3 := s.s0, s.s1, s.s2, s.s3

		if ms > lastMS {
			lastMS = ms
			seq = 0
		} else {
			seq++
			if seq > 0x3FFF {
				lastMS++
				ms = lastMS
				seq = 0
			} else {
				ms = lastMS
			}
		}

		res := rotl(s1*5, 7) * 9
		t := s1 << 17
		s2 ^= s0
		s3 ^= s1
		s1 ^= s2
		s0 ^= s3
		s2 ^= t
		s3 = rotl(s3, 45)

		s.lastMS = lastMS
		s.sequence = seq
		s.s0, s.s1, s.s2, s.s3 = s0, s1, s2, s3
		procUnpin()

		var id ID
		seq64 := uint64(seq)
		binary.BigEndian.PutUint64(id[:8], (ms<<16)|0x7000|(seq64>>2))
		binary.BigEndian.PutUint64(id[8:], 0x8000000000000000|
			((seq64&0x03)<<60)|
			((uint64(mid)&0xFFFFFF)<<36)|
			((res>>28)&0xFFFFFFFFF))
		return id
	}
	procUnpin()

	// Fallback path if GOMAXPROCS changed after init.
	return newFastShared()
}

func newFastShared() ID {
	s := nextShard()
	s.mu.Lock()
	ms := coarseClock.value.Load()
	if ms > s.lastMS {
		s.lastMS = ms
		s.sequence = 0
	} else {
		s.sequence++
		if s.sequence > 0x3FFF {
			s.lastMS++
			ms = s.lastMS
			s.sequence = 0
		} else {
			ms = s.lastMS
		}
	}
	res := rotl(s.s1*5, 7) * 9
	t := s.s1 << 17
	s.s2 ^= s.s0
	s.s3 ^= s.s1
	s.s1 ^= s.s2
	s.s0 ^= s.s3
	s.s2 ^= t
	s.s3 = rotl(s.s3, 45)
	seq, mid := uint64(s.sequence), uint64(s.machineID)
	s.mu.Unlock()

	var id ID
	binary.BigEndian.PutUint64(id[:8], (ms<<16)|0x7000|(seq>>2))
	binary.BigEndian.PutUint64(id[8:], 0x8000000000000000|
		((seq&0x03)<<60)|
		((mid&0xFFFFFF)<<36)|
		((res>>28)&0xFFFFFFFFF))
	return id
}

// NewProjected generates an ID with routing bits for multi-tenancy.
func NewProjected(tenantID string, shard uint16) ID {
	s, internalShard := nextShardWithIndex()

	s.mu.Lock()
	ms := coarseClock.value.Load()
	if ms > s.lastMS {
		s.lastMS = ms
		s.sequence = 0
	} else {
		s.sequence++
		if s.sequence > 0x3FFF {
			s.lastMS++
			ms = s.lastMS
			s.sequence = 0
		} else {
			ms = s.lastMS
		}
	}
	seq := s.sequence
	// Reserve the final 16 bits for process-local uniqueness.
	// This prevents collisions between internal shards when caller inputs match.
	internalSalt := (uint16(internalShard&0xFF) << 8) | uint16(s.machineID&0xFF)
	s.mu.Unlock()

	var id ID
	binary.BigEndian.PutUint64(id[:8], (ms<<16)|0x7000|(uint64(seq)>>2))
	binary.BigEndian.PutUint64(id[8:], 0x8000000000000000|
		((uint64(seq)&0x03)<<60)|
		((uint64(shard)&0xFFF)<<48)|
		(uint64(fnv1a(tenantID))<<16)|
		uint64(internalSalt))
	return id
}

// NewBatch generates count IDs efficiently.
func NewBatch(atom Atom, count int) []ID {
	if count <= 0 {
		return nil
	}
	ids := make([]ID, count)

	s := nextShard()

	s.mu.Lock()
	ms := coarseClock.value.Load()
	if ms > s.lastMS {
		s.lastMS = ms
		s.sequence = 0
	}

	seq := s.sequence
	mid := uint64(s.machineID)
	s0, s1, s2, s3 := s.s0, s.s1, s.s2, s.s3

	for i := range ids {
		seq++
		if seq > 0x3FFF {
			s.lastMS++
			ms = s.lastMS
			seq = 0
		}

		res := rotl(s1*5, 7) * 9
		t := s1 << 17
		s2 ^= s0
		s3 ^= s1
		s1 ^= s2
		s0 ^= s3
		s2 ^= t
		s3 = rotl(s3, 45)

		binary.BigEndian.PutUint64(ids[i][:8], (ms<<16)|0x7000|(uint64(seq)>>2))
		binary.BigEndian.PutUint64(ids[i][8:], 0x8000000000000000|
			((uint64(seq)&0x03)<<60)|
			((mid&0xFFFFFF)<<36)|
			((res>>28)&0xFFFFFFFFF))
	}

	s.sequence = seq
	s.s0, s.s1, s.s2, s.s3 = s0, s1, s2, s3
	s.mu.Unlock()

	return ids
}

// =============================================================================
// HELPERS
// =============================================================================

// splitMix64 implements the corresponding operation.
func splitMix64(seed *uint64) uint64 {
	*seed += 0x9e3779b97f4a7c15
	z := *seed
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// rotl implements the corresponding operation.
func rotl(x uint64, k int) uint64 { return (x << k) | (x >> (64 - k)) }

// fnv1a implements the corresponding operation.
func fnv1a(s string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

// fnv1a32Upper is a zero-alloc FNV-1a 32-bit hash with inline ASCII uppercasing.
func fnv1a32Upper(s string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 32
		}
		h ^= uint32(c)
		h *= 16777619
	}
	return h
}

// fnv1a64Upper is a zero-alloc FNV-1a 64-bit hash with inline ASCII uppercasing.
func fnv1a64Upper(s string) uint64 {
	h := uint64(14695981039346656037)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 32
		}
		h ^= uint64(c)
		h *= 1099511628211
	}
	return h
}

// fnv1a64Lower is a zero-alloc FNV-1a 64-bit hash with inline ASCII lowercasing.
func fnv1a64Lower(s string) uint64 {
	h := uint64(14695981039346656037)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		h ^= uint64(c)
		h *= 1099511628211
	}
	return h
}

// nextPowerOf2 implements the corresponding operation.
func nextPowerOf2(n int) int {
	if n <= 0 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	return n + 1
}

// TenantHash returns the embedded tenant hash field from the ID payload.
func (id ID) TenantHash() uint32 { return binary.BigEndian.Uint32(id[10:14]) }

// NewFastDirect creates a new fast time-ordered ID.
func NewFastDirect(atom Atom) ID { return NewFast() }

// assemble creates an ID from components
func assemble(ms uint64, seq uint64, tenant uint32, shard uint16, entropy uint32) ID {
	hi := (ms << 16) | 0x7000 | (seq >> 2)
	lo := 0x8000000000000000 |
		((seq & 0x03) << 60) |
		((uint64(shard) & 0xFFF) << 48) |
		(uint64(tenant) << 16) |
		(uint64(entropy) & 0xFFFF)

	var id ID
	binary.BigEndian.PutUint64(id[:8], hi)
	binary.BigEndian.PutUint64(id[8:], lo)
	return id
}

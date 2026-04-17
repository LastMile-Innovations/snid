package snid

import (
	"runtime"
	"sync/atomic"
	"time"
)

// =============================================================================
// SNID v5.3 - PRODUCTION COARSE CLOCK
// =============================================================================
//
// This is the heartbeat of the SNID system. It provides millisecond-resolution
// timestamps at ~0.5ns read cost (vs ~20ns for time.Now()).
//
// Key Design Decisions:
//
// 1. CACHE LINE ISOLATION
//    Heavy padding ([64]byte) ensures the atomic counter never shares a CPU
//    cache line with other variables. This prevents "false sharing" where
//    unrelated writes invalidate the clock's cache line.
//
// 2. ADAPTIVE TICK RATE
//    The ticker speed scales with available CPU:
//    - 0.1-1 vCPU:  10ms ticks  (100 wakeups/sec) - saves precious CPU
//    - 1-4 vCPU:    2ms ticks   (500 wakeups/sec)
//    - 4-16 vCPU:   1ms ticks   (1000 wakeups/sec) - UUID v7 spec
//    - 16+ vCPU:    500μs ticks (2000 wakeups/sec) - extra precision
//
// 3. LOCK-FREE READS
//    atomic.Uint64.Load() compiles to a single MOV instruction with
//    acquire semantics. No locks, no syscalls, no GC pressure.
//
// 4. GRACEFUL SHUTDOWN
//    For testing and clean service teardown, StopCoarseClock() can halt
//    the background goroutine without panics or race conditions.
//
// =============================================================================

// alignedClock uses heavy padding to isolate the hot atomic value on its own
// CPU cache line (64 bytes on modern x86/ARM). This prevents false sharing.
type alignedClock struct {
	_     [56]byte      // Padding before (56 + 8 = 64)
	value atomic.Uint64 // The hot counter (8 bytes)
	_     [56]byte      // Padding after
}

var (
	coarseClock alignedClock  // The global cached timestamp
	clockActive atomic.Bool   // Ensures exactly-once start
	clockDone   chan struct{} // Shutdown signal
)

// init implements the corresponding operation.
func init() {
	// Auto-start on package load. This ensures generators never see 0.
	startCoarseClock()
}

// startCoarseClock spins up the background ticker at the appropriate rate.
// This is called automatically on init() but can be called manually after
// StopCoarseClock() if clock restart is needed.
func startCoarseClock() {
	// CAS ensures exactly-once initialization even under concurrent calls
	if !clockActive.CompareAndSwap(false, true) {
		return // Already running
	}

	// Seed immediately with fresh time
	coarseClock.value.Store(uint64(time.Now().UnixMilli()))
	done := make(chan struct{})
	clockDone = done

	// Select tick rate based on available CPU
	tickRate := selectTickRate()

	go func() {
		ticker := time.NewTicker(tickRate)
		defer ticker.Stop()

		for {
			select {
			case t := <-ticker.C:
				// atomic.Store compiles to a single store instruction
				coarseClock.value.Store(uint64(t.UnixMilli()))
			case <-done:
				return
			}
		}
	}()
}

// selectTickRate returns the optimal tick interval based on GOMAXPROCS.
// This makes SNID a "good neighbor" on shared-CPU systems.
func selectTickRate() time.Duration {
	numCPU := runtime.GOMAXPROCS(0)
	switch {
	case numCPU <= 1:
		// 0.1-1 vCPU: Minimize wakeups (Lambda, tiny containers)
		return 10 * time.Millisecond
	case numCPU <= 4:
		// Small containers
		return 2 * time.Millisecond
	case numCPU <= 16:
		// Standard servers - UUID v7 spec compliance
		return time.Millisecond
	default:
		// High-core machines - extra precision for high throughput
		return 500 * time.Microsecond
	}
}

// StopCoarseClock stops the background ticker gracefully.
// Safe to call multiple times. Typically only used for testing.
func StopCoarseClock() {
	if !clockActive.CompareAndSwap(true, false) {
		return // Not running
	}
	done := clockDone
	if done != nil {
		close(done)
	}
}

// RestartCoarseClock restarts the clock after it was stopped.
// Primarily for testing scenarios.
func RestartCoarseClock() {
	startCoarseClock()
}

// unixMilliCoarse returns the cached timestamp (~0.5ns cost).
// This is the primary timestamp source for high-throughput ID generation.
// The timestamp may be up to [tickRate] milliseconds behind wall clock.
//
//go:nosplit
func unixMilliCoarse() uint64 {
	return coarseClock.value.Load()
}

// unixMilliFresh returns a fresh wall-clock timestamp (~20ns cost).
// Use this only when exact wall-clock time is required:
// - Security tokens with precise expiry
// - Audit logs requiring wall-clock accuracy
// - Cross-system timestamp comparisons
func unixMilliFresh() uint64 {
	return uint64(time.Now().UnixMilli())
}

// ClockDrift returns the difference between coarse and fresh timestamps.
// Useful for observability and debugging.
func ClockDrift() time.Duration {
	coarse := unixMilliCoarse()
	fresh := unixMilliFresh()
	if fresh >= coarse {
		return time.Duration(fresh-coarse) * time.Millisecond
	}
	return -time.Duration(coarse-fresh) * time.Millisecond
}

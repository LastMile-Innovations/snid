package snid

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Concurrency stress tests to identify lock contention, cache thrashing, and NUMA issues

func BenchmarkConcurrency_HighContention(b *testing.B) {
	workers := []int{10, 50, 100, 200}

	for _, w := range workers {
		b.Run(fmt.Sprintf("%d_workers", w), func(b *testing.B) {
			var wg sync.WaitGroup
			var ops atomic.Int64

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				wg.Add(w)
				for j := 0; j < w; j++ {
					go func() {
						defer wg.Done()
						_ = NewFast()
						ops.Add(1)
					}()
				}
				wg.Wait()
			}
			b.ReportMetric(float64(ops.Load())/b.Elapsed().Seconds(), "ops/sec")
		})
	}
}

func BenchmarkConcurrency_CacheLineContention(b *testing.B) {
	// Test cache line contention by having workers operate on adjacent memory
	workers := runtime.NumCPU()

	b.Run("shared_generator", func(b *testing.B) {
		var wg sync.WaitGroup
		var ops atomic.Int64

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			wg.Add(workers)
			for j := 0; j < workers; j++ {
				go func() {
					defer wg.Done()
					_ = NewFast() // Shared generator state
					ops.Add(1)
				}()
			}
			wg.Wait()
		}
		b.ReportMetric(float64(ops.Load())/b.Elapsed().Seconds(), "ops/sec")
	})

	b.Run("per_worker_generators", func(b *testing.B) {
		var wg sync.WaitGroup
		var ops atomic.Int64

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			wg.Add(workers)
			for j := 0; j < workers; j++ {
				go func() {
					defer wg.Done()
					_ = NewFast() // Per-call generation
					ops.Add(1)
				}()
			}
			wg.Wait()
		}
		b.ReportMetric(float64(ops.Load())/b.Elapsed().Seconds(), "ops/sec")
	})
}

func BenchmarkConcurrency_MemoryBarrierOverhead(b *testing.B) {
	// Measure memory barrier overhead with atomic operations
	var counter atomic.Int64

	b.Run("atomic_add", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			counter.Add(1)
		}
	})

	b.Run("atomic_compare_and_swap", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			counter.CompareAndSwap(int64(i), int64(i+1))
		}
	})

	b.Run("atomic_load", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = counter.Load()
		}
	})
}

func BenchmarkConcurrency_SustainedLoad(b *testing.B) {
	// Test sustained load over time to detect memory leaks or GC pressure
	duration := 5 * time.Second

	b.Run("sustained_5s", func(b *testing.B) {
		workers := runtime.NumCPU()
		var wg sync.WaitGroup
		var ops atomic.Int64
		stop := make(chan struct{})

		// Start workers
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-stop:
						return
					default:
						_ = NewFast()
						ops.Add(1)
					}
				}
			}()
		}

		// Run for duration
		b.ResetTimer()
		time.Sleep(duration)
		close(stop)
		wg.Wait()

		b.ReportMetric(float64(ops.Load())/duration.Seconds(), "ops/sec")
	})
}

func BenchmarkConcurrency_BatchGeneration(b *testing.B) {
	batchSizes := []int{10, 100, 1000, 10000}

	for _, batchSize := range batchSizes {
		b.Run(fmt.Sprintf("batch_%d", batchSize), func(b *testing.B) {
			workers := runtime.NumCPU()
			var wg sync.WaitGroup

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				wg.Add(workers)
				for j := 0; j < workers; j++ {
					go func() {
						defer wg.Done()
						for k := 0; k < batchSize; k++ {
							_ = NewFast()
						}
					}()
				}
				wg.Wait()
			}
			b.ReportMetric(float64(batchSize*workers*b.N)/b.Elapsed().Seconds(), "ops/sec")
		})
	}
}

func BenchmarkConcurrency_TurboStreamer(b *testing.B) {
	// Test TurboStreamer in high-contention scenarios
	// Note: TurboStreamer is NOT thread-safe, so we use one per worker
	workers := []int{1, 2, 4, 8}

	for _, w := range workers {
		b.Run(fmt.Sprintf("%d_workers", w), func(b *testing.B) {
			var wg sync.WaitGroup
			var ops atomic.Int64

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				wg.Add(w)
				for j := 0; j < w; j++ {
					go func() {
						defer wg.Done()
						streamer := NewTurboStreamer(1) // Per-worker streamer
						_ = streamer.Next()
						ops.Add(1)
					}()
				}
				wg.Wait()
			}
			b.ReportMetric(float64(ops.Load())/b.Elapsed().Seconds(), "ops/sec")
		})
	}
}

func BenchmarkConcurrency_AdaptiveStreamer(b *testing.B) {
	// Test adaptive streamer under varying load
	// Note: AdaptiveStreamer is NOT thread-safe, so we use one per worker
	workers := []int{1, 2, 4, 8}

	for _, w := range workers {
		b.Run(fmt.Sprintf("%d_workers", w), func(b *testing.B) {
			var wg sync.WaitGroup
			var ops atomic.Int64

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				wg.Add(w)
				for j := 0; j < w; j++ {
					go func() {
						defer wg.Done()
						streamer := NewAdaptiveStreamer(4096) // Per-worker streamer
						_ = streamer.Next()
						ops.Add(1)
					}()
				}
				wg.Wait()
			}
			b.ReportMetric(float64(ops.Load())/b.Elapsed().Seconds(), "ops/sec")
		})
	}
}

func BenchmarkConcurrency_ColdStart(b *testing.B) {
	// Measure first-call overhead under concurrency
	workers := runtime.NumCPU()

	b.Run("first_call", func(b *testing.B) {
		var wg sync.WaitGroup

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			wg.Add(workers)
			for j := 0; j < workers; j++ {
				go func() {
					defer wg.Done()
					_ = NewFast()
				}()
			}
			wg.Wait()
		}
	})
}

func BenchmarkConcurrency_WarmPath(b *testing.B) {
	// Measure hot path after warmup
	workers := runtime.NumCPU()

	// Warmup
	for i := 0; i < 10000; i++ {
		_ = NewFast()
	}

	b.Run("warm_path", func(b *testing.B) {
		var wg sync.WaitGroup
		var ops atomic.Int64

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			wg.Add(workers)
			for j := 0; j < workers; j++ {
				go func() {
					defer wg.Done()
					_ = NewFast()
					ops.Add(1)
				}()
			}
			wg.Wait()
		}
		b.ReportMetric(float64(ops.Load())/b.Elapsed().Seconds(), "ops/sec")
	})
}

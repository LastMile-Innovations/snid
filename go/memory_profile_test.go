package snid

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

// Memory profiling benchmarks to analyze allocation patterns, GC pressure, and stack vs heap usage

func BenchmarkMemory_HeapAllocationPattern(b *testing.B) {
	// Measure heap allocations during ID generation
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewFast()
	}

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
	b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "allocs/op")
}

func BenchmarkMemory_StackVsHeap(b *testing.B) {
	// Compare stack vs heap allocation patterns
	b.Run("stack_friendly", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var id ID
			id = NewFast()
			_ = id
		}
	})

	b.Run("heap_friendly", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			id := NewFast()
			_ = &id
		}
	})
}

func BenchmarkMemory_GCPressure(b *testing.B) {
	// Measure GC pressure under sustained load
	duration := 5 * time.Second

	b.Run("sustained_5s", func(b *testing.B) {
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		start := time.Now()
		ops := 0
		for time.Since(start) < duration {
			for i := 0; i < 1000; i++ {
				_ = NewFast()
				ops++
			}
		}

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		b.ReportMetric(float64(ops)/duration.Seconds(), "ops/sec")
		b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(ops), "bytes/op")
		b.ReportMetric(float64(m2.NumGC-m1.NumGC)/duration.Seconds(), "gc/sec")
	})
}

func BenchmarkMemory_BatchAllocation(b *testing.B) {
	batchSizes := []int{10, 100, 1000, 10000}

	for _, batchSize := range batchSizes {
		b.Run(fmt.Sprintf("batch_%d", batchSize), func(b *testing.B) {
			var m1 runtime.MemStats
			runtime.ReadMemStats(&m1)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ids := make([]ID, batchSize)
				for j := 0; j < batchSize; j++ {
					ids[j] = NewFast()
				}
				_ = ids
			}

			var m2 runtime.MemStats
			runtime.ReadMemStats(&m2)

			b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N*batchSize), "bytes/op")
		})
	}
}

func BenchmarkMemory_StreamerAllocation(b *testing.B) {
	// Test Streamer allocation pattern (should be zero-alloc after init)
	b.Run("streamer_init", func(b *testing.B) {
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewStreamer(4096)
		}

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
	})

	b.Run("streamer_next", func(b *testing.B) {
		streamer := NewStreamer(4096)
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = streamer.Next()
		}

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
	})
}

func BenchmarkMemory_TurboStreamerAllocation(b *testing.B) {
	// Test TurboStreamer allocation pattern
	b.Run("turbo_init", func(b *testing.B) {
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewTurboStreamer(1)
		}

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
	})

	b.Run("turbo_next", func(b *testing.B) {
		streamer := NewTurboStreamer(1)
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = streamer.Next()
		}

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
	})
}

func BenchmarkMemory_ExtendedFamilies(b *testing.B) {
	// Test memory allocation for extended families
	b.Run("sgid", func(b *testing.B) {
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewSpatialPrecise(37.7749, -122.4194, 9)
		}

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
	})

	b.Run("nid", func(b *testing.B) {
		base := NewFast()
		vec := make([]float32, 128)
		for i := range vec {
			vec[i] = 0.5
		}

		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = NewNeural(base, vec)
		}

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
	})

	b.Run("bid", func(b *testing.B) {
		content := []byte("test content for hashing")

		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewBIDFromContent(content)
		}

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
	})

	b.Run("eid", func(b *testing.B) {
		session := uint16(12345)

		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewEphemeral(session)
		}

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
	})
}

func BenchmarkMemory_StringConversion(b *testing.B) {
	// Test memory allocation for string conversions
	id := NewFast()

	b.Run("string_compact", func(b *testing.B) {
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = id.StringCompact()
		}

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
	})

	b.Run("string_canonical", func(b *testing.B) {
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = id.String(Matter)
		}

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
	})
}

func BenchmarkMemory_EscapeAnalysis(b *testing.B) {
	// Test escape analysis - ensure IDs don't escape to heap unnecessarily
	b.Run("no_escape", func(b *testing.B) {
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			id := NewFast()
			_ = id.Bytes() // ID stays on stack
		}

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
	})

	b.Run("escape", func(b *testing.B) {
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			id := NewFast()
			_ = &id // ID escapes to heap
		}

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
	})
}

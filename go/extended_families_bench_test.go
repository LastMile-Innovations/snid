package snid

import (
	"testing"
)

// SGID Benchmarks - Spatial operations

func BenchmarkSGID_NewSpatialPrecise(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewSpatialPrecise(37.7749, -122.4194, 9)
	}
}

func BenchmarkSGID_H3Cell(b *testing.B) {
	id := NewSpatialPrecise(37.7749, -122.4194, 9)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = id.H3Cell()
	}
}

func BenchmarkSGID_H3String(b *testing.B) {
	id := NewSpatialPrecise(37.7749, -122.4194, 9)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = id.H3String()
	}
}

func BenchmarkSGID_LatLng(b *testing.B) {
	id := NewSpatialPrecise(37.7749, -122.4194, 9)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = id.LatLng()
	}
}

func BenchmarkSGID_SpatialParent(b *testing.B) {
	id := NewSpatialPrecise(37.7749, -122.4194, 9)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = id.SpatialParent(7)
	}
}

// NID Benchmarks - Neural operations

func BenchmarkNID_NewNeural(b *testing.B) {
	base := NewFast()
	vec := make([]float32, 128)
	for i := range vec {
		vec[i] = 0.5
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewNeural(base, vec)
	}
}

func BenchmarkNID_Distance(b *testing.B) {
	base := NewFast()
	vec := make([]float32, 128)
	for i := range vec {
		vec[i] = 0.5
	}
	nid1, _ := NewNeural(base, vec)
	nid2, _ := NewNeural(base, vec)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = nid1.Distance(nid2)
	}
}

func BenchmarkNID_Similarity(b *testing.B) {
	base := NewFast()
	vec := make([]float32, 128)
	for i := range vec {
		vec[i] = 0.5
	}
	nid1, _ := NewNeural(base, vec)
	nid2, _ := NewNeural(base, vec)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = nid1.Similarity(nid2)
	}
}

func BenchmarkNID_IsSimilar(b *testing.B) {
	base := NewFast()
	vec := make([]float32, 128)
	for i := range vec {
		vec[i] = 0.5
	}
	nid1, _ := NewNeural(base, vec)
	nid2, _ := NewNeural(base, vec)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = nid1.IsSimilar(nid2, 10)
	}
}

// LID Benchmarks - Ledger operations

func BenchmarkLID_NewLID(b *testing.B) {
	prev := LID{}
	payload := []byte("test payload")
	key := []byte("test-key-32-bytes-long-1234567890")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewLID(prev, payload, key)
	}
}

func BenchmarkLID_Verify(b *testing.B) {
	prev := LID{}
	payload := []byte("test payload")
	key := []byte("test-key-32-bytes-long-1234567890")
	lid, _ := NewLID(prev, payload, key)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = lid.Verify(prev, payload, key)
	}
}

// BID Benchmarks - Content-addressable operations

func BenchmarkBID_NewBIDFromContent(b *testing.B) {
	content := []byte("test content for hashing")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewBIDFromContent(content)
	}
}

func BenchmarkBID_WireFormat(b *testing.B) {
	content := []byte("test content for hashing")
	bid := NewBIDFromContent(content)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bid.WireFormat()
	}
}

func BenchmarkBID_ParseWire(b *testing.B) {
	content := []byte("test content for hashing")
	bid := NewBIDFromContent(content)
	wire := bid.WireFormat()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseBIDWire(wire)
	}
}

// EID Benchmarks - Ephemeral operations

func BenchmarkEID_NewEphemeral(b *testing.B) {
	session := uint16(12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewEphemeral(session)
	}
}

func BenchmarkEID_Bytes(b *testing.B) {
	eid := NewEphemeral(12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eid.Bytes()
	}
}

func BenchmarkEID_Time(b *testing.B) {
	eid := NewEphemeral(12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eid.Time()
	}
}

func BenchmarkEID_String(b *testing.B) {
	eid := NewEphemeral(12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eid.String()
	}
}

// Batch operations for extended families

func BenchmarkSGID_BatchNew(b *testing.B) {
	b.Run("100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < 100; j++ {
				_ = NewSpatialPrecise(37.7749, -122.4194, 9)
			}
		}
	})
	b.Run("1000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < 1000; j++ {
				_ = NewSpatialPrecise(37.7749, -122.4194, 9)
			}
		}
	})
}

func BenchmarkNID_BatchNew(b *testing.B) {
	base := NewFast()
	vec := make([]float32, 128)
	for i := range vec {
		vec[i] = 0.5
	}
	b.Run("100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < 100; j++ {
				_, _ = NewNeural(base, vec)
			}
		}
	})
	b.Run("1000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < 1000; j++ {
				_, _ = NewNeural(base, vec)
			}
		}
	})
}

func BenchmarkEID_BatchNew(b *testing.B) {
	session := uint16(12345)
	b.Run("100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < 100; j++ {
				_ = NewEphemeral(session)
			}
		}
	})
	b.Run("1000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < 1000; j++ {
				_ = NewEphemeral(session)
			}
		}
	})
}

// Cold start benchmarks (first-call overhead)

func BenchmarkSGID_ColdStart(b *testing.B) {
	b.Run("FirstCall", func(b *testing.B) {
		b.StopTimer()
		for i := 0; i < b.N; i++ {
			b.StartTimer()
			_ = NewSpatialPrecise(37.7749, -122.4194, 9)
			b.StopTimer()
		}
	})
}

func BenchmarkNID_ColdStart(b *testing.B) {
	base := NewFast()
	vec := make([]float32, 128)
	for i := range vec {
		vec[i] = 0.5
	}
	b.Run("FirstCall", func(b *testing.B) {
		b.StopTimer()
		for i := 0; i < b.N; i++ {
			b.StartTimer()
			_, _ = NewNeural(base, vec)
			b.StopTimer()
		}
	})
}

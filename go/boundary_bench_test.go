package snid

import (
	"testing"

	"github.com/google/uuid"
)

var (
	benchAtom       Atom
	benchBIDWire    string
	benchBool       bool
	benchEID        EID
	benchID         ID
	benchIDs        []ID
	benchLLM        LLMFormatV1
	benchString     string
	benchTensorHi   int64
	benchTensorLo   int64
	benchUUID       uuid.UUID
	benchUUIDString string
)

// Performance targets (from AGENTS.md):
// - NewFast target: ~3.7ns latency (single ID, thread-safe)
// - TurboStreamer.Next target: ~1.7ns (hot loop, single-thread)
// - NewBurst target: ~2μs for 1000 IDs (batch mode)

// Industry Standard Baseline: UUIDv7
func BenchmarkUUIDv7New(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id, _ := uuid.NewV7()
		benchUUID = id
	}
}

func BenchmarkUUIDv7NewString(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id, _ := uuid.NewV7()
		benchUUIDString = id.String()
	}
}

// SNID Baseline
func BenchmarkSNIDNewFast(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchID = NewFast()
	}
}

func BenchmarkSNIDNewFastString(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchString = NewFast().String(Matter)
	}
}

func BenchmarkSNIDNewFastParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			benchID = NewFast()
		}
	})
}

func BenchmarkSNIDStringMatterParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		id := NewFast()
		for pb.Next() {
			benchString = id.String(Matter)
		}
	})
}

func BenchmarkLIDVerifyParallel(b *testing.B) {
	prev, err := NewLID(LID{}, []byte("genesis"), []byte("secret"))
	if err != nil {
		b.Fatalf("NewLID genesis: %v", err)
	}
	payload := []byte("transaction_data")
	key := []byte("secret")
	lid, err := NewLID(prev, payload, key)
	if err != nil {
		b.Fatalf("NewLID payload: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			benchBool = lid.Verify(prev, payload, key)
			if !benchBool {
				b.Fatal("verify failed")
			}
		}
	})
}

func BenchmarkBIDWireFormat(b *testing.B) {
	bid := NewBIDFromContent([]byte("file_content_buffer"))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchBIDWire = bid.WireFormat()
	}
}

func BenchmarkEIDNew(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchEID = NewEphemeral(42)
	}
}

func BenchmarkSNIDToTensorWords(b *testing.B) {
	id := NewFast()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchTensorHi, benchTensorLo = id.ToTensorWords()
	}
}

func BenchmarkSNIDToLLMFormat(b *testing.B) {
	id := NewFast()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchLLM = id.ToLLMFormat(Matter)
	}
}

func BenchmarkSNIDDeterministicIngestID(b *testing.B) {
	hash := []byte("bench-content")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchID = NewDeterministicIngestID(1700000000123, hash)
	}
}

func BenchmarkNewBurst(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchIDs = NewBurst(1000)
	}
}

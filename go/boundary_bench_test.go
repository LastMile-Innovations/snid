package snid

import "testing"

func BenchmarkSNIDNewFastParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = NewFast()
		}
	})
}

func BenchmarkSNIDStringMatterParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		id := NewFast()
		for pb.Next() {
			_ = id.String(Matter)
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
			if !lid.Verify(prev, payload, key) {
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
		_ = bid.WireFormat()
	}
}

func BenchmarkEIDNew(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewEphemeral(42)
	}
}

func BenchmarkSNIDToTensorWords(b *testing.B) {
	id := NewFast()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = id.ToTensorWords()
	}
}

func BenchmarkSNIDToLLMFormat(b *testing.B) {
	id := NewFast()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = id.ToLLMFormat(Matter)
	}
}

func BenchmarkSNIDDeterministicIngestID(b *testing.B) {
	hash := []byte("bench-content")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewDeterministicIngestID(1700000000123, hash)
	}
}

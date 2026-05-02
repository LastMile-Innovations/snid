package snidbridge

import (
	"testing"
	"time"

	"github.com/LastMile-Innovations/snhash"
)

func BenchmarkAttachSNID(b *testing.B) {
	manifest, err := snhash.HashBytes(
		[]byte("content for bid"),
		snhash.WithProfile(snhash.ProfileEvidence),
		snhash.WithCreatedAt(time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)),
	)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := AttachSNID(manifest); err != nil {
			b.Fatal(err)
		}
	}
}

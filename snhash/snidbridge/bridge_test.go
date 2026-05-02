package snidbridge

import (
	"strings"
	"testing"
	"time"

	"github.com/LastMile-Innovations/snhash"
)

func TestAttachSNID(t *testing.T) {
	manifest, err := snhash.HashBytes(
		[]byte("content for bid"),
		snhash.WithProfile(snhash.ProfileEvidence),
		snhash.WithCreatedAt(time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)),
	)
	if err != nil {
		t.Fatal(err)
	}
	bid, err := AttachSNID(manifest)
	if err != nil {
		t.Fatal(err)
	}
	blake3Digest, err := manifest.BLAKE3()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := bid.R2Key(), blake3Digest.Base32(); got != want {
		t.Fatalf("r2 key = %s, want %s", got, want)
	}
	if wire := bid.WireFormat(); !strings.HasPrefix(wire, "CAS:") {
		t.Fatalf("wire = %q", wire)
	}
}

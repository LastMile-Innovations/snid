package snhash

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkHashBytesFastCAS(b *testing.B) {
	data := bytes.Repeat([]byte("a"), 4096)
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		if _, err := HashBytes(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHashBytesCache(b *testing.B) {
	data := bytes.Repeat([]byte("cache"), 8192)
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		if _, err := HashBytes(data, WithProfile(ProfileCache)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHashReaderEvidence1MiB(b *testing.B) {
	data := bytes.Repeat([]byte("evidence"), 128*1024)
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		if _, err := HashReader(bytes.NewReader(data), WithProfile(ProfileEvidence)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHashFileEvidence1MiB(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "evidence.bin")
	data := bytes.Repeat([]byte("file"), 256*1024)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		if _, err := HashFile(path, WithProfile(ProfileEvidence)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkChunkedHash1MiB(b *testing.B) {
	data := bytes.Repeat([]byte("chunked"), 150*1024)
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		if _, err := HashBytes(data, WithChunkSize(64*1024)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProofRoot(b *testing.B) {
	manifest, err := HashBytes([]byte("manifest proof root"), WithProfile(ProfileEvidence), WithCreatedAt(fixedTime))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := ProofRoot(manifest); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHashIDFromManifest(b *testing.B) {
	manifest, err := HashBytes([]byte("manifest proof root"), WithProfile(ProfileEvidence), WithCreatedAt(fixedTime))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := manifest.HashID(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCanonicalManifestJSON(b *testing.B) {
	manifest, err := HashBytes([]byte("manifest proof root"), WithProfile(ProfileEvidence), WithCreatedAt(fixedTime))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := manifest.CanonicalJSON(); err != nil {
			b.Fatal(err)
		}
	}
}

func FuzzParseDigestWire(f *testing.F) {
	manifest, err := HashBytes([]byte("seed"), WithProfile(ProfileEvidence), WithCreatedAt(fixedTime))
	if err != nil {
		f.Fatal(err)
	}
	for _, digest := range manifest.Digests {
		f.Add(digest.Wire())
	}
	f.Add("h1:sha-256:not-base32")
	f.Add("bad")

	f.Fuzz(func(t *testing.T, wire string) {
		digest, err := ParseDigestWire(wire)
		if err != nil {
			return
		}
		roundTrip, err := ParseDigestWire(digest.Wire())
		if err != nil {
			t.Fatal(err)
		}
		if roundTrip.Algorithm != digest.Algorithm || roundTrip.Hex() != digest.Hex() {
			t.Fatalf("round trip mismatch")
		}
	})
}

func FuzzManifestJSON(f *testing.F) {
	manifest, err := HashBytes([]byte("seed"), WithProfile(ProfileEvidence), WithCreatedAt(fixedTime))
	if err != nil {
		f.Fatal(err)
	}
	data, err := manifest.CanonicalJSON()
	if err != nil {
		f.Fatal(err)
	}
	f.Add(data)
	f.Add([]byte(`{"version":"snhash-manifest-v1","digests":[]}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var manifest Manifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			return
		}
		if _, err := manifest.CanonicalJSON(); err != nil {
			t.Fatal(err)
		}
	})
}

func FuzzDigestVerify(f *testing.F) {
	manifest, err := HashBytes([]byte("seed"), WithProfile(ProfileEvidence), WithCreatedAt(fixedTime))
	if err != nil {
		f.Fatal(err)
	}
	blake3Digest, err := manifest.BLAKE3()
	if err != nil {
		f.Fatal(err)
	}
	f.Add(blake3Digest.Wire(), []byte("seed"))
	f.Add(blake3Digest.Wire(), []byte("changed"))

	f.Fuzz(func(t *testing.T, wire string, data []byte) {
		digest, err := ParseDigestWire(wire)
		if err != nil {
			return
		}
		_ = digest.Verify(data)
	})
}

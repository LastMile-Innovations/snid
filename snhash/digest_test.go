package snhash

import (
	"bytes"
	"encoding/hex"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var fixedTime = time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)

func TestHashBytesEvidenceVectors(t *testing.T) {
	manifest, err := HashBytes([]byte("abc"), WithProfile(ProfileEvidence), WithCreatedAt(fixedTime))
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Size != 3 {
		t.Fatalf("size = %d, want 3", manifest.Size)
	}

	blake3Digest := manifest.Digest(AlgorithmBLAKE3_256)
	if blake3Digest == nil {
		t.Fatal("missing blake3 digest")
	}
	if got, want := blake3Digest.Hex(), "6437b3ac38465133ffb63b75273a8db548c558465d79db03fd359c6cd5bd9d85"; got != want {
		t.Fatalf("blake3 = %s, want %s", got, want)
	}

	shaDigest := manifest.Digest(AlgorithmSHA256)
	if shaDigest == nil {
		t.Fatal("missing sha-256 digest")
	}
	if got, want := shaDigest.Hex(), "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"; got != want {
		t.Fatalf("sha-256 = %s, want %s", got, want)
	}
	if !manifest.VerifyBytes([]byte("abc")) {
		t.Fatal("manifest should verify original bytes")
	}
	if manifest.VerifyBytes([]byte("abd")) {
		t.Fatal("manifest should reject changed bytes")
	}
}

func TestDigestWireRoundTrip(t *testing.T) {
	raw, _ := hex.DecodeString("ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad")
	digest, err := NewDigest(AlgorithmSHA256, raw)
	if err != nil {
		t.Fatal(err)
	}
	wire := digest.Wire()
	if !strings.HasPrefix(wire, "h1:sha-256:") {
		t.Fatalf("wire = %q", wire)
	}
	parsed, err := ParseDigestWire(wire)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Algorithm != digest.Algorithm || parsed.Hex() != digest.Hex() {
		t.Fatalf("parsed = %#v, want %#v", parsed, digest)
	}
}

func TestLegacyAlgorithmsAreImportOnly(t *testing.T) {
	manifest, err := HashBytes([]byte("legacy"), WithAlgorithms(AlgorithmMD5, AlgorithmSHA1), WithCreatedAt(fixedTime))
	if err != nil {
		t.Fatal(err)
	}
	for _, digest := range manifest.Digests {
		if digest.Trusted() {
			t.Fatalf("%s should not be trusted", digest.Algorithm)
		}
		if !digest.Verify([]byte("legacy")) {
			t.Fatalf("%s should still be parseable/verifiable as legacy metadata", digest.Algorithm)
		}
	}
}

func TestCacheProfileUsesXXH3AsUntrustedAccelerator(t *testing.T) {
	manifest, err := HashBytes([]byte("cache payload"), WithProfile(ProfileCache), WithCreatedAt(fixedTime))
	if err != nil {
		t.Fatal(err)
	}
	xxh3Digest := manifest.Digest(AlgorithmXXH3_64)
	if xxh3Digest == nil {
		t.Fatal("missing xxh3 digest")
	}
	if xxh3Digest.Size != 8 {
		t.Fatalf("xxh3 size = %d, want 8", xxh3Digest.Size)
	}
	if xxh3Digest.Trusted() {
		t.Fatal("xxh3 must remain untrusted")
	}
	if manifest.Digest(AlgorithmBLAKE3_256) == nil {
		t.Fatal("cache profile should include blake3 confirmation")
	}
	if !manifest.VerifyBytes([]byte("cache payload")) {
		t.Fatal("cache profile should verify through trusted blake3")
	}
}

func TestHashIDRoundTrip(t *testing.T) {
	manifest, err := HashBytes([]byte("hash id payload"), WithProfile(ProfileEvidence), WithCreatedAt(fixedTime))
	if err != nil {
		t.Fatal(err)
	}
	id, err := manifest.HashID()
	if err != nil {
		t.Fatal(err)
	}
	if !id.Verify([]byte("hash id payload")) {
		t.Fatal("hash id should verify original payload")
	}
	if id.Verify([]byte("changed")) {
		t.Fatal("hash id should reject changed payload")
	}
	parsed, err := ParseHashID(id.Wire())
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Wire() != id.Wire() {
		t.Fatalf("hash id round trip = %q, want %q", parsed.Wire(), id.Wire())
	}
}

func TestProofRootDeterministic(t *testing.T) {
	manifest, err := HashBytes([]byte("case evidence"), WithProfile(ProfileEvidence), WithCreatedAt(fixedTime))
	if err != nil {
		t.Fatal(err)
	}
	first, err := ProofRoot(manifest)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ProofRoot(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if first.Hex() != second.Hex() {
		t.Fatalf("proof root not deterministic: %s != %s", first.Hex(), second.Hex())
	}
	canonical, err := manifest.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !first.Verify(canonical) {
		t.Fatal("proof root should verify canonical manifest bytes")
	}
	ok, err := first.VerifyReader(bytes.NewReader(canonical))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("proof root should verify canonical manifest stream")
	}
}

func TestCustodyEventsInManifest(t *testing.T) {
	event := CustodyEvent{
		At:     fixedTime,
		Actor:  "clerk",
		Action: "received",
		Note:   "original intake",
	}
	manifest, err := HashBytes([]byte("case evidence"), WithProfile(ProfileEvidence), WithCreatedAt(fixedTime), WithCustody(event))
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Custody) != 1 {
		t.Fatalf("custody events = %d, want 1", len(manifest.Custody))
	}
	canonical, err := manifest.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(canonical), `"action":"received"`) {
		t.Fatalf("canonical manifest missing custody event: %s", canonical)
	}
}

func TestManifestFileRoundTrip(t *testing.T) {
	manifest, err := HashBytes([]byte("case evidence"), WithProfile(ProfileEvidence), WithCreatedAt(fixedTime))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := manifest.WriteFile(path, 0o600); err != nil {
		t.Fatal(err)
	}
	parsed, err := ReadManifestFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.String() != manifest.String() {
		t.Fatalf("manifest round trip mismatch\n%s\n%s", parsed.String(), manifest.String())
	}
}

func TestChunkManifest(t *testing.T) {
	manifest, err := HashBytes([]byte("0123456789"), WithChunkSize(4), WithCreatedAt(fixedTime))
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Chunks) != 3 {
		t.Fatalf("chunks = %d, want 3", len(manifest.Chunks))
	}
	if manifest.TreeRoot == nil {
		t.Fatal("chunked manifest should include tree root")
	}
	second, err := HashBytes([]byte("0123456789"), WithChunkSize(4), WithCreatedAt(fixedTime))
	if err != nil {
		t.Fatal(err)
	}
	if second.TreeRoot == nil || second.TreeRoot.Hex() != manifest.TreeRoot.Hex() {
		t.Fatal("tree root should be deterministic")
	}
	if !manifest.VerifyChunk(0, []byte("0123")) {
		t.Fatal("chunk 0 should verify")
	}
	if manifest.VerifyChunk(0, []byte("0124")) {
		t.Fatal("changed chunk should not verify")
	}
	expected := []struct {
		offset int64
		size   int64
	}{
		{0, 4},
		{4, 4},
		{8, 2},
	}
	for i, want := range expected {
		if manifest.Chunks[i].Offset != want.offset || manifest.Chunks[i].Size != want.size {
			t.Fatalf("chunk %d = offset %d size %d", i, manifest.Chunks[i].Offset, manifest.Chunks[i].Size)
		}
	}
}

func TestManifestVerifyReader(t *testing.T) {
	manifest, err := HashBytes([]byte("stream body"), WithProfile(ProfileEvidence), WithCreatedAt(fixedTime))
	if err != nil {
		t.Fatal(err)
	}
	ok, err := manifest.VerifyReader(bytes.NewReader([]byte("stream body")))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("reader should verify")
	}
	ok, err = manifest.VerifyReader(bytes.NewReader([]byte("stream body changed")))
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("changed reader should not verify")
	}
}

func TestAPISignatures(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	data := []byte("request body")

	hmacDigest, err := HMACSHA256(key, data)
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyHMACSHA256(key, data, hmacDigest) {
		t.Fatal("hmac should verify")
	}
	if VerifyHMACSHA256(key, []byte("changed"), hmacDigest) {
		t.Fatal("hmac should reject changed data")
	}

	keyed, err := KeyedBLAKE3(key, data)
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyKeyedBLAKE3(key, data, keyed) {
		t.Fatal("keyed blake3 should verify")
	}
	if VerifyKeyedBLAKE3(key, []byte("changed"), keyed) {
		t.Fatal("keyed blake3 should reject changed data")
	}
}

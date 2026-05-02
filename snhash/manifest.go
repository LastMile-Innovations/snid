package snhash

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"sort"
	"time"

	"github.com/zeebo/blake3"
)

const ManifestVersion = "snhash-manifest-v1"

var ErrDigestNotFound = errors.New("snhash: digest not found")

// Manifest is the canonical proof envelope for content integrity.
type Manifest struct {
	Version          string         `json:"version"`
	Profile          Profile        `json:"profile"`
	Size             int64          `json:"size"`
	MediaType        string         `json:"media_type"`
	Canonicalization string         `json:"canonicalization"`
	SourceName       string         `json:"source_name,omitempty"`
	Digests          []Digest       `json:"digests"`
	TreeRoot         *Digest        `json:"tree_root,omitempty"`
	Chunks           []Chunk        `json:"chunks,omitempty"`
	Entries          []Entry        `json:"entries,omitempty"`
	Custody          []CustodyEvent `json:"custody,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
}

// Chunk records per-chunk digests for streaming or Merkle-style verification.
type Chunk struct {
	Index   int      `json:"index"`
	Offset  int64    `json:"offset"`
	Size    int64    `json:"size"`
	Digests []Digest `json:"digests"`
}

// Entry records a deterministic directory manifest entry.
type Entry struct {
	Path    string   `json:"path"`
	Type    string   `json:"type"`
	Mode    string   `json:"mode,omitempty"`
	Size    int64    `json:"size"`
	Digests []Digest `json:"digests,omitempty"`
}

// CustodyEvent records an optional chain-of-custody event.
type CustodyEvent struct {
	At     time.Time `json:"at"`
	Actor  string    `json:"actor,omitempty"`
	Action string    `json:"action"`
	Note   string    `json:"note,omitempty"`
}

// Digest returns the manifest digest for an algorithm, if present.
func (m Manifest) Digest(algorithm Algorithm) *Digest {
	for i := range m.Digests {
		if m.Digests[i].Algorithm == algorithm {
			return &m.Digests[i]
		}
	}
	return nil
}

// BLAKE3 returns the canonical BLAKE3-256 digest.
func (m Manifest) BLAKE3() (Digest, error) {
	digest := m.Digest(AlgorithmBLAKE3_256)
	if digest == nil {
		return Digest{}, ErrDigestNotFound
	}
	return *digest, nil
}

// SHA256 returns the canonical SHA-256 digest.
func (m Manifest) SHA256() (Digest, error) {
	digest := m.Digest(AlgorithmSHA256)
	if digest == nil {
		return Digest{}, ErrDigestNotFound
	}
	return *digest, nil
}

// HashID returns the strongest compact content identifier available.
func (m Manifest) HashID() (HashID, error) {
	for _, algorithm := range []Algorithm{
		AlgorithmBLAKE3_256,
		AlgorithmSHA256,
		AlgorithmSHA3_256,
		AlgorithmSHAKE256_256,
	} {
		if digest := m.Digest(algorithm); digest != nil {
			return NewHashID(m.Profile, m.Size, *digest)
		}
	}
	if m.TreeRoot != nil && m.TreeRoot.Trusted() {
		return NewHashID(m.Profile, m.Size, *m.TreeRoot)
	}
	return HashID{}, ErrDigestNotFound
}

// CanonicalJSON returns deterministic manifest JSON.
func (m Manifest) CanonicalJSON() ([]byte, error) {
	normalized := m.normalized()
	return json.Marshal(normalized)
}

// String returns canonical JSON, or an empty string if the manifest is invalid.
func (m Manifest) String() string {
	out, err := m.CanonicalJSON()
	if err != nil {
		return ""
	}
	return string(out)
}

// WriteFile writes canonical manifest JSON to path.
func (m Manifest) WriteFile(path string, perm os.FileMode) error {
	out, err := m.CanonicalJSON()
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, perm)
}

// ParseManifestJSON parses deterministic manifest JSON.
func ParseManifestJSON(data []byte) (Manifest, error) {
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, err
	}
	return manifest.normalized(), nil
}

// ReadManifestFile reads a manifest from a JSON file.
func ReadManifestFile(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	return ParseManifestJSON(data)
}

// ProofRoot returns BLAKE3-256 over canonical manifest JSON.
func ProofRoot(m Manifest) (Digest, error) {
	canonical, err := m.CanonicalJSON()
	if err != nil {
		return Digest{}, err
	}
	sum := blake3.Sum256(canonical)
	return NewDigest(AlgorithmBLAKE3_256, sum[:])
}

// VerifyBytes checks every trusted digest on the manifest against data.
func (m Manifest) VerifyBytes(data []byte) bool {
	for _, expected := range m.Digests {
		if !expected.Trusted() {
			continue
		}
		if !expected.Verify(data) {
			return false
		}
	}
	return len(m.Digests) > 0
}

// VerifyReader checks every trusted manifest digest against a stream.
func (m Manifest) VerifyReader(r io.Reader) (bool, error) {
	algorithms := trustedAlgorithms(m.Digests)
	if len(algorithms) == 0 {
		return false, nil
	}
	actual, err := HashReader(r, WithAlgorithms(algorithms...))
	if err != nil {
		return false, err
	}
	if actual.Size != m.Size {
		return false, nil
	}
	for _, expected := range m.Digests {
		if !expected.Trusted() {
			continue
		}
		got := actual.Digest(expected.Algorithm)
		if got == nil || !equalBytes(got.Bytes, expected.Bytes) {
			return false, nil
		}
	}
	return true, nil
}

// VerifyChunk checks one chunk against its trusted per-chunk digest.
func (m Manifest) VerifyChunk(index int, data []byte) bool {
	if index < 0 || index >= len(m.Chunks) {
		return false
	}
	chunk := m.Chunks[index]
	if int64(len(data)) != chunk.Size {
		return false
	}
	algorithms := trustedAlgorithms(chunk.Digests)
	if len(algorithms) == 0 {
		return false
	}
	actual, err := HashBytes(data, WithAlgorithms(algorithms...))
	if err != nil {
		return false
	}
	for _, expected := range chunk.Digests {
		if !expected.Trusted() {
			continue
		}
		got := actual.Digest(expected.Algorithm)
		if got == nil || !equalBytes(got.Bytes, expected.Bytes) {
			return false
		}
	}
	return true
}

func (m Manifest) normalized() Manifest {
	out := m
	if out.Version == "" {
		out.Version = ManifestVersion
	}
	out.Profile = out.Profile.normalized()
	out.CreatedAt = out.CreatedAt.UTC()
	out.Digests = normalizeDigests(out.Digests)
	if out.TreeRoot != nil {
		treeRoot, err := NewDigest(out.TreeRoot.Algorithm, out.TreeRoot.Bytes)
		if err == nil {
			out.TreeRoot = &treeRoot
		}
	}
	out.Chunks = append([]Chunk(nil), out.Chunks...)
	for i := range out.Chunks {
		out.Chunks[i].Digests = normalizeDigests(out.Chunks[i].Digests)
	}
	sort.Slice(out.Chunks, func(i, j int) bool {
		if out.Chunks[i].Index == out.Chunks[j].Index {
			return out.Chunks[i].Offset < out.Chunks[j].Offset
		}
		return out.Chunks[i].Index < out.Chunks[j].Index
	})
	out.Entries = append([]Entry(nil), out.Entries...)
	for i := range out.Entries {
		out.Entries[i].Digests = normalizeDigests(out.Entries[i].Digests)
	}
	sort.Slice(out.Entries, func(i, j int) bool {
		return out.Entries[i].Path < out.Entries[j].Path
	})
	out.Custody = append([]CustodyEvent(nil), out.Custody...)
	for i := range out.Custody {
		out.Custody[i].At = out.Custody[i].At.UTC()
	}
	return out
}

func normalizeDigests(digests []Digest) []Digest {
	out := append([]Digest(nil), digests...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Algorithm < out[j].Algorithm
	})
	return out
}

func trustedAlgorithms(digests []Digest) []Algorithm {
	out := make([]Algorithm, 0, len(digests))
	for _, digest := range digests {
		if digest.Trusted() {
			out = append(out, digest.Algorithm)
		}
	}
	return normalizeAlgorithms(out)
}

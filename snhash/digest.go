package snhash

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha3"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash"
	"io"
	"strings"

	"github.com/zeebo/blake3"
	"github.com/zeebo/xxh3"
)

// Algorithm names are lower-case wire values. Legacy algorithms are parseable
// for import metadata, but IsTrusted returns false for verification policy.
type Algorithm string

const (
	AlgorithmBLAKE3_256      Algorithm = "blake3-256"
	AlgorithmBLAKE3Keyed_256 Algorithm = "blake3-keyed-256"
	AlgorithmSHA256          Algorithm = "sha-256"
	AlgorithmHMACSHA256      Algorithm = "hmac-sha-256"
	AlgorithmSHA3_256        Algorithm = "sha3-256"
	AlgorithmSHAKE256_256    Algorithm = "shake256-256"
	AlgorithmXXH3_64         Algorithm = "xxh3-64"
	AlgorithmMD5             Algorithm = "md5"
	AlgorithmSHA1            Algorithm = "sha-1"
	wireVersion                        = "h1"
	hashIDWireVersion                  = "hid1"
	defaultDigestSize                  = 32
)

var (
	ErrUnknownAlgorithm = errors.New("snhash: unknown algorithm")
	ErrInvalidDigest    = errors.New("snhash: invalid digest")
	ErrInvalidWire      = errors.New("snhash: invalid digest wire")
)

var digestBase32 = base32.StdEncoding.WithPadding(base32.NoPadding)

// Digest is a self-describing hash output. Size is the number of digest bytes.
type Digest struct {
	Algorithm Algorithm `json:"algorithm"`
	Size      int       `json:"size"`
	Bytes     []byte    `json:"-"`
}

// NewDigest validates and copies raw digest bytes.
func NewDigest(algorithm Algorithm, raw []byte) (Digest, error) {
	if err := algorithm.ValidateDigestSize(len(raw)); err != nil {
		return Digest{}, err
	}
	out := make([]byte, len(raw))
	copy(out, raw)
	return Digest{Algorithm: algorithm, Size: len(out), Bytes: out}, nil
}

// Hex returns lower-case hexadecimal digest bytes.
func (d Digest) Hex() string {
	return hex.EncodeToString(d.Bytes)
}

// Base32 returns lower-case RFC 4648 base32 without padding.
func (d Digest) Base32() string {
	return strings.ToLower(digestBase32.EncodeToString(d.Bytes))
}

// Wire returns h1:<algorithm>:<base32-lower-no-padding>.
func (d Digest) Wire() string {
	return wireVersion + ":" + string(d.Algorithm) + ":" + d.Base32()
}

// Trusted reports whether this digest can be used for current verification.
func (d Digest) Trusted() bool {
	return d.Algorithm.IsTrusted()
}

// Verify hashes data with the digest algorithm and compares bytes.
func (d Digest) Verify(data []byte) bool {
	actual, err := HashBytes(data, WithAlgorithms(d.Algorithm))
	if err != nil {
		return false
	}
	got := actual.Digest(d.Algorithm)
	return got != nil && equalBytes(got.Bytes, d.Bytes)
}

// VerifyReader hashes a stream with the digest algorithm and compares bytes.
func (d Digest) VerifyReader(r io.Reader) (bool, error) {
	actual, err := HashReader(r, WithAlgorithms(d.Algorithm))
	if err != nil {
		return false, err
	}
	got := actual.Digest(d.Algorithm)
	return got != nil && equalBytes(got.Bytes, d.Bytes), nil
}

// MarshalJSON emits a deterministic, human-readable digest object.
func (d Digest) MarshalJSON() ([]byte, error) {
	type digestJSON struct {
		Algorithm Algorithm `json:"algorithm"`
		Size      int       `json:"size"`
		Hex       string    `json:"hex"`
		Base32    string    `json:"base32"`
		Wire      string    `json:"wire"`
		Trusted   bool      `json:"trusted"`
	}
	return json.Marshal(digestJSON{
		Algorithm: d.Algorithm,
		Size:      d.Size,
		Hex:       d.Hex(),
		Base32:    d.Base32(),
		Wire:      d.Wire(),
		Trusted:   d.Trusted(),
	})
}

// UnmarshalJSON accepts objects emitted by MarshalJSON.
func (d *Digest) UnmarshalJSON(data []byte) error {
	var in struct {
		Algorithm Algorithm `json:"algorithm"`
		Size      int       `json:"size"`
		Hex       string    `json:"hex"`
		Wire      string    `json:"wire"`
	}
	if err := json.Unmarshal(data, &in); err != nil {
		return err
	}
	if in.Wire != "" && in.Hex == "" {
		parsed, err := ParseDigestWire(in.Wire)
		if err != nil {
			return err
		}
		*d = parsed
		return nil
	}
	raw, err := hex.DecodeString(in.Hex)
	if err != nil {
		return err
	}
	parsed, err := NewDigest(in.Algorithm, raw)
	if err != nil {
		return err
	}
	if in.Size != 0 && in.Size != parsed.Size {
		return ErrInvalidDigest
	}
	*d = parsed
	return nil
}

// ParseDigestWire parses h1:<algorithm>:<base32-lower-no-padding>.
func ParseDigestWire(wire string) (Digest, error) {
	parts := strings.Split(wire, ":")
	if len(parts) != 3 || parts[0] != wireVersion {
		return Digest{}, ErrInvalidWire
	}
	algorithm := Algorithm(parts[1])
	raw, err := digestBase32.DecodeString(strings.ToUpper(parts[2]))
	if err != nil {
		return Digest{}, ErrInvalidWire
	}
	return NewDigest(algorithm, raw)
}

// IsTrusted reports whether an algorithm is acceptable for current integrity
// checks. XXH3, MD5, and SHA-1 are import/cache metadata only.
func (a Algorithm) IsTrusted() bool {
	switch a {
	case AlgorithmBLAKE3_256, AlgorithmBLAKE3Keyed_256, AlgorithmSHA256,
		AlgorithmHMACSHA256, AlgorithmSHA3_256, AlgorithmSHAKE256_256:
		return true
	default:
		return false
	}
}

// ValidateDigestSize verifies the canonical digest length for an algorithm.
func (a Algorithm) ValidateDigestSize(size int) error {
	switch a {
	case AlgorithmBLAKE3_256, AlgorithmBLAKE3Keyed_256, AlgorithmSHA256,
		AlgorithmHMACSHA256, AlgorithmSHA3_256, AlgorithmSHAKE256_256:
		if size == defaultDigestSize {
			return nil
		}
	case AlgorithmMD5:
		if size == md5.Size {
			return nil
		}
	case AlgorithmSHA1:
		if size == sha1.Size {
			return nil
		}
	case AlgorithmXXH3_64:
		if size == 8 {
			return nil
		}
	default:
		return ErrUnknownAlgorithm
	}
	return ErrInvalidDigest
}

type digestState interface {
	Write([]byte) (int, error)
	SumBytes() []byte
}

type hashState struct {
	h hash.Hash
}

func (s hashState) Write(p []byte) (int, error) { return s.h.Write(p) }
func (s hashState) SumBytes() []byte            { return s.h.Sum(nil) }

type shakeState struct {
	h    *sha3.SHAKE
	size int
}

func (s shakeState) Write(p []byte) (int, error) { return s.h.Write(p) }
func (s shakeState) SumBytes() []byte {
	out := make([]byte, s.size)
	_, _ = s.h.Read(out)
	return out
}

func newDigestState(algorithm Algorithm) (digestState, error) {
	switch algorithm {
	case AlgorithmBLAKE3_256:
		return hashState{h: blake3.New()}, nil
	case AlgorithmSHA256:
		return hashState{h: sha256.New()}, nil
	case AlgorithmSHA3_256:
		return hashState{h: sha3.New256()}, nil
	case AlgorithmSHAKE256_256:
		return shakeState{h: sha3.NewSHAKE256(), size: defaultDigestSize}, nil
	case AlgorithmXXH3_64:
		return hashState{h: xxh3.New()}, nil
	case AlgorithmMD5:
		return hashState{h: md5.New()}, nil
	case AlgorithmSHA1:
		return hashState{h: sha1.New()}, nil
	default:
		return nil, ErrUnknownAlgorithm
	}
}

func digestFromState(algorithm Algorithm, state digestState) (Digest, error) {
	return NewDigest(algorithm, state.SumBytes())
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := range a {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

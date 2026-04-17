package snid

import (
	"encoding/base32"
	"errors"
	"strings"

	"github.com/zeebo/blake3"
)

var (
	// ErrInvalidBIDWire is returned when BID wire format is malformed.
	ErrInvalidBIDWire = errors.New("snid: invalid BID wire format")
	// ErrInvalidContentHash is returned when hash length is not 32 bytes.
	ErrInvalidContentHash = errors.New("snid: invalid content hash length")
)

var bidBase32 = base32.StdEncoding.WithPadding(base32.NoPadding)

// BID binds topological identity (SNID) and content identity (BLAKE3-256).
// Topology is optimized for graph/index locality; Content is optimized for dedup/integrity.
type BID struct {
	Topology ID
	Content  [32]byte
}

// NewBID creates a BID from a precomputed BLAKE3-256 hash.
func NewBID(contentHash [32]byte) BID {
	return BID{
		Topology: NewFast(),
		Content:  contentHash,
	}
}

// NewBIDFromHash creates a BID from a raw 32-byte hash digest.
func NewBIDFromHash(contentHash []byte) (BID, error) {
	if len(contentHash) != 32 {
		return BID{}, ErrInvalidContentHash
	}
	var sum [32]byte
	copy(sum[:], contentHash)
	return NewBID(sum), nil
}

// NewBIDFromContent hashes plaintext bytes with BLAKE3-256 and returns a BID.
func NewBIDFromContent(content []byte) BID {
	sum := blake3.Sum256(content)
	return NewBID(sum)
}

// WireFormat serializes BID as: CAS:<snid_payload_base58>:<cid_base32_lower>
func (b BID) WireFormat() string {
	return "CAS:" + b.Topology.StringCompact() + ":" + b.R2Key()
}

// ParseBIDWire parses BID wire format: CAS:<snid_payload_base58>:<cid_base32_lower>
func ParseBIDWire(s string) (BID, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 3 || parts[0] != "CAS" {
		return BID{}, ErrInvalidBIDWire
	}

	var id ID
	if err := id.ParseCompact(parts[1]); err != nil {
		return BID{}, err
	}

	raw, err := bidBase32.DecodeString(strings.ToUpper(parts[2]))
	if err != nil {
		return BID{}, err
	}
	if len(raw) != 32 {
		return BID{}, ErrInvalidContentHash
	}

	var content [32]byte
	copy(content[:], raw)

	return BID{
		Topology: id,
		Content:  content,
	}, nil
}

// Neo4jID returns a canonical graph ID for blob topology nodes.
func (b BID) Neo4jID() string {
	return b.Topology.String(Matter)
}

// Neo4jIDWithAtom allows explicit graph atom selection while enforcing canonical atoms.
func (b BID) Neo4jIDWithAtom(atom Atom) string {
	if !IsValidAtom(atom) {
		atom = Matter
	}
	return b.Topology.String(CanonicalAtom(atom))
}

// R2Key returns a deterministic content-key-safe base32 string (lowercase, no padding).
func (b BID) R2Key() string {
	return strings.ToLower(bidBase32.EncodeToString(b.Content[:]))
}

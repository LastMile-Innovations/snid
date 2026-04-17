package snid

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"

	"github.com/zeebo/blake3"
)

// WID is a 256-bit world or scenario identifier.
type WID [32]byte

// XID is a 256-bit edge or relationship identifier.
type XID [32]byte

// KID is a 256-bit self-verifying capability identifier.
type KID [32]byte

// NewWID creates a world identifier from a causal head and scenario hash.
func NewWID(head ID, scenarioHash [16]byte) WID {
	var wid WID
	copy(wid[:16], head[:])
	copy(wid[16:], scenarioHash[:])
	return wid
}

// NewWIDFromScope derives a deterministic world identifier from timestamped scope data.
func NewWIDFromScope(unixMillis uint64, scopeHash []byte) WID {
	sum := sha256.Sum256(scopeHash)
	var tail [16]byte
	copy(tail[:], sum[:16])
	return NewWID(NewDeterministicIngestID(unixMillis, scopeHash), tail)
}

// Head returns the join-friendly SNID head.
func (w WID) Head() ID {
	var id ID
	copy(id[:], w[:16])
	return id
}

// ScenarioHash returns the scenario/world hash tail.
func (w WID) ScenarioHash() [16]byte {
	var out [16]byte
	copy(out[:], w[16:])
	return out
}

// ToTensor256Words exports the full 256-bit value as four big-endian int64 words.
func (w WID) ToTensor256Words() (int64, int64, int64, int64) {
	return int64FromBytes(w[0:8]), int64FromBytes(w[8:16]), int64FromBytes(w[16:24]), int64FromBytes(w[24:32])
}

// NewXID creates a relationship identifier bound to the supplied edge hash.
func NewXID(head ID, edgeHash [16]byte) XID {
	var xid XID
	copy(xid[:16], head[:])
	copy(xid[16:], edgeHash[:])
	return xid
}

// NewXIDFromParts deterministically derives an edge identifier from two endpoints and edge kind.
func NewXIDFromParts(head ID, left ID, right ID, edgeKind []byte) XID {
	material := make([]byte, 0, 16+16+len(edgeKind))
	material = append(material, left[:]...)
	material = append(material, right[:]...)
	material = append(material, edgeKind...)
	sum := blake3.Sum256(material)
	var tail [16]byte
	copy(tail[:], sum[:16])
	return NewXID(head, tail)
}

// Head returns the join-friendly SNID head.
func (x XID) Head() ID {
	var id ID
	copy(id[:], x[:16])
	return id
}

// EdgeHash returns the edge-specific hash tail.
func (x XID) EdgeHash() [16]byte {
	var out [16]byte
	copy(out[:], x[16:])
	return out
}

// ToTensor256Words exports the full 256-bit value as four big-endian int64 words.
func (x XID) ToTensor256Words() (int64, int64, int64, int64) {
	return int64FromBytes(x[0:8]), int64FromBytes(x[8:16]), int64FromBytes(x[16:24]), int64FromBytes(x[24:32])
}

// NewKID creates a self-verifying capability identifier from an explicit head and MAC tail.
func NewKID(head ID, macTail [16]byte) KID {
	var kid KID
	copy(kid[:16], head[:])
	copy(kid[16:], macTail[:])
	return kid
}

// NewKIDForCapability derives a self-verifying capability token from actor, resource, and capability bytes.
func NewKIDForCapability(head ID, actor ID, resource []byte, capability []byte, key []byte) (KID, error) {
	if len(key) == 0 {
		return KID{}, ErrInvalidLIDKey
	}
	mac := hmac.New(sha256.New, key)
	mac.Write(head[:])
	mac.Write(actor[:])
	mac.Write(resource)
	mac.Write(capability)
	sum := mac.Sum(nil)
	var tail [16]byte
	copy(tail[:], sum[:16])
	return NewKID(head, tail), nil
}

// Verify validates the KID MAC using the supplied capability inputs.
func (k KID) Verify(actor ID, resource []byte, capability []byte, key []byte) bool {
	if len(key) == 0 {
		return false
	}
	expected, err := NewKIDForCapability(k.Head(), actor, resource, capability, key)
	if err != nil {
		return false
	}
	return hmac.Equal(k[16:], expected[16:])
}

// Head returns the join-friendly SNID head.
func (k KID) Head() ID {
	var id ID
	copy(id[:], k[:16])
	return id
}

// ToTensor256Words exports the full 256-bit value as four big-endian int64 words.
func (k KID) ToTensor256Words() (int64, int64, int64, int64) {
	return int64FromBytes(k[0:8]), int64FromBytes(k[8:16]), int64FromBytes(k[16:24]), int64FromBytes(k[24:32])
}

// LIDBLAKE3 returns a 256-bit ledger identifier with a BLAKE3 keyed tail for target-state migration work.
func LIDBLAKE3(head ID, prev LID, payload []byte, key []byte) (LID, error) {
	if len(key) == 0 {
		return LID{}, ErrInvalidLIDKey
	}
	material := make([]byte, 0, len(key)+16+32+len(payload))
	material = append(material, key...)
	material = append(material, head[:]...)
	material = append(material, prev[:]...)
	material = append(material, payload...)
	sum := blake3.Sum256(material)
	var lid LID
	copy(lid[:16], head[:])
	copy(lid[16:], sum[:16])
	return lid, nil
}

// StorageBytes returns the canonical binary storage form.
func (w WID) StorageBytes() []byte { return append([]byte(nil), w[:]...) }

// StorageBytes returns the canonical binary storage form.
func (x XID) StorageBytes() []byte { return append([]byte(nil), x[:]...) }

// StorageBytes returns the canonical binary storage form.
func (k KID) StorageBytes() []byte { return append([]byte(nil), k[:]...) }

func int64FromBytes(raw []byte) int64 {
	return int64(binary.BigEndian.Uint64(raw))
}

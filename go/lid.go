package snid

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

var ErrInvalidLIDKey = errors.New("snid: invalid LID key")

// LID is a 256-bit ledger identifier: 128-bit SNID head + 128-bit causal hash.
type LID [32]byte

// NewLID creates a ledger ID bound to the previous LID and current payload.
func NewLID(prev LID, payload []byte, key []byte) (lid LID, err error) {
	if len(key) == 0 {
		return LID{}, ErrInvalidLIDKey
	}
	doSecret(func() {
		head := NewLedger()
		tail := computeLIDTail(head, prev, payload, key)

		copy(lid[:16], head[:])
		copy(lid[16:], tail[:])
	})
	return lid, nil
}

func computeLIDTail(head ID, prev LID, payload []byte, key []byte) [16]byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(head[:])
	mac.Write(prev[:])
	mac.Write(payload)
	sum := mac.Sum(nil)
	var tail [16]byte
	copy(tail[:], sum[:16])
	return tail
}

func computeLIDTailLegacy(prev LID, payload []byte, key []byte) [16]byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(prev[:])
	mac.Write(payload)
	sum := mac.Sum(nil)
	var tail [16]byte
	copy(tail[:], sum[:16])
	return tail
}

// Head returns the 128-bit SNID head used for joins/indexes.
func (l LID) Head() ID {
	var id ID
	copy(id[:], l[:16])
	return id
}

// ChainHash returns the 128-bit causal hash tail.
func (l LID) ChainHash() [16]byte {
	var out [16]byte
	copy(out[:], l[16:])
	return out
}

// Verify checks that the LID tail matches (prev, payload, key).
func (l LID) Verify(prev LID, payload []byte, key []byte) (ok bool) {
	if len(key) == 0 {
		return false
	}
	doSecret(func() {
		expected := computeLIDTail(l.Head(), prev, payload, key)
		if hmac.Equal(l[16:], expected[:]) {
			ok = true
			return
		}
		// Backward compatibility for tails minted before head-binding.
		legacy := computeLIDTailLegacy(prev, payload, key)
		ok = hmac.Equal(l[16:], legacy[:])
	})
	return
}

// String returns a hex representation for diagnostics.
func (l LID) String() string {
	return hex.EncodeToString(l[:])
}

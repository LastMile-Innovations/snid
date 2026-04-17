package snid

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"time"
)

// GrantKey holds the AES cipher for hardware-accelerated signing.
type GrantKey struct {
	block cipher.Block
}

// NewGrantKey initializes the AES cipher.
// This is expensive (~800ns) so should be done once and reused.
func NewGrantKey(secret []byte) (gk *GrantKey, err error) {
	if len(secret) != 16 && len(secret) != 24 && len(secret) != 32 {
		return nil, errors.New("snid: secret must be 16, 24, or 32 bytes for AES")
	}
	doSecret(func() {
		block, e := aes.NewCipher(secret)
		if e != nil {
			err = e
			return
		}
		gk = &GrantKey{block: block}
	})
	return
}

// NewGrantTurbo generates a signed ID using AES-CBC-MAC (Fixed Length).
// Cost: ~60ns (Hardware AES-NI)
func (gk *GrantKey) NewGrantTurbo(atom Atom, ttl time.Duration) (g GrantID) {
	doSecret(func() {
		atom = CanonicalAtom(atom)
		id := NewFast()
		var exp time.Time
		if ttl > 0 {
			exp = time.Now().Add(ttl)
		}

		g = GrantID{
			ID:        id,
			Atom:      atom,
			ExpiresAt: exp,
			Signature: signAESCBCWithAtom(gk.block, id, atom, exp),
		}
	})
	return
}

// VerifyTurbo verifies the signature using constant-time comparison.
// Cost: ~60ns (Hardware AES-NI)
func (gk *GrantKey) VerifyTurbo(g GrantID) (ok bool) {
	doSecret(func() {
		if !g.ExpiresAt.IsZero() && time.Now().After(g.ExpiresAt) {
			ok = false
			return
		}

		atom := CanonicalAtom(g.Atom)
		if atom != "" {
			expected := signAESCBCWithAtom(gk.block, g.ID, atom, g.ExpiresAt)
			if subtle.ConstantTimeCompare(g.Signature[:], expected[:]) == 1 {
				ok = true
				return
			}
		}
		// Backward compatibility for pre-v8.4 signatures.
		expected := signAESCBC(gk.block, g.ID, g.ExpiresAt)
		ok = subtle.ConstantTimeCompare(g.Signature[:], expected[:]) == 1
	})
	return
}

// signAESCBC implements a secure MAC for FIXED LENGTH (24 byte) messages.
// Algorithm: AES-CBC-MAC with zero IV.
// Block 1: ID (16 bytes)
// Block 2: ExpiresAt (8 bytes) || Padding (8 bytes)
//
// OPTIMIZATION: Vectorized 64-bit XOR instead of 8 byte-by-byte XORs.
func signAESCBC(block cipher.Block, id ID, exp time.Time) [16]byte {
	// Block 1: Encrypt ID
	var buf1 [16]byte
	block.Encrypt(buf1[:], id[:])

	// Block 2 Construction: XOR timestamp into Block 1 Output
	if !exp.IsZero() {
		t := uint64(exp.Unix())
		// VECTORIZED XOR: Load-XOR-Store as single 64-bit operations
		// This is faster than 8 byte-by-byte XORs
		v := binary.BigEndian.Uint64(buf1[:8])
		v ^= t
		binary.BigEndian.PutUint64(buf1[:8], v)
	}

	// Final Encryption
	var sig [16]byte
	block.Encrypt(sig[:], buf1[:])

	return sig
}

func signAESCBCWithAtom(block cipher.Block, id ID, atom Atom, exp time.Time) [16]byte {
	sig := signAESCBC(block, id, exp)
	// Mix atom bytes into the signature block and encrypt once more.
	a := [3]byte{'-', '-', '-'}
	copy(a[:], []byte(atom))
	sig[8] ^= a[0]
	sig[9] ^= a[1]
	sig[10] ^= a[2]
	var out [16]byte
	block.Encrypt(out[:], sig[:])
	return out
}

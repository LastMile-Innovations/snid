package snhash

import (
	"crypto/hmac"
	"crypto/sha256"

	"github.com/zeebo/blake3"
)

// HMACSHA256 signs data for public API compatibility.
func HMACSHA256(key, data []byte) (Digest, error) {
	mac := hmac.New(sha256.New, key)
	if _, err := mac.Write(data); err != nil {
		return Digest{}, err
	}
	return NewDigest(AlgorithmHMACSHA256, mac.Sum(nil))
}

// VerifyHMACSHA256 verifies a public-compatible HMAC-SHA256 signature.
func VerifyHMACSHA256(key, data []byte, expected Digest) bool {
	if expected.Algorithm != AlgorithmHMACSHA256 {
		return false
	}
	actual, err := HMACSHA256(key, data)
	if err != nil {
		return false
	}
	return hmac.Equal(actual.Bytes, expected.Bytes)
}

// KeyedBLAKE3 signs data with BLAKE3 keyed mode for fast internal service calls.
func KeyedBLAKE3(key, data []byte) (Digest, error) {
	hasher, err := blake3.NewKeyed(key)
	if err != nil {
		return Digest{}, err
	}
	if _, err := hasher.Write(data); err != nil {
		return Digest{}, err
	}
	return NewDigest(AlgorithmBLAKE3Keyed_256, hasher.Sum(nil))
}

// VerifyKeyedBLAKE3 verifies a keyed BLAKE3 digest.
func VerifyKeyedBLAKE3(key, data []byte, expected Digest) bool {
	if expected.Algorithm != AlgorithmBLAKE3Keyed_256 {
		return false
	}
	actual, err := KeyedBLAKE3(key, data)
	if err != nil {
		return false
	}
	return equalBytes(actual.Bytes, expected.Bytes)
}

package snidbridge

import (
	"errors"

	"github.com/LastMile-Innovations/snhash"
	snid "github.com/LastMile-Innovations/snid"
)

var ErrMissingBLAKE3 = errors.New("snhash/snidbridge: manifest has no blake3-256 digest")

// NewBIDFromDigest converts a BLAKE3-256 digest into a SNID BID.
func NewBIDFromDigest(digest snhash.Digest) (snid.BID, error) {
	if digest.Algorithm != snhash.AlgorithmBLAKE3_256 || digest.Size != 32 || len(digest.Bytes) != 32 {
		return snid.BID{}, ErrMissingBLAKE3
	}
	var content [32]byte
	copy(content[:], digest.Bytes)
	return snid.NewBID(content), nil
}

// AttachSNID creates a SNID BID from a manifest's BLAKE3-256 content digest.
func AttachSNID(manifest snhash.Manifest) (snid.BID, error) {
	digest, err := manifest.BLAKE3()
	if err != nil {
		return snid.BID{}, ErrMissingBLAKE3
	}
	return NewBIDFromDigest(digest)
}

package snhash

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidHashID = errors.New("snhash: invalid hash id")

// HashID is the compact, user-facing content identifier derived from a digest.
// It identifies bytes; SNID identifies topology/time.
type HashID struct {
	Profile Profile `json:"profile"`
	Size    int64   `json:"size"`
	Root    Digest  `json:"root"`
}

// NewHashID validates and copies a digest into a content identifier.
func NewHashID(profile Profile, size int64, root Digest) (HashID, error) {
	if !root.Trusted() {
		return HashID{}, ErrInvalidHashID
	}
	copied, err := NewDigest(root.Algorithm, root.Bytes)
	if err != nil {
		return HashID{}, err
	}
	return HashID{Profile: profile.normalized(), Size: size, Root: copied}, nil
}

// Wire returns hid1:<profile>:<size>:<algorithm>:<base32-lower-no-padding>.
func (id HashID) Wire() string {
	return hashIDWireVersion + ":" + string(id.Profile.normalized()) + ":" +
		strconv.FormatInt(id.Size, 10) + ":" + string(id.Root.Algorithm) + ":" + id.Root.Base32()
}

// String returns Wire.
func (id HashID) String() string {
	return id.Wire()
}

// Verify checks whether data hashes to the HashID root digest and size.
func (id HashID) Verify(data []byte) bool {
	return int64(len(data)) == id.Size && id.Root.Verify(data)
}

// ParseHashID parses hid1:<profile>:<size>:<algorithm>:<base32-lower-no-padding>.
func ParseHashID(wire string) (HashID, error) {
	parts := strings.Split(wire, ":")
	if len(parts) != 5 || parts[0] != hashIDWireVersion {
		return HashID{}, ErrInvalidHashID
	}
	size, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || size < 0 {
		return HashID{}, ErrInvalidHashID
	}
	digest, err := ParseDigestWire(wireVersion + ":" + parts[3] + ":" + parts[4])
	if err != nil {
		return HashID{}, err
	}
	return NewHashID(Profile(parts[1]), size, digest)
}

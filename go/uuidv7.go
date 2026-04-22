package snid

import (
	"time"

	"github.com/google/uuid"
)

// NewUUIDv7 generates a UUIDv7-compatible SNID (RFC 9562).
// The returned ID produces byte-for-byte identical output to
// reference implementations (.NET 9, uuid crate v7, Python 3.14 uuid7, PostgreSQL uuid_generate_v7()).
//
// This is a thin wrapper around NewFast() since the current generator
// already produces correct UUIDv7 byte layout.
func NewUUIDv7() ID {
	return NewFast()
}

// NewUUIDv7WithTime generates a UUIDv7-compatible SNID with a specific timestamp.
// This is useful for testing and migration scenarios where deterministic IDs are needed.
//
// The timestamp is in milliseconds since Unix epoch.
func NewUUIDv7WithTime(ts time.Time) ID {
	return NewDeterministicIngestID(uint64(ts.UnixMilli()), nil)
}

// GenerateV7 is an alias for NewUUIDv7() for API familiarity.
func GenerateV7() ID {
	return NewUUIDv7()
}

// ToUUIDv7 converts a SNID to standard UUID format.
// Returns the UUID as a github.com/google/uuid.UUID type for interoperability.
func (id ID) ToUUIDv7() uuid.UUID {
	var u uuid.UUID
	copy(u[:], id[:])
	return u
}

// FromUUIDv7 converts a standard UUID to SNID format.
// Only accepts UUIDv7 (version 7) UUIDs; returns error for other versions.
func FromUUIDv7(u uuid.UUID) (ID, error) {
	// Check version (bits 48-51 should be 0b0111 for v7)
	version := (u[6] >> 4) & 0x0F
	if version != 7 {
		return Zero, ErrInvalidFormat
	}

	var id ID
	copy(id[:], u[:])
	return id, nil
}

// UUIDString returns the ID in standard UUID string format (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx).
// This is compatible with RFC 9562 UUID string representation.
func (id ID) UUIDString() string {
	var u uuid.UUID
	copy(u[:], id[:])
	return u.String()
}

// ParseUUIDString parses a standard UUID string and returns a SNID.
// Accepts both hyphenated and non-hyphenated UUID formats.
func ParseUUIDString(s string) (ID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return Zero, err
	}
	return FromUUIDv7(u)
}

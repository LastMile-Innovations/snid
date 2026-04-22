package snid

const uuidTextLength = 36

// NewUUIDv7 generates a UUIDv7-compatible SNID (RFC 9562).
// The returned ID produces byte-for-byte identical output to
// reference implementations (.NET 9, uuid crate v7, Python 3.14 uuid7, PostgreSQL uuid_generate_v7()).
//
// This is a thin wrapper around NewFast() since the current generator
// already produces correct UUIDv7 byte layout.
func NewUUIDv7() ID {
	return NewFast()
}

// UUID is SNID's dependency-free 16-byte UUID representation.
type UUID [16]byte

// String returns the standard 36-character UUID representation.
func (u UUID) String() string {
	var out [uuidTextLength]byte
	writeUUIDText(out[:], u)
	return string(out[:])
}

// ParseUUID parses a standard 36-character UUID string.
func ParseUUID(s string) (UUID, error) {
	return ParseUUIDBytes(stringToBytes(s))
}

// ParseUUIDBytes parses a standard 36-character UUID byte slice.
func ParseUUIDBytes(text []byte) (UUID, error) {
	var u UUID
	if err := parseUUIDBytes(text, &u); err != nil {
		return UUID{}, err
	}
	return u, nil
}

// ToUUIDv7 converts a SNID to standard UUID format.
func (id ID) ToUUIDv7() UUID {
	return UUID(id)
}

// FromUUIDv7 converts a standard UUID to SNID format.
// Only accepts UUIDv7 (version 7) UUIDs; returns error for other versions.
func FromUUIDv7(u UUID) (ID, error) {
	if !isUUIDv7(u) {
		return Zero, ErrInvalidFormat
	}
	return ID(u), nil
}

// UUIDString returns the ID in standard UUID string format (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx).
// This is compatible with RFC 9562 UUID string representation.
func (id ID) UUIDString() string {
	return UUID(id).String()
}

// ParseUUIDString parses a standard UUID string and returns a SNID.
func ParseUUIDString(s string) (ID, error) {
	u, err := ParseUUID(s)
	if err != nil {
		return Zero, err
	}
	return FromUUIDv7(u)
}

func writeUUIDText(dst []byte, src UUID) {
	const hex = "0123456789abcdef"

	j := 0
	for i, b := range src {
		if i == 4 || i == 6 || i == 8 || i == 10 {
			dst[j] = '-'
			j++
		}
		dst[j] = hex[b>>4]
		dst[j+1] = hex[b&0x0F]
		j += 2
	}
}

func parseUUIDBytes(src []byte, dst *UUID) error {
	if len(src) != uuidTextLength {
		return ErrInvalidLength
	}
	if src[8] != '-' || src[13] != '-' || src[18] != '-' || src[23] != '-' {
		return ErrInvalidFormat
	}

	j := 0
	for i := 0; i < 16; i++ {
		if j == 8 || j == 13 || j == 18 || j == 23 {
			j++
		}
		hi, ok := uuidHexValue(src[j])
		if !ok {
			return ErrInvalidFormat
		}
		lo, ok := uuidHexValue(src[j+1])
		if !ok {
			return ErrInvalidFormat
		}
		dst[i] = (hi << 4) | lo
		j += 2
	}
	return nil
}

func uuidHexValue(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	default:
		return 0, false
	}
}

func isUUIDv7(u UUID) bool {
	return (u[6]>>4) == 7 && (u[8]&0xC0) == 0x80
}

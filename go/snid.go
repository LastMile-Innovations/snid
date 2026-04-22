// Package snid provides a polyglot sortable identifier protocol.
//
// SNID defines 128-bit time-ordered identifiers compatible with UUID v7,
// plus extended families for spatial, neural, ledger, and capability use cases.
//
// Basic usage:
//
//	id := snid.New(snid.Matter)
//	wire := id.String(snid.Matter)
//	parsed, atom, err := snid.FromString(wire)
//
// Extended identifier families:
//
//	SGID - Spatial IDs with H3 geospatial encoding
//	NID  - Neural IDs with semantic tail for vector search
//	LID  - Ledger IDs with HMAC verification tail
//	WID  - World/scenario IDs for simulation isolation
//	XID  - Edge IDs for relationship identity
//	KID  - Capability IDs with MAC-based verification
//	EID  - Ephemeral 64-bit IDs
//	BID  - Content-addressable IDs (topology + content hash)
//	AKID - Dual-part public+secret credentials
//
// For protocol specification, see https://github.com/LastMile-Innovations/snid/blob/main/docs/SPEC.md
package snid

import (
	"encoding/binary"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/uber/h3-go/v4"
)

// IDType defines the structural mode (The "How").
type IDType uint8

const (
	TypeTime    IDType = iota // K-Sortable Time (v7)
	TypeSpatial               // Geospatial H3 (v8)
	TypeCatalog               // Taxonomy Hash (v9)
	TypeAsset                 // Instance Hash (v10)
)

const MaxAtomLength = 8
const MaxPayloadLength = 24

const (
	spatialMarkerByte = 0xA5
	spatialTailNibble = 0xA0
)

var (
	ErrInvalidFormat   = errors.New("snid: invalid format")
	ErrInvalidLength   = errors.New("snid: invalid length")
	ErrChecksum        = errors.New("snid: checksum mismatch")
	ErrInvalidAtom     = errors.New("snid: invalid atom")
	ErrPayloadTooShort = errors.New("snid: payload too short")
	ErrPayloadTooLong  = errors.New("snid: payload too long")
	ErrNilID           = errors.New("snid: nil id")
	ErrTooOld          = errors.New("snid: id is too old")
)

type ID [16]byte

var Zero ID

// New generates a time-ordered ID. The atom parameter is accepted for API
// consistency but is NOT embedded in the ID bytes—atoms are type-tags applied
// at serialization time via id.String(atom). Use NewFast() for hot paths.
func New(atom Atom) ID { return NewFast() }

// --- DOMAIN CONSTRUCTORS (Context-Aware) ---

// NewSpatial generates a Location-Ordered ID (v8).
// Use for: Buildings, Land, Static Sensors.
// Speed: ~20ns (Includes H3 Math).
func NewSpatial(lat, lng float64) ID {
	InitAdaptive()
	return adaptive.nextSpatial(lat, lng, 12) // Default Res 12 (~10m)
}

// NewSpatialPrecise allows specifying resolution.
func NewSpatialPrecise(lat, lng float64, res int) ID {
	InitAdaptive()
	return adaptive.nextSpatial(lat, lng, res)
}

// NewAsset generates an Instance-Ordered ID (v10).
// Use for: Physical Inventory, Serialized Items.
// Speed: ~15ns (Includes Hashing).
func NewAsset(catalogID ID, tenantID string, serial string) ID {
	return generateAssetID(catalogID, tenantID, serial)
}

// NewCatalog generates a Taxonomy-Ordered ID (v9).
// Use for: Product Definitions, SKU Bases.
// Speed: ~15ns.
func NewCatalog(category, brand, specs string) ID {
	return generateCatalogID(category, brand, specs)
}

// --- METHODS ---

// Type returns the inferred structural type for this ID.
func (id ID) Type() IDType {
	if id.IsSpatial() {
		return TypeSpatial
	}
	v := id.Version()
	switch v {
	case 9:
		return TypeCatalog
	case 10:
		return TypeAsset
	default:
		return TypeTime
	}
}

// IsType reports whether the ID matches the provided structural type.
func (id ID) IsType(t IDType) bool {
	return id.Type() == t
}

// IsZero reports whether all 16 bytes of the ID are zero.
func (id ID) IsZero() bool {
	return id == ID{}
}

// Version returns the UUID/SNID version nibble.
func (id ID) Version() int { return int(id[6] >> 4) }

// Time decodes the millisecond timestamp prefix from a time-ordered ID.
// The top 48 bits of the high word are the ms timestamp; >> 16 strips the
// version nibble and the low 12 bits of sequence. Emits MOVQ + SHR on amd64/arm64.
func (id ID) Time() time.Time {
	return time.UnixMilli(int64(binary.BigEndian.Uint64(id[:8]) >> 16))
}

// UUID converts ID to a UUID value.
func (id ID) UUID() uuid.UUID { return uuid.UUID(id) }

// Bytes exposes the raw ID bytes.
func (id ID) Bytes() []byte { return id[:] }

// Compare performs lexicographic ordering across full 128-bit IDs.
func (id ID) Compare(other ID) int {
	h1, h2 := binary.BigEndian.Uint64(id[:8]), binary.BigEndian.Uint64(other[:8])
	if h1 != h2 {
		if h1 < h2 {
			return -1
		}
		return 1
	}
	l1, l2 := binary.BigEndian.Uint64(id[8:]), binary.BigEndian.Uint64(other[8:])
	if l1 < l2 {
		return -1
	}
	if l1 > l2 {
		return 1
	}
	return 0
}

// CompareFast is a hot-path alias of Compare.
func (id ID) CompareFast(other ID) int { return id.Compare(other) }

// Before reports whether id sorts before other.
func (id ID) Before(other ID) bool { return id.Compare(other) < 0 }

// After reports whether id sorts after other.
func (id ID) After(other ID) bool { return id.Compare(other) > 0 }

// Debug returns a human-readable diagnostic string for this ID.
// Uses string concatenation (single concatstrings allocation) instead of
// fmt.Sprintf to avoid interface boxing and reflection on id[:].
func (id ID) Debug() string {
	var hexBuf [32]byte
	for i, b := range id {
		hexBuf[i*2] = hexChars[b>>4]
		hexBuf[i*2+1] = hexChars[b&0x0F]
	}
	return "SNID(Time: " + id.Time().Format(time.RFC3339Nano) +
		", Ver: " + strconv.Itoa(id.Version()) +
		", Bytes: " + string(hexBuf[:]) + ")"
}

// IsSpatial checks if ID is spatial (Version 8).
func (id ID) IsSpatial() bool {
	return id.Version() == 8 && id[14] == spatialMarkerByte && (id[15]&0xF0) == spatialTailNibble
}

// ExtractLocation returns the H3 index if spatial.
func (id ID) ExtractLocation() h3.Cell {
	if !id.IsSpatial() {
		return 0
	}
	var raw [8]byte
	copy(raw[:], id[:8])
	// Spatial IDs store the original high nibble of byte 6 in the low nibble of byte 15.
	raw[6] = ((id[15] & 0x0F) << 4) | (raw[6] & 0x0F)
	return h3.Cell(binary.BigEndian.Uint64(raw[:]))
}

// FromBytes parses a raw 16-byte ID payload.
func FromBytes(b []byte) (ID, error) {
	if len(b) != 16 {
		return Zero, ErrInvalidLength
	}
	var id ID
	copy(id[:], b)
	return id, nil
}

// FromUUID converts a UUID into ID.
func FromUUID(u uuid.UUID) ID { return ID(u) }

// FromString parses an atom-prefixed SNID string.
func FromString(s string) (ID, Atom, error) {
	var id ID
	atom, err := id.Parse(s)
	return id, atom, err
}

// ParseWithFormat parses a wire string and returns the detected delimiter format.
func ParseWithFormat(s string) (ID, Atom, WireFormat, error) {
	var id ID
	atom, f, err := id.parseWithFormat(s)
	return id, atom, f, err
}

// IsValidAtom validates an atom against supported tags.
func IsValidAtom(atom Atom) bool {
	switch atom {
	case Identity, Tenant, Matter, Space, Time, Ledger, Legal, Trust, Kinetic, Cognition, Semantic, System, Vault, Key, Event, Session:
		return true
	case LegacyObject, LegacyTransaction, LegacySchedule, LegacyNetwork, LegacyOperations, LegacyPersona, LegacyGroup, LegacyBio, LegacyAtmosphere:
		return true
	default:
		return false
	}
}

// Parse decodes an atom-prefixed SNID string into this ID.
func (id *ID) Parse(s string) (Atom, error) {
	atom, _, err := id.parseWithFormat(s)
	return atom, err
}

func (id *ID) parseWithFormat(s string) (Atom, WireFormat, error) {
	if id == nil {
		return "", WireColon, ErrNilID
	}
	src := stringToBytes(s)
	if len(src) < 4 {
		return "", WireColon, ErrInvalidFormat
	}

	// Locate atom delimiter (`:` or `_`) near prefix start.
	delimIdx := -1
	delim := byte(':')
	for i := 0; i < len(src) && i <= MaxAtomLength; i++ {
		if src[i] == ':' || src[i] == '_' {
			delimIdx = i
			delim = src[i]
			break
		}
	}
	if delimIdx < 2 {
		return "", WireColon, ErrInvalidFormat
	}
	format := formatFromDelimiter(delim)
	if format == WireUnderscore && !AcceptUnderscore() {
		return "", format, ErrInvalidFormat
	}

	atom := Atom(s[:delimIdx])
	if !IsValidAtom(atom) {
		return "", format, ErrInvalidAtom
	}
	atom = CanonicalAtom(atom)

	payload := src[delimIdx+1:]
	if len(payload) < 2 {
		return "", format, ErrPayloadTooShort
	}
	if len(payload) > MaxPayloadLength {
		return "", format, ErrPayloadTooLong
	}

	dataLen := len(payload) - 1
	data := payload[:dataLen]
	if err := decode16Base58(id, data); err != nil {
		return "", format, err
	}
	if !isCanonicalBase58Data(data, *id) {
		return "", format, ErrInvalidFormat
	}

	if payload[dataLen] != base58Alphabet[crc8(id[:])%58] {
		return "", format, ErrChecksum
	}

	countParsedFormat(format)
	return atom, format, nil
}

// ParseCompact decodes payload-only SNID strings without atom prefix.
func (id *ID) ParseCompact(s string) error {
	if id == nil {
		return ErrNilID
	}
	src := stringToBytes(s)
	if len(src) < 2 {
		return ErrPayloadTooShort
	}
	if len(src) > MaxPayloadLength {
		return ErrPayloadTooLong
	}
	dataLen := len(src) - 1
	data := src[:dataLen]
	if err := decode16Base58(id, data); err != nil {
		return err
	}
	if !isCanonicalBase58Data(data, *id) {
		return ErrInvalidFormat
	}
	if src[dataLen] != base58Alphabet[crc8(id[:])%58] {
		return ErrChecksum
	}
	return nil
}

func isCanonicalBase58Data(src []byte, id ID) bool {
	leadingOnes := 0
	for leadingOnes < len(src) && src[leadingOnes] == '1' {
		leadingOnes++
	}
	leadingZeros := 0
	for leadingZeros < len(id) && id[leadingZeros] == 0 {
		leadingZeros++
	}
	return leadingOnes == leadingZeros
}

// ParseBatch parses many SNID strings in order.
func ParseBatch(srcs []string) ([]ID, []Atom, error) {
	ids, atoms := make([]ID, len(srcs)), make([]Atom, len(srcs))
	for i, s := range srcs {
		atom, err := ids[i].Parse(s)
		if err != nil {
			return nil, nil, err
		}
		atoms[i] = atom
	}
	return ids, atoms, nil
}

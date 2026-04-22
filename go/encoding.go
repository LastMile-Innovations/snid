package snid

import (
	"database/sql/driver"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/bits"
	"unsafe"

	"github.com/google/uuid"
)

// Removed unused bufferPool - zero-alloc paths use stack allocation

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var crc8Table = [256]byte{
	0x00, 0x07, 0x0E, 0x09, 0x1C, 0x1B, 0x12, 0x15, 0x38, 0x3F, 0x36, 0x31, 0x24, 0x23, 0x2A, 0x2D,
	0x70, 0x77, 0x7E, 0x79, 0x6C, 0x6B, 0x62, 0x65, 0x48, 0x4F, 0x46, 0x41, 0x54, 0x53, 0x5A, 0x5D,
	0xE0, 0xE7, 0xEE, 0xE9, 0xFC, 0xFB, 0xF2, 0xF5, 0xD8, 0xDF, 0xD6, 0xD1, 0xC4, 0xC3, 0xCA, 0xCD,
	0x90, 0x97, 0x9E, 0x99, 0x8C, 0x8B, 0x82, 0x85, 0xA8, 0xAF, 0xA6, 0xA1, 0xB4, 0xB3, 0xBA, 0xBD,
	0xC7, 0xC0, 0xC9, 0xCE, 0xDB, 0xDC, 0xD5, 0xD2, 0xFF, 0xF8, 0xF1, 0xF6, 0xE3, 0xE4, 0xED, 0xEA,
	0xB7, 0xB0, 0xB9, 0xBE, 0xAB, 0xAC, 0xA5, 0xA2, 0x8F, 0x88, 0x81, 0x86, 0x93, 0x94, 0x9D, 0x9A,
	0x27, 0x20, 0x29, 0x2E, 0x3B, 0x3C, 0x35, 0x32, 0x1F, 0x18, 0x11, 0x16, 0x03, 0x04, 0x0D, 0x0A,
	0x57, 0x50, 0x59, 0x5E, 0x4B, 0x4C, 0x45, 0x42, 0x6F, 0x68, 0x61, 0x66, 0x73, 0x74, 0x7D, 0x7A,
	0x89, 0x8E, 0x87, 0x80, 0x95, 0x92, 0x9B, 0x9C, 0xB1, 0xB6, 0xBF, 0xB8, 0xAD, 0xAA, 0xA3, 0xA4,
	0xF9, 0xFE, 0xF7, 0xF0, 0xE5, 0xE2, 0xEB, 0xEC, 0xC1, 0xC6, 0xCF, 0xC8, 0xDD, 0xDA, 0xD3, 0xD4,
	0x69, 0x6E, 0x67, 0x60, 0x75, 0x72, 0x7B, 0x7C, 0x51, 0x56, 0x5F, 0x58, 0x4D, 0x4A, 0x43, 0x44,
	0x19, 0x1E, 0x17, 0x10, 0x05, 0x02, 0x0B, 0x0C, 0x21, 0x26, 0x2F, 0x28, 0x3D, 0x3A, 0x33, 0x34,
	0x4E, 0x49, 0x40, 0x47, 0x52, 0x55, 0x5C, 0x5B, 0x76, 0x71, 0x78, 0x7F, 0x6A, 0x6D, 0x64, 0x63,
	0x3E, 0x39, 0x30, 0x37, 0x22, 0x25, 0x2C, 0x2B, 0x06, 0x01, 0x08, 0x0F, 0x1A, 0x1D, 0x14, 0x13,
	0xAE, 0xA9, 0xA0, 0xA7, 0xB2, 0xB5, 0xBC, 0xBB, 0x96, 0x91, 0x98, 0x9F, 0x8A, 0x8D, 0x84, 0x83,
	0xDE, 0xD9, 0xD0, 0xD7, 0xC2, 0xC5, 0xCC, 0xCB, 0xE6, 0xE1, 0xE8, 0xEF, 0xFA, 0xFD, 0xF4, 0xF3,
}

var b58Map [256]int8

// init implements the corresponding operation.
func init() {
	for i := range b58Map {
		b58Map[i] = -1
	}
	for i, c := range base58Alphabet {
		b58Map[c] = int8(i)
	}
}

// crc8 implements the corresponding operation.
func crc8(data []byte) byte {
	crc := byte(0)
	if len(data) == 16 {
		// Unrolled for performance in the common case
		crc = crc8Table[crc^data[0]]
		crc = crc8Table[crc^data[1]]
		crc = crc8Table[crc^data[2]]
		crc = crc8Table[crc^data[3]]
		crc = crc8Table[crc^data[4]]
		crc = crc8Table[crc^data[5]]
		crc = crc8Table[crc^data[6]]
		crc = crc8Table[crc^data[7]]
		crc = crc8Table[crc^data[8]]
		crc = crc8Table[crc^data[9]]
		crc = crc8Table[crc^data[10]]
		crc = crc8Table[crc^data[11]]
		crc = crc8Table[crc^data[12]]
		crc = crc8Table[crc^data[13]]
		crc = crc8Table[crc^data[14]]
		crc = crc8Table[crc^data[15]]
		return crc
	}
	for _, b := range data {
		crc = crc8Table[crc^b]
	}
	return crc
}

// String formats an ID using the default wire format (`SNID_WIRE_OUTPUT_FORMAT`).
func (id ID) String(atom Atom) string {
	return id.StringWithFormat(atom, DefaultWireFormat())
}

// StringWithFormat formats an ID as `<atom><delim><base58+checksum>`.
func (id ID) StringWithFormat(atom Atom, f WireFormat) string {
	var buf [48]byte
	sAtom := string(atom)
	n := copy(buf[:], sAtom)
	buf[n] = delimiterForFormat(f)
	res := id.appendPayload(buf[:n+1])
	return bytesToString(res)
}

// StringFast is a hot-path alias of String.
func (id ID) StringFast(atom Atom) string { return id.String(atom) }

// Format implements fmt.Formatter and avoids reflective formatting for IDs.
// `%s`/`%v` use the default wire format with the `Matter` atom.
func (id ID) Format(state fmt.State, verb rune) {
	switch verb {
	case 'x', 'X':
		var buf [32]byte
		const lower = "0123456789abcdef"
		const upper = "0123456789ABCDEF"
		hex := lower
		if verb == 'X' {
			hex = upper
		}
		for i, b := range id {
			buf[i*2] = hex[b>>4]
			buf[i*2+1] = hex[b&0x0F]
		}
		_, _ = state.Write(buf[:])
	default:
		_, _ = state.Write(stringToBytes(id.StringWithFormat(Matter, DefaultWireFormat())))
	}
}

// StringCompact formats only the payload portion without atom prefix.
func (id ID) StringCompact() string {
	var buf [24]byte
	res := id.appendPayload(buf[:0])
	return bytesToString(res)
}

// AppendTo appends the atom-prefixed encoded ID using the default wire format.
func (id ID) AppendTo(dst []byte, atom Atom) []byte {
	return id.AppendToWithFormat(dst, atom, DefaultWireFormat())
}

// AppendToWithFormat appends the atom-prefixed encoded ID using the specified wire format.
func (id ID) AppendToWithFormat(dst []byte, atom Atom, f WireFormat) []byte {
	dst = append(dst, string(atom)...)
	dst = append(dst, delimiterForFormat(f))
	return id.appendPayload(dst)
}

// appendPayload encodes id as base58 + checksum and appends to dst.
// Uses 128-bit integer division: 44 divisions total vs 352 in the
// byte-by-byte approach, while still benefiting from compiler strength
// reduction for the constant divisor 58 on the hi word.
func (id ID) appendPayload(dst []byte) []byte {
	var buf [24]byte
	idx := 23

	// Checksum at the final position
	chk := crc8(id[:])
	buf[idx] = base58Alphabet[chk%58]
	idx--

	// Treat the 16-byte ID as a 128-bit big-endian integer (hi:lo).
	hi := binary.BigEndian.Uint64(id[:8])
	lo := binary.BigEndian.Uint64(id[8:])

	// Repeatedly divide (hi:lo) by 58, recording remainders right-to-left.
	// Each iteration:
	//   qhi  = hi / 58              (compiler strength-reduces constant 58)
	//   rhi  = hi % 58              (free from the division above)
	//   qlo, rem = (rhi:lo) / 58    (bits.Div64: safe since rhi < 58)
	// Precondition for bits.Div64: hi_arg < divisor → rhi < 58 ✓
	// Max quotient: (57·2⁶⁴ + 2⁶⁴-1) / 58 = 2⁶⁴-1, fits in uint64 ✓
	for hi > 0 || lo > 0 {
		qhi := hi / 58
		rhi := hi - qhi*58
		qlo, rem := bits.Div64(rhi, lo, 58)
		hi, lo = qhi, qlo
		buf[idx] = base58Alphabet[rem]
		idx--
	}

	// Pad leading zero bytes with '1' (base58 convention)
	for i := 0; i < 16 && id[i] == 0; i++ {
		buf[idx] = '1'
		idx--
	}

	return append(dst, buf[idx+1:]...)
}

// decode16Base58 implements the corresponding operation.
func decode16Base58(id *ID, src []byte) error {
	*id = Zero
	for _, c := range src {
		val := b58Map[c]
		if val == -1 {
			return errors.New("snid: invalid char")
		}
		carry := uint32(val)
		for i := 15; i >= 0; i-- {
			res := uint32(id[i])*58 + carry
			id[i] = byte(res)
			carry = res >> 8
		}
		if carry > 0 {
			return errors.New("snid: overflow")
		}
	}
	return nil
}

// _encode12Base58 implements the corresponding operation.
func _encode12Base58(src [12]byte) string {
	var buf [17]byte
	idx, input, zeros := 16, src, 0
	for zeros < 12 && input[zeros] == 0 {
		zeros++
	}
	for zeros < 12 {
		rem := uint32(0)
		for i := zeros; i < 12; i++ {
			val := (rem << 8) + uint32(input[i])
			input[i] = byte(val / 58)
			rem = val % 58
		}
		buf[idx] = base58Alphabet[rem]
		idx--
		if input[zeros] == 0 {
			zeros++
		}
	}
	for i := 0; i < 12 && src[i] == 0; i++ {
		buf[idx] = '1'
		idx--
	}
	return bytesToString(buf[idx+1:])
}

// --- VOICE (PROQUINT) ---

var (
	pqConsonants = []byte("bdfghjklmnprstvz")
	pqVowels     = []byte("aiou")
)

// ToVoice formats the tail bytes as a proquint voice-friendly token.
func (id ID) ToVoice(atom Atom) string {
	w1 := uint16(id[12])<<8 | uint16(id[13])
	w2 := uint16(id[14])<<8 | uint16(id[15])
	var b [24]byte
	idx := copy(b[:], string(atom))
	b[idx] = ':'
	idx++
	writeProquint(b[idx:], w1)
	b[idx+5] = '-'
	writeProquint(b[idx+6:], w2)
	return string(b[:idx+11])
}

// writeProquint implements the corresponding operation.
func writeProquint(b []byte, v uint16) {
	// Proquint: 16 bits -> CVCVC (4+2+4+2+4 = 16 bits)
	// Layout: C(15:12) V(11:10) C(9:6) V(5:4) C(3:0)
	b[0] = pqConsonants[(v>>12)&0x0F] // bits 15-12 (4 bits)
	b[1] = pqVowels[(v>>10)&0x03]     // bits 11-10 (2 bits)
	b[2] = pqConsonants[(v>>6)&0x0F]  // bits 9-6  (4 bits)
	b[3] = pqVowels[(v>>4)&0x03]      // bits 5-4  (2 bits)
	b[4] = pqConsonants[v&0x0F]       // bits 3-0  (4 bits)
}

// --- INTERFACES ---

// MarshalJSON encodes ID as a UUID string.
func (id ID) MarshalJSON() ([]byte, error) { return json.Marshal(id.UUID().String()) }

// UnmarshalJSON decodes either UUID or SNID textual JSON representations.
func (id *ID) UnmarshalJSON(data []byte) error {
	if id == nil {
		return ErrNilID
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if len(s) == 36 {
		u, err := uuid.Parse(s)
		if err == nil {
			*id = ID(u)
			return nil
		}
	}
	_, err := id.Parse(s)
	return err
}

// MarshalBinary returns the 16-byte raw ID representation.
func (id ID) MarshalBinary() ([]byte, error) { return id[:], nil }

// UnmarshalBinary decodes a 16-byte binary ID.
func (id *ID) UnmarshalBinary(data []byte) error {
	if id == nil {
		return ErrNilID
	}
	if len(data) != 16 {
		return ErrInvalidLength
	}
	copy(id[:], data)
	return nil
}

// MarshalText encodes ID as UUID text.
func (id ID) MarshalText() ([]byte, error) { return []byte(id.UUID().String()), nil }

// UnmarshalText parses UUID text into ID bytes.
func (id *ID) UnmarshalText(text []byte) error {
	if id == nil {
		return ErrNilID
	}
	u, err := uuid.ParseBytes(text)
	if err != nil {
		return err
	}
	*id = ID(u)
	return nil
}

// MarshalProto returns protobuf-ready raw ID bytes.
func (id ID) MarshalProto() ([]byte, error) { return id[:], nil }

// UnmarshalProto decodes protobuf raw ID bytes.
func (id *ID) UnmarshalProto(data []byte) error {
	if id == nil {
		return ErrNilID
	}
	if len(data) != 16 {
		return errors.New("snid: invalid proto length")
	}
	copy(id[:], data)
	return nil
}

// ProtoSize returns the fixed encoded protobuf byte size.
func (id ID) ProtoSize() int { return 16 }

// Value implements database/sql Valuer for persisting IDs.
func (id ID) Value() (driver.Value, error) {
	if id.IsZero() {
		return nil, nil
	}
	raw := make([]byte, 16)
	copy(raw, id[:])
	return raw, nil
}

// Scan implements database/sql Scanner for UUID, SNID, and raw-byte inputs.
func (id *ID) Scan(src any) error {
	if id == nil {
		return ErrNilID
	}
	switch v := src.(type) {
	case nil:
		*id = Zero
		return nil
	case []byte:
		if len(v) == 16 {
			copy(id[:], v)
			return nil
		}
		s := bytesToString(v)
		if len(s) == 36 {
			u, err := uuid.Parse(s)
			if err == nil {
				*id = ID(u)
				return nil
			}
		}
		_, err := id.Parse(s)
		return err
	case string:
		if len(v) == 36 {
			u, err := uuid.Parse(v)
			if err == nil {
				*id = ID(u)
				return nil
			}
		}
		_, err := id.Parse(v)
		return err
	}
	return fmt.Errorf("snid: cannot scan %T", src)
}

// bytesToString implements the corresponding operation.
func bytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// stringToBytes implements the corresponding operation.
func stringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// encodeBase58Bytes encodes arbitrary bytes as canonical Bitcoin-style Base58.
// It is used for non-SNID payloads such as short IDs and AKID secrets. The
// 16-byte SNID wire path remains specialized in appendPayload.
func encodeBase58Bytes(src []byte) string {
	if len(src) == 0 {
		return ""
	}

	zeros := 0
	for zeros < len(src) && src[zeros] == 0 {
		zeros++
	}

	var scratchStack [64]byte
	scratch := scratchStack[:]
	if len(src) > len(scratchStack) {
		scratch = make([]byte, len(src))
	} else {
		scratch = scratch[:len(src)]
	}
	copy(scratch, src)

	maxEncoded := len(src)*138/100 + 2
	var encodedStack [96]byte
	encoded := encodedStack[:]
	if maxEncoded > len(encodedStack) {
		encoded = make([]byte, maxEncoded)
	} else {
		encoded = encoded[:maxEncoded]
	}
	idx := len(encoded)

	for start := zeros; start < len(scratch); {
		var carry uint32
		for i := start; i < len(scratch); i++ {
			val := (carry << 8) | uint32(scratch[i])
			scratch[i] = byte(val / 58)
			carry = val % 58
		}
		idx--
		encoded[idx] = base58Alphabet[carry]
		for start < len(scratch) && scratch[start] == 0 {
			start++
		}
	}

	for i := 0; i < zeros; i++ {
		idx--
		encoded[idx] = '1'
	}

	return string(encoded[idx:])
}

// decodeBase58Bytes decodes base58 string to bytes without checksum validation.
func decodeBase58Bytes(s string) ([]byte, error) {
	if len(s) == 0 {
		return nil, nil
	}

	leadingOnes := 0
	for i := 0; i < len(s) && s[i] == '1'; i++ {
		leadingOnes++
	}

	maxDecoded := len(s)*733/1000 + 1
	var decodedStack [64]byte
	decoded := decodedStack[:]
	if maxDecoded > len(decodedStack) {
		decoded = make([]byte, maxDecoded)
	} else {
		decoded = decoded[:maxDecoded]
	}

	length := 0
	for i := leadingOnes; i < len(s); i++ {
		c := s[i]
		val := b58Map[c]
		if val == -1 {
			return nil, errors.New("snid: invalid char")
		}
		carry := int(val)
		j := 0
		for k := len(decoded) - 1; (carry != 0 || j < length) && k >= 0; k-- {
			carry += int(decoded[k]) * 58
			decoded[k] = byte(carry)
			carry >>= 8
			j++
		}
		if carry != 0 {
			return nil, errors.New("snid: overflow")
		}
		length = j
	}

	start := len(decoded) - length
	for start < len(decoded) && decoded[start] == 0 {
		start++
	}

	out := make([]byte, leadingOnes+len(decoded)-start)
	copy(out[leadingOnes:], decoded[start:])
	return out, nil
}

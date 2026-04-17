package snid

import "unsafe"

// Crockford's Base32 Alphabet (No I, L, O, U)
const turboAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

var turboMap [256]byte

// init implements the corresponding operation.
func init() {
	for i := range turboMap {
		turboMap[i] = 0xFF
	}
	for i := range len(turboAlphabet) {
		turboMap[turboAlphabet[i]] = byte(i)
		// Support lower case decoding
		if turboAlphabet[i] >= 'A' && turboAlphabet[i] <= 'Z' {
			turboMap[turboAlphabet[i]+32] = byte(i)
		}
	}
}

// StringTurbo returns a Base32 encoded string using raw unrolled bit blasting.
func (id ID) StringTurbo(atom Atom) string {
	var b [32]byte

	sAtom := string(atom)
	delim := delimiterForFormat(DefaultWireFormat())
	if len(sAtom) == 3 {
		b[0] = sAtom[0]
		b[1] = sAtom[1]
		b[2] = sAtom[2]
		b[3] = delim
	} else {
		n := copy(b[:], sAtom)
		b[n] = delim
	}

	// Base32 Encode (128 bits -> 26 chars)
	b[4] = turboAlphabet[id[0]>>3]
	b[5] = turboAlphabet[((id[0]&7)<<2)|(id[1]>>6)]
	b[6] = turboAlphabet[(id[1]>>1)&0x1F]
	b[7] = turboAlphabet[((id[1]&1)<<4)|(id[2]>>4)]
	b[8] = turboAlphabet[((id[2]&0xF)<<1)|(id[3]>>7)]
	b[9] = turboAlphabet[(id[3]>>2)&0x1F]
	b[10] = turboAlphabet[((id[3]&3)<<3)|(id[4]>>5)]
	b[11] = turboAlphabet[id[4]&0x1F]

	b[12] = turboAlphabet[id[5]>>3]
	b[13] = turboAlphabet[((id[5]&7)<<2)|(id[6]>>6)]
	b[14] = turboAlphabet[(id[6]>>1)&0x1F]
	b[15] = turboAlphabet[((id[6]&1)<<4)|(id[7]>>4)]
	b[16] = turboAlphabet[((id[7]&0xF)<<1)|(id[8]>>7)]
	b[17] = turboAlphabet[(id[8]>>2)&0x1F]
	b[18] = turboAlphabet[((id[8]&3)<<3)|(id[9]>>5)]
	b[19] = turboAlphabet[id[9]&0x1F]

	b[20] = turboAlphabet[id[10]>>3]
	b[21] = turboAlphabet[((id[10]&7)<<2)|(id[11]>>6)]
	b[22] = turboAlphabet[(id[11]>>1)&0x1F]
	b[23] = turboAlphabet[((id[11]&1)<<4)|(id[12]>>4)]
	b[24] = turboAlphabet[((id[12]&0xF)<<1)|(id[13]>>7)]
	b[25] = turboAlphabet[(id[13]>>2)&0x1F]
	b[26] = turboAlphabet[((id[13]&3)<<3)|(id[14]>>5)]
	b[27] = turboAlphabet[id[14]&0x1F]
	b[28] = turboAlphabet[id[15]>>3]
	b[29] = turboAlphabet[(id[15]&7)<<2]

	return unsafe.String(&b[0], 30)
}

// ParseTurbo decodes a Base32 machine string in ~20ns.
func (id *ID) ParseTurbo(s string) (Atom, error) {
	if len(s) != 30 || (s[3] != ':' && s[3] != '_') {
		return "", ErrInvalidFormat
	}
	if s[3] == '_' && !AcceptUnderscore() {
		return "", ErrInvalidFormat
	}
	countParsedFormat(formatFromDelimiter(s[3]))

	atom := Atom(s[:3])
	if !IsValidAtom(atom) {
		return "", ErrInvalidAtom
	}
	atom = CanonicalAtom(atom)

	src := s[4:]

	// Validate all characters are in turboMap
	for i := range 26 {
		if turboMap[src[i]] == 0xFF {
			return "", ErrInvalidFormat
		}
	}
	// Canonical form requires the final symbol's low 2 bits to be zero
	// because only the top 3 bits carry payload for byte 15.
	if turboMap[src[25]]&0x03 != 0 {
		return "", ErrInvalidFormat
	}

	// Byte 0
	v0 := turboMap[src[0]]
	v1 := turboMap[src[1]]
	id[0] = (v0 << 3) | (v1 >> 2)

	// Byte 1
	v2 := turboMap[src[2]]
	id[1] = (v1 << 6) | (v2 << 1) | (turboMap[src[3]] >> 4)

	// Byte 2
	v3 := turboMap[src[3]]
	v4 := turboMap[src[4]]
	id[2] = (v3 << 4) | (v4 >> 1)

	// Byte 3
	v5 := turboMap[src[5]]
	id[3] = (v4 << 7) | (v5 << 2) | (turboMap[src[6]] >> 3)

	// Byte 4
	v6 := turboMap[src[6]]
	id[4] = (v6 << 5) | turboMap[src[7]]

	// Byte 5
	id[5] = (turboMap[src[8]] << 3) | (turboMap[src[9]] >> 2)

	// Byte 6
	v10 := turboMap[src[10]]
	id[6] = (turboMap[src[9]] << 6) | (v10 << 1) | (turboMap[src[11]] >> 4)

	// Byte 7
	id[7] = (turboMap[src[11]] << 4) | (turboMap[src[12]] >> 1)

	// Byte 8
	id[8] = (turboMap[src[12]] << 7) | (turboMap[src[13]] << 2) | (turboMap[src[14]] >> 3)

	// Byte 9
	id[9] = (turboMap[src[14]] << 5) | turboMap[src[15]]

	// Byte 10
	id[10] = (turboMap[src[16]] << 3) | (turboMap[src[17]] >> 2)

	// Byte 11
	id[11] = (turboMap[src[17]] << 6) | (turboMap[src[18]] << 1) | (turboMap[src[19]] >> 4)

	// Byte 12
	id[12] = (turboMap[src[19]] << 4) | (turboMap[src[20]] >> 1)

	// Byte 13
	id[13] = (turboMap[src[20]] << 7) | (turboMap[src[21]] << 2) | (turboMap[src[22]] >> 3)

	// Byte 14
	id[14] = (turboMap[src[22]] << 5) | turboMap[src[23]]

	// Byte 15
	id[15] = (turboMap[src[24]] << 3) | (turboMap[src[25]] >> 2)

	return atom, nil
}

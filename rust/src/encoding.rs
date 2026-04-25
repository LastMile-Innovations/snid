//! Base58 encoding, wire format, checksum, and atom handling.

use crate::error::Error;

pub const BASE58_ALPHABET: &[u8; 58] =
    b"123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz";

// Crockford Base32 alphabet (case-insensitive, excludes I, L, O)
pub const BASE32_CROCKFORD: &[u8; 32] = b"0123456789ABCDEFGHJKMNPQRSTVWXYZ";
pub const RFC4648_BASE32_LOWER: &[u8; 32] = b"abcdefghijklmnopqrstuvwxyz234567";

const BASE58_DECODE: [i8; 256] = build_decode_table(BASE58_ALPHABET);
const BASE32_CROCKFORD_DECODE: [i8; 256] = build_crockford_decode_table();
const RFC4648_BASE32_DECODE: [i8; 256] = build_rfc4648_base32_decode_table();

const fn build_decode_table(alphabet: &[u8]) -> [i8; 256] {
    let mut out = [-1i8; 256];
    let mut idx = 0usize;
    while idx < alphabet.len() {
        out[alphabet[idx] as usize] = idx as i8;
        idx += 1;
    }
    out
}

const fn build_crockford_decode_table() -> [i8; 256] {
    let mut out = [-1i8; 256];
    let mut idx = 0usize;
    while idx < BASE32_CROCKFORD.len() {
        let upper = BASE32_CROCKFORD[idx];
        out[upper as usize] = idx as i8;
        if upper >= b'A' && upper <= b'Z' {
            out[(upper + 32) as usize] = idx as i8;
        }
        idx += 1;
    }
    out
}

const fn build_rfc4648_base32_decode_table() -> [i8; 256] {
    let mut out = [-1i8; 256];
    let mut idx = 0usize;
    while idx < RFC4648_BASE32_LOWER.len() {
        let lower = RFC4648_BASE32_LOWER[idx];
        out[lower as usize] = idx as i8;
        if lower >= b'a' && lower <= b'z' {
            out[(lower - 32) as usize] = idx as i8;
        }
        idx += 1;
    }
    out
}

pub const CRC8_TABLE: [u8; 256] = [
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
];

#[inline(always)]
#[allow(dead_code)]
pub fn crc8(data: &[u8]) -> u8 {
    let mut crc = 0u8;
    for byte in data {
        crc = CRC8_TABLE[(crc ^ byte) as usize];
    }
    crc
}

#[inline(always)]
pub fn crc8_16(data: &[u8; 16]) -> u8 {
    let mut crc = 0u8;
    crc = CRC8_TABLE[(crc ^ data[0]) as usize];
    crc = CRC8_TABLE[(crc ^ data[1]) as usize];
    crc = CRC8_TABLE[(crc ^ data[2]) as usize];
    crc = CRC8_TABLE[(crc ^ data[3]) as usize];
    crc = CRC8_TABLE[(crc ^ data[4]) as usize];
    crc = CRC8_TABLE[(crc ^ data[5]) as usize];
    crc = CRC8_TABLE[(crc ^ data[6]) as usize];
    crc = CRC8_TABLE[(crc ^ data[7]) as usize];
    crc = CRC8_TABLE[(crc ^ data[8]) as usize];
    crc = CRC8_TABLE[(crc ^ data[9]) as usize];
    crc = CRC8_TABLE[(crc ^ data[10]) as usize];
    crc = CRC8_TABLE[(crc ^ data[11]) as usize];
    crc = CRC8_TABLE[(crc ^ data[12]) as usize];
    crc = CRC8_TABLE[(crc ^ data[13]) as usize];
    crc = CRC8_TABLE[(crc ^ data[14]) as usize];
    CRC8_TABLE[(crc ^ data[15]) as usize]
}

#[inline(always)]
pub fn encode_payload(bytes: [u8; 16]) -> String {
    let mut buf = [0u8; 24];
    encode_payload_to(bytes, &mut buf).to_owned()
}

#[inline(always)]
pub fn encode_payload_to(bytes: [u8; 16], buf: &mut [u8; 24]) -> &str {
    let mut idx = 23usize;
    let checksum = crc8_16(&bytes);
    buf[idx] = BASE58_ALPHABET[(checksum % 58) as usize];
    idx -= 1;

    let mut hi = u64::from_be_bytes(bytes[..8].try_into().unwrap());
    let mut lo = u64::from_be_bytes(bytes[8..].try_into().unwrap());
    while hi > 0 || lo > 0 {
        let qhi = hi / 58;
        let rhi = hi - qhi * 58;
        let (qlo, rem) = div_rem_128_by_58(rhi, lo);
        hi = qhi;
        lo = qlo;
        buf[idx] = BASE58_ALPHABET[rem as usize];
        if idx == 0 {
            break;
        }
        idx -= 1;
    }

    for byte in bytes {
        if byte == 0 {
            buf[idx] = b'1';
            if idx == 0 {
                break;
            }
            idx -= 1;
        } else {
            break;
        }
    }
    // SAFETY: every written byte comes from the ASCII Base58 alphabet.
    unsafe { std::str::from_utf8_unchecked(&buf[idx + 1..]) }
}

#[inline(always)]
pub fn decode_payload(payload: &str) -> Result<[u8; 16], Error> {
    if payload.len() < 2 {
        return Err(Error::InvalidPayload);
    }
    let bytes = payload.as_bytes();
    let (body, checksum_part) = bytes.split_at(bytes.len() - 1);
    let mut decoded = 0u128;
    for &ch in body {
        let digit = BASE58_DECODE[ch as usize];
        if digit < 0 {
            return Err(Error::InvalidPayload);
        }
        decoded = decoded
            .checked_mul(58)
            .and_then(|next| next.checked_add(digit as u128))
            .ok_or(Error::InvalidPayload)?;
    }
    let out = decoded.to_be_bytes();
    let expected = BASE58_ALPHABET[(crc8_16(&out) % 58) as usize];
    if checksum_part.first().copied() != Some(expected) {
        return Err(Error::ChecksumMismatch);
    }
    Ok(out)
}

#[inline(always)]
pub fn decode_base58_value(byte: u8) -> Option<u8> {
    let value = BASE58_DECODE[byte as usize];
    if value < 0 { None } else { Some(value as u8) }
}

pub fn encode_base32(bytes: [u8; 16]) -> String {
    let mut buf = [0u8; 27]; // 26 chars for 128 bits + check digit
    encode_base32_to(bytes, &mut buf).to_owned()
}

pub fn encode_base32_to(bytes: [u8; 16], buf: &mut [u8; 27]) -> &str {
    let mut idx = 26;

    let checksum = crc8_16(&bytes);
    buf[idx] = BASE32_CROCKFORD[(checksum % 32) as usize];
    idx -= 1;

    let mut hi = u64::from_be_bytes(bytes[..8].try_into().unwrap());
    let mut lo = u64::from_be_bytes(bytes[8..].try_into().unwrap());

    while hi > 0 || lo > 0 {
        let qhi = hi >> 5;
        let rhi = hi & 31;
        let qlo = (lo >> 5) | (rhi << 59);
        let rem = (lo & 31) as usize;
        hi = qhi;
        lo = qlo;
        buf[idx] = BASE32_CROCKFORD[rem];
        if idx == 0 {
            break;
        }
        idx -= 1;
    }

    // SAFETY: every written byte comes from the ASCII Crockford Base32 alphabet.
    unsafe { std::str::from_utf8_unchecked(&buf[idx + 1..]) }
}

#[inline(always)]
fn div_rem_128_by_58(hi: u64, lo: u64) -> (u64, u64) {
    debug_assert!(hi < 58);
    let n = (hi << 32) | (lo >> 32);
    let q_hi = n / 58;
    let rem = n - q_hi * 58;

    let n = (rem << 32) | (lo & 0xFFFF_FFFF);
    let q_lo = n / 58;
    let rem = n - q_lo * 58;

    ((q_hi << 32) | q_lo, rem)
}

#[allow(dead_code)]
pub fn decode_base32(payload: &str) -> Result<[u8; 16], Error> {
    // Remove check digit if present (last character)
    let body = if !payload.is_empty() {
        &payload[..payload.len() - 1]
    } else {
        return Err(Error::InvalidPayload);
    };

    let mut value: u128 = 0;

    // Process left-to-right (most significant digit first)
    for ch in body.bytes() {
        let digit = BASE32_CROCKFORD_DECODE[ch as usize];
        if digit < 0 {
            return Err(Error::InvalidPayload);
        }

        value = value.wrapping_mul(32).wrapping_add(digit as u128);
    }

    let mut out = [0u8; 16];
    out[..8].copy_from_slice(&((value >> 64) as u64).to_be_bytes());
    out[8..].copy_from_slice(&(value as u64).to_be_bytes());

    Ok(out)
}

#[allow(dead_code)]
pub fn encode_base32_nopad_lower_to<'a>(bytes: &[u8], out: &'a mut [u8]) -> Result<&'a str, Error> {
    let encoded_len = bytes.len().saturating_mul(8).div_ceil(5);
    if out.len() < encoded_len {
        return Err(Error::InvalidLength);
    }

    let mut buffer = 0u16;
    let mut bits = 0u8;
    let mut cursor = 0usize;
    for byte in bytes {
        buffer = (buffer << 8) | *byte as u16;
        bits += 8;
        while bits >= 5 {
            let shift = bits - 5;
            let idx = ((buffer >> shift) & 0x1F) as usize;
            out[cursor] = RFC4648_BASE32_LOWER[idx];
            cursor += 1;
            bits -= 5;
            buffer &= (1u16 << bits) - 1;
        }
    }
    if bits > 0 {
        let idx = ((buffer << (5 - bits)) & 0x1F) as usize;
        out[cursor] = RFC4648_BASE32_LOWER[idx];
        cursor += 1;
    }

    debug_assert_eq!(cursor, encoded_len);
    // SAFETY: every written byte comes from the lowercase RFC 4648 Base32 alphabet.
    unsafe { Ok(std::str::from_utf8_unchecked(&out[..cursor])) }
}

pub fn encode_base32_32_lower_to<'a>(bytes: &[u8; 32], out: &'a mut [u8; 52]) -> &'a str {
    let mut src = 0usize;
    let mut dst = 0usize;
    for _ in 0..6 {
        encode_base32_5(
            bytes[src..src + 5].try_into().unwrap(),
            &mut out[dst..dst + 8],
        );
        src += 5;
        dst += 8;
    }
    let b0 = bytes[30];
    let b1 = bytes[31];
    out[48] = RFC4648_BASE32_LOWER[(b0 >> 3) as usize];
    out[49] = RFC4648_BASE32_LOWER[(((b0 & 0x07) << 2) | (b1 >> 6)) as usize];
    out[50] = RFC4648_BASE32_LOWER[((b1 >> 1) & 0x1F) as usize];
    out[51] = RFC4648_BASE32_LOWER[((b1 & 0x01) << 4) as usize];
    // SAFETY: every written byte comes from the lowercase RFC 4648 Base32 alphabet.
    unsafe { std::str::from_utf8_unchecked(out) }
}

#[inline(always)]
fn encode_base32_5(input: &[u8; 5], out: &mut [u8]) {
    let b0 = input[0];
    let b1 = input[1];
    let b2 = input[2];
    let b3 = input[3];
    let b4 = input[4];
    out[0] = RFC4648_BASE32_LOWER[(b0 >> 3) as usize];
    out[1] = RFC4648_BASE32_LOWER[(((b0 & 0x07) << 2) | (b1 >> 6)) as usize];
    out[2] = RFC4648_BASE32_LOWER[((b1 >> 1) & 0x1F) as usize];
    out[3] = RFC4648_BASE32_LOWER[(((b1 & 0x01) << 4) | (b2 >> 4)) as usize];
    out[4] = RFC4648_BASE32_LOWER[(((b2 & 0x0F) << 1) | (b3 >> 7)) as usize];
    out[5] = RFC4648_BASE32_LOWER[((b3 >> 2) & 0x1F) as usize];
    out[6] = RFC4648_BASE32_LOWER[(((b3 & 0x03) << 3) | (b4 >> 5)) as usize];
    out[7] = RFC4648_BASE32_LOWER[(b4 & 0x1F) as usize];
}

pub fn decode_base32_32(input: &str) -> Result<[u8; 32], Error> {
    if input.len() != 52 {
        return Err(Error::InvalidContentHash);
    }
    let input = input.as_bytes();
    let mut out = [0u8; 32];
    let mut src = 0usize;
    let mut dst = 0usize;
    for _ in 0..6 {
        decode_base32_8(&input[src..src + 8], &mut out[dst..dst + 5])?;
        src += 8;
        dst += 5;
    }

    let a = decode_rfc4648(input[48])?;
    let b = decode_rfc4648(input[49])?;
    let c = decode_rfc4648(input[50])?;
    let d = decode_rfc4648(input[51])?;
    if d & 0x0F != 0 {
        return Err(Error::InvalidPayload);
    }
    out[30] = (a << 3) | (b >> 2);
    out[31] = ((b & 0x03) << 6) | (c << 1) | (d >> 4);
    Ok(out)
}

#[inline(always)]
fn decode_base32_8(input: &[u8], out: &mut [u8]) -> Result<(), Error> {
    let a = decode_rfc4648(input[0])?;
    let b = decode_rfc4648(input[1])?;
    let c = decode_rfc4648(input[2])?;
    let d = decode_rfc4648(input[3])?;
    let e = decode_rfc4648(input[4])?;
    let f = decode_rfc4648(input[5])?;
    let g = decode_rfc4648(input[6])?;
    let h = decode_rfc4648(input[7])?;
    out[0] = (a << 3) | (b >> 2);
    out[1] = ((b & 0x03) << 6) | (c << 1) | (d >> 4);
    out[2] = ((d & 0x0F) << 4) | (e >> 1);
    out[3] = ((e & 0x01) << 7) | (f << 2) | (g >> 3);
    out[4] = ((g & 0x07) << 5) | h;
    Ok(())
}

#[inline(always)]
fn decode_rfc4648(byte: u8) -> Result<u8, Error> {
    let value = RFC4648_BASE32_DECODE[byte as usize];
    if value < 0 {
        Err(Error::InvalidPayload)
    } else {
        Ok(value as u8)
    }
}

#[allow(dead_code)]
pub fn decode_base32_nopad(input: &str, out: &mut [u8]) -> Result<usize, Error> {
    let expected_len = input.len().saturating_mul(5) / 8;
    if out.len() < expected_len {
        return Err(Error::InvalidLength);
    }

    let mut buffer = 0u16;
    let mut bits = 0u8;
    let mut cursor = 0usize;
    for byte in input.bytes() {
        if byte == b'=' {
            return Err(Error::InvalidPayload);
        }
        let value = decode_rfc4648(byte)?;
        buffer = (buffer << 5) | value as u16;
        bits += 5;
        if bits >= 8 {
            let shift = bits - 8;
            if cursor >= out.len() {
                return Err(Error::InvalidLength);
            }
            out[cursor] = (buffer >> shift) as u8;
            cursor += 1;
            bits -= 8;
            buffer &= (1u16 << bits) - 1;
        }
    }

    if bits > 0 && buffer != 0 {
        return Err(Error::InvalidPayload);
    }
    Ok(cursor)
}

#[test]
fn test_encode_base32() {
    let bytes = [0u8; 16];
    let encoded = encode_base32(bytes);
    assert!(!encoded.is_empty());
    assert!(encoded.len() <= 27); // 26 chars + check digit
}

#[test]
fn test_encode_base32_roundtrip() {
    let bytes = [1u8, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16];
    let encoded = encode_base32(bytes);
    let decoded = decode_base32(&encoded).unwrap();
    assert_eq!(bytes, decoded);
}

#[test]
fn test_decode_base32_invalid() {
    let invalid = "INVALID!@#";
    let result = decode_base32(invalid);
    assert!(result.is_err());
}

#[test]
fn test_decode_base32_case_insensitive() {
    let bytes = [1u8, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16];
    let encoded = encode_base32(bytes);
    let decoded_upper = decode_base32(&encoded).unwrap();
    let decoded_lower = decode_base32(&encoded.to_lowercase()).unwrap();
    assert_eq!(decoded_upper, decoded_lower);
}

#[test]
fn test_rfc4648_base32_nopad_roundtrip() {
    let bytes = [0x42u8; 32];
    let mut encoded = [0u8; 52];
    let encoded = encode_base32_nopad_lower_to(&bytes, &mut encoded).unwrap();
    assert_eq!(encoded.len(), 52);
    let mut decoded = [0u8; 32];
    let len = decode_base32_nopad(encoded, &mut decoded).unwrap();
    assert_eq!(len, 32);
    assert_eq!(decoded, bytes);
}

pub fn encode_fixed64_pair(hi: i64, lo: i64) -> [u8; 16] {
    let mut out = [0u8; 16];
    out[..8].copy_from_slice(&hi.to_be_bytes());
    out[8..].copy_from_slice(&lo.to_be_bytes());
    out
}

pub fn decode_fixed64_pair(raw: &[u8]) -> Result<(i64, i64), Error> {
    if raw.len() != 16 {
        return Err(Error::InvalidLength);
    }
    Ok((
        i64::from_be_bytes(raw[..8].try_into().unwrap()),
        i64::from_be_bytes(raw[8..16].try_into().unwrap()),
    ))
}

pub fn split_wire(value: &str) -> Result<(&str, &str, char), Error> {
    let bytes = value.as_bytes();
    for idx in 0..bytes.len().min(9) {
        if bytes[idx] == b':' || bytes[idx] == b'_' {
            if idx < 2 {
                return Err(Error::InvalidFormat);
            }
            return Ok((&value[..idx], &value[idx + 1..], bytes[idx] as char));
        }
    }
    Err(Error::InvalidFormat)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_crc8() {
        let data = [1u8, 2, 3, 4];
        let crc = crc8(&data);
        assert_ne!(crc, 0);
    }

    #[test]
    fn test_crc8_empty() {
        let data: [u8; 0] = [];
        let crc = crc8(&data);
        assert_eq!(crc, 0);
    }

    #[test]
    fn test_encode_payload() {
        let bytes = [1u8; 16];
        let encoded = encode_payload(bytes);
        assert!(!encoded.is_empty());
        assert!(encoded.len() <= 24);
    }

    #[test]
    fn test_encode_decode_roundtrip() {
        let bytes = [
            0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66,
            0x77, 0x88,
        ];
        let encoded = encode_payload(bytes);
        let decoded = decode_payload(&encoded).unwrap();
        assert_eq!(decoded, bytes);
    }

    #[test]
    fn test_div_rem_128_by_58_matches_u128_reference() {
        let los = [
            0,
            1,
            57,
            58,
            u32::MAX as u64,
            1u64 << 32,
            0x0123_4567_89AB_CDEF,
            u64::MAX,
        ];

        for hi in 0..58 {
            for lo in los {
                let (quot, rem) = div_rem_128_by_58(hi, lo);
                let value = ((hi as u128) << 64) | lo as u128;
                assert_eq!(quot as u128, value / 58, "quot hi={hi} lo={lo}");
                assert_eq!(rem as u128, value % 58, "rem hi={hi} lo={lo}");
            }
        }
    }

    #[test]
    fn test_decode_payload_invalid_length() {
        let result = decode_payload("1");
        assert!(matches!(result, Err(Error::InvalidPayload)));
    }

    #[test]
    fn test_decode_payload_invalid_chars() {
        let result = decode_payload("!!!@#$%^&*()");
        assert!(matches!(result, Err(Error::InvalidPayload)));
    }

    #[test]
    fn test_decode_payload_checksum_mismatch() {
        let bytes = [1u8; 16];
        let mut encoded = encode_payload(bytes);
        // Corrupt the checksum (last character)
        // SAFETY: the encoded Base58 string is ASCII, and replacing one byte
        // with another ASCII byte preserves UTF-8 validity.
        unsafe {
            let bytes = encoded.as_bytes_mut();
            let len = bytes.len();
            if len > 0 {
                bytes[len - 1] = b'1';
            }
        }
        let result = decode_payload(&encoded);
        assert!(matches!(result, Err(Error::ChecksumMismatch)));
    }

    #[test]
    fn test_encode_fixed64_pair() {
        let hi = 0x0123456789ABCDEF;
        let lo = 0xFEDCBA9876543210u64 as i64;
        let encoded = encode_fixed64_pair(hi, lo);
        assert_eq!(encoded.len(), 16);
    }

    #[test]
    fn test_decode_fixed64_pair() {
        let hi = 0x0123456789ABCDEF;
        let lo = 0xFEDCBA9876543210u64 as i64;
        let encoded = encode_fixed64_pair(hi, lo);
        let (decoded_hi, decoded_lo) = decode_fixed64_pair(&encoded).unwrap();
        assert_eq!(decoded_hi, hi);
        assert_eq!(decoded_lo, lo);
    }

    #[test]
    fn test_decode_fixed64_pair_invalid_length() {
        let result = decode_fixed64_pair(&[1u8; 8]);
        assert!(matches!(result, Err(Error::InvalidLength)));
    }

    #[test]
    fn test_split_wire_colon() {
        let wire = "MAT:payload123";
        let (atom, payload, delim) = split_wire(wire).unwrap();
        assert_eq!(atom, "MAT");
        assert_eq!(payload, "payload123");
        assert_eq!(delim, ':');
    }

    #[test]
    fn test_split_wire_underscore() {
        let wire = "MAT_payload123";
        let (atom, payload, delim) = split_wire(wire).unwrap();
        assert_eq!(atom, "MAT");
        assert_eq!(payload, "payload123");
        assert_eq!(delim, '_');
    }

    #[test]
    fn test_split_wire_invalid_format() {
        let result = split_wire("invalid");
        assert!(matches!(result, Err(Error::InvalidFormat)));
    }

    #[test]
    fn test_split_wire_short_atom() {
        let result = split_wire("X:payload");
        assert!(matches!(result, Err(Error::InvalidFormat)));
    }

    #[test]
    fn test_split_wire_no_delimiter() {
        let result = split_wire("MATpayload");
        assert!(matches!(result, Err(Error::InvalidFormat)));
    }

    #[test]
    fn test_base58_alphabet() {
        assert_eq!(BASE58_ALPHABET.len(), 58);
        assert_eq!(BASE58_ALPHABET[0], b'1');
        assert_eq!(BASE58_ALPHABET[57], b'z');
    }

    #[test]
    fn test_crc8_table() {
        assert_eq!(CRC8_TABLE.len(), 256);
    }

    #[test]
    fn test_encode_payload_zero_bytes() {
        let bytes = [0u8; 16];
        let encoded = encode_payload(bytes);
        // Zero bytes should encode to all '1's (base58 for 0)
        assert!(encoded.starts_with('1'));
    }

    #[test]
    fn test_encode_decode_all_zeros() {
        let bytes = [0u8; 16];
        let encoded = encode_payload(bytes);
        let decoded = decode_payload(&encoded).unwrap();
        assert_eq!(decoded, bytes);
    }

    #[test]
    fn test_encode_decode_all_ones() {
        let bytes = [0xFFu8; 16];
        let encoded = encode_payload(bytes);
        let decoded = decode_payload(&encoded).unwrap();
        assert_eq!(decoded, bytes);
    }
}

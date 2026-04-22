//! Base58 encoding, wire format, checksum, and atom handling.

use crate::error::Error;

pub const BASE58_ALPHABET: &[u8; 58] =
    b"123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz";

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

pub fn crc8(data: &[u8]) -> u8 {
    let mut crc = 0u8;
    for byte in data {
        crc = CRC8_TABLE[(crc ^ byte) as usize];
    }
    crc
}

pub fn encode_payload(bytes: [u8; 16]) -> String {
    let mut buf = [0u8; 24];
    let mut idx = 23usize;
    let checksum = crc8(&bytes);
    buf[idx] = BASE58_ALPHABET[(checksum % 58) as usize];
    idx -= 1;

    let mut hi = u64::from_be_bytes(bytes[..8].try_into().unwrap());
    let mut lo = u64::from_be_bytes(bytes[8..].try_into().unwrap());
    while hi > 0 || lo > 0 {
        let qhi = hi / 58;
        let rhi = hi - qhi * 58;
        let dividend = ((rhi as u128) << 64) | lo as u128;
        let qlo = (dividend / 58) as u64;
        let rem = (dividend % 58) as usize;
        hi = qhi;
        lo = qlo;
        buf[idx] = BASE58_ALPHABET[rem];
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
    String::from_utf8(buf[idx + 1..].to_vec()).unwrap()
}

pub fn decode_payload(payload: &str) -> Result<[u8; 16], Error> {
    if payload.len() < 2 {
        return Err(Error::InvalidPayload);
    }
    let (body, checksum_part) = payload.split_at(payload.len() - 1);
    let mut out = [0u8; 16];
    for ch in body.bytes() {
        let value = BASE58_ALPHABET
            .iter()
            .position(|candidate| *candidate == ch)
            .ok_or(Error::InvalidPayload)? as u32;
        let mut carry = value;
        for byte in out.iter_mut().rev() {
            let res = (*byte as u32) * 58 + carry;
            *byte = res as u8;
            carry = res >> 8;
        }
        if carry > 0 {
            return Err(Error::InvalidPayload);
        }
    }
    let expected = BASE58_ALPHABET[(crc8(&out) % 58) as usize] as char;
    if checksum_part.chars().next() != Some(expected) {
        return Err(Error::ChecksumMismatch);
    }
    Ok(out)
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
        let bytes = [0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88];
        let encoded = encode_payload(bytes);
        let decoded = decode_payload(&encoded).unwrap();
        assert_eq!(decoded, bytes);
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

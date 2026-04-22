//! Core SNID type and all Snid methods.

use crate::error::Error;
use crate::encoding::{encode_payload, decode_payload, split_wire};
use crate::helpers::expand_hash_material;
use crate::generator::{init_coarse_clock, GENERATOR};
use std::fmt;

#[derive(Clone, Copy, PartialEq, Eq, Hash)]
pub struct Snid(pub [u8; 16]);

impl fmt::Debug for Snid {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "Snid({})", hex::encode(self.0))
    }
}

impl Snid {
    pub fn new_fast() -> Self {
        init_coarse_clock();
        GENERATOR.with(|cell| unsafe {
            (*cell.get()).next()
        })
    }

    pub fn from_bytes(bytes: [u8; 16]) -> Self {
        Self(bytes)
    }

    pub fn to_bytes(self) -> [u8; 16] {
        self.0
    }

    pub fn from_hex(hex_value: &str) -> Result<Self, Error> {
        let bytes = hex::decode(hex_value)?;
        if bytes.len() != 16 {
            return Err(Error::InvalidLength);
        }
        let mut out = [0u8; 16];
        out.copy_from_slice(&bytes);
        Ok(Self(out))
    }

    pub fn canonical_atom(atom: &str) -> Option<&'static str> {
        match atom {
            "IAM" | "TEN" | "MAT" | "LOC" | "CHR" | "LED" | "LEG" | "TRU" | "KIN" | "COG"
            | "SEM" | "SYS" | "EVT" | "SES" => Some(match atom {
                "IAM" => "IAM",
                "TEN" => "TEN",
                "MAT" => "MAT",
                "LOC" => "LOC",
                "CHR" => "CHR",
                "LED" => "LED",
                "LEG" => "LEG",
                "TRU" => "TRU",
                "KIN" => "KIN",
                "COG" => "COG",
                "SEM" => "SEM",
                "SYS" => "SYS",
                "EVT" => "EVT",
                _ => "SES",
            }),
            "OBJ" => Some("MAT"),
            "TXN" => Some("LED"),
            "SCH" => Some("CHR"),
            "NET" => Some("TRU"),
            "OPS" => Some("EVT"),
            "ACT" => Some("IAM"),
            "GRP" => Some("TEN"),
            "BIO" => Some("IAM"),
            "ATM" => Some("LOC"),
            _ => None,
        }
    }

    pub fn to_wire(self, atom: &str) -> Result<String, Error> {
        let atom = Self::canonical_atom(atom).ok_or(Error::InvalidAtom)?;
        Ok(format!("{atom}:{}", encode_payload(self.0)))
    }

    pub fn parse_wire(value: &str) -> Result<(Self, String), Error> {
        let (atom, payload, delim) = split_wire(value)?;
        if delim == '_' && value.contains('_') {
            // compatibility only; still accepted
        }
        let atom = Self::canonical_atom(atom).ok_or(Error::InvalidAtom)?;
        let bytes = decode_payload(payload)?;
        Ok((Self(bytes), atom.to_string()))
    }

    pub fn timestamp_millis(self) -> i64 {
        let mut hi = [0u8; 8];
        hi.copy_from_slice(&self.0[..8]);
        (u64::from_be_bytes(hi) >> 16) as i64
    }

    pub fn to_tensor_words(self) -> (i64, i64) {
        (
            i64::from_be_bytes(self.0[..8].try_into().unwrap()),
            i64::from_be_bytes(self.0[8..].try_into().unwrap()),
        )
    }

    pub fn sequence(self) -> u16 {
        let hi = (u64::from_be_bytes(self.0[..8].try_into().unwrap()) & 0x0FFF) as u16;
        let lo = ((u64::from_be_bytes(self.0[8..].try_into().unwrap()) >> 60) & 0x03) as u16;
        (hi << 2) | lo
    }

    pub fn machine_or_shard(self) -> u32 {
        ((u64::from_be_bytes(self.0[8..].try_into().unwrap()) >> 36) & 0x00FF_FFFF) as u32
    }

    pub fn time_bin(self, resolution_millis: i64) -> i64 {
        let ts = self.timestamp_millis();
        if resolution_millis <= 1 {
            return ts;
        }
        (ts / resolution_millis) * resolution_millis
    }

    pub fn is_ghosted(self) -> bool {
        (u64::from_be_bytes(self.0[8..].try_into().unwrap()) & (1 << 35)) != 0
    }

    pub fn with_ghost_bit(mut self, enabled: bool) -> Self {
        let mut lo = u64::from_be_bytes(self.0[8..].try_into().unwrap());
        if enabled {
            lo |= 1 << 35;
        } else {
            lo &= !(1 << 35);
        }
        self.0[8..].copy_from_slice(&lo.to_be_bytes());
        self
    }

    pub fn from_hash_with_timestamp(unix_millis: u64, hash: &[u8]) -> Self {
        let material = expand_hash_material(hash);
        let seq = u16::from_be_bytes([material[0], material[1]]) & 0x3FFF;
        let machine =
            ((material[2] as u32) << 16) | ((material[3] as u32) << 8) | material[4] as u32;
        let entropy = (((material[5] as u64) << 32)
            | ((material[6] as u64) << 24)
            | ((material[7] as u64) << 16)
            | ((material[8] as u64) << 8)
            | material[9] as u64)
            & 0xFFFF_FFFFF
            & !(1 << 35);

        let hi = (unix_millis << 16) | 0x7000 | ((seq as u64) >> 2);
        let lo = 0x8000_0000_0000_0000
            | (((seq as u64) & 0x03) << 60)
            | (((machine as u64) & 0x00FF_FFFF) << 36)
            | entropy;
        let mut out = [0u8; 16];
        out[..8].copy_from_slice(&hi.to_be_bytes());
        out[8..].copy_from_slice(&lo.to_be_bytes());
        Self(out)
    }

    pub fn h3_cell(self) -> Option<u64> {
        if self.0[6] >> 4 != 8 || self.0[14] != 0xA5 || (self.0[15] & 0xF0) != 0xA0 {
            return None;
        }
        let mut raw = [0u8; 8];
        raw.copy_from_slice(&self.0[..8]);
        raw[6] = ((self.0[15] & 0x0F) << 4) | (raw[6] & 0x0F);
        Some(u64::from_be_bytes(raw))
    }

    pub fn from_spatial_parts(cell: u64, entropy: u64) -> Self {
        let mut out = [0u8; 16];
        out[..8].copy_from_slice(&cell.to_be_bytes());
        let original_high_nibble = out[6] >> 4;
        out[6] = 0x80 | (out[6] & 0x0F);
        out[8..].copy_from_slice(&entropy.to_be_bytes());
        out[8] = (out[8] & 0x3F) | 0x80;
        out[14] = 0xA5;
        out[15] = 0xA0 | (original_high_nibble & 0x0F);
        Self(out)
    }

    pub fn h3_feature_vector(self) -> Vec<u64> {
        self.h3_cell().map(|cell| vec![cell]).unwrap_or_default()
    }

    #[cfg(feature = "data")]
    pub fn generate_binary_batch(count: usize) -> Vec<u8> {
        let mut out = Vec::with_capacity(count.saturating_mul(16));
        GENERATOR.with(|cell| unsafe {
            let state = &mut *cell.get();
            for _ in 0..count {
                out.extend_from_slice(&state.next().0);
            }
        });
        out
    }

    #[cfg(feature = "data")]
    pub fn generate_tensor_batch(count: usize) -> Vec<i64> {
        let mut out = Vec::with_capacity(count.saturating_mul(2));
        GENERATOR.with(|cell| unsafe {
            let state = &mut *cell.get();
            for _ in 0..count {
                let (hi, lo) = state.next().to_tensor_words();
                out.push(hi);
                out.push(lo);
            }
        });
        out
    }

    #[cfg(feature = "data")]
    pub fn generate_tensor_batch_be_bytes(count: usize) -> Vec<u8> {
        let mut out = Vec::with_capacity(count.saturating_mul(16));
        GENERATOR.with(|cell| unsafe {
            let state = &mut *cell.get();
            for _ in 0..count {
                let (hi, lo) = state.next().to_tensor_words();
                out.extend_from_slice(&hi.to_be_bytes());
                out.extend_from_slice(&lo.to_be_bytes());
            }
        });
        out
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_new_fast() {
        let id = Snid::new_fast();
        assert_ne!(id.0, [0u8; 16]);
    }

    #[test]
    fn test_from_bytes() {
        let bytes = [1u8; 16];
        let id = Snid::from_bytes(bytes);
        assert_eq!(id.0, bytes);
    }

    #[test]
    fn test_to_bytes() {
        let bytes = [2u8; 16];
        let id = Snid::from_bytes(bytes);
        assert_eq!(id.to_bytes(), bytes);
    }

    #[test]
    fn test_from_hex_valid() {
        let hex = "0123456789abcdef0123456789abcdef";
        let id = Snid::from_hex(hex).unwrap();
        assert_eq!(hex::encode(id.0), hex);
    }

    #[test]
    fn test_from_hex_invalid_length() {
        let result = Snid::from_hex("0123");
        assert!(matches!(result, Err(Error::InvalidLength)));
    }

    #[test]
    fn test_from_hex_invalid_chars() {
        let result = Snid::from_hex("gggggggggggggggggggggggggggggggg");
        assert!(result.is_err());
    }

    #[test]
    fn test_canonical_atom_valid() {
        assert_eq!(Snid::canonical_atom("MAT"), Some("MAT"));
        assert_eq!(Snid::canonical_atom("IAM"), Some("IAM"));
        assert_eq!(Snid::canonical_atom("TEN"), Some("TEN"));
    }

    #[test]
    fn test_canonical_atom_legacy() {
        assert_eq!(Snid::canonical_atom("OBJ"), Some("MAT"));
        assert_eq!(Snid::canonical_atom("TXN"), Some("LED"));
        assert_eq!(Snid::canonical_atom("SCH"), Some("CHR"));
    }

    #[test]
    fn test_canonical_atom_invalid() {
        assert_eq!(Snid::canonical_atom("XXX"), None);
        assert_eq!(Snid::canonical_atom(""), None);
    }

    #[test]
    fn test_to_wire() {
        let id = Snid::from_bytes([1u8; 16]);
        let wire = id.to_wire("MAT").unwrap();
        assert!(wire.starts_with("MAT:"));
    }

    #[test]
    fn test_to_wire_invalid_atom() {
        let id = Snid::from_bytes([1u8; 16]);
        let result = id.to_wire("XXX");
        assert!(matches!(result, Err(Error::InvalidAtom)));
    }

    #[test]
    fn test_parse_wire_valid() {
        let id = Snid::from_bytes([1u8; 16]);
        let wire = id.to_wire("MAT").unwrap();
        let (parsed, atom) = Snid::parse_wire(&wire).unwrap();
        assert_eq!(parsed, id);
        assert_eq!(atom, "MAT");
    }

    #[test]
    fn test_parse_wire_underscore() {
        let id = Snid::from_bytes([1u8; 16]);
        let wire = id.to_wire("MAT").unwrap();
        let underscore_wire = wire.replace(':', "_");
        let (parsed, _atom) = Snid::parse_wire(&underscore_wire).unwrap();
        assert_eq!(parsed, id);
    }

    #[test]
    fn test_parse_wire_invalid_format() {
        let result = Snid::parse_wire("invalid");
        assert!(matches!(result, Err(Error::InvalidFormat)));
    }

    #[test]
    fn test_parse_wire_invalid_atom() {
        let result = Snid::parse_wire("XXX:payload");
        assert!(matches!(result, Err(Error::InvalidAtom)));
    }

    #[test]
    fn test_timestamp_millis() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        assert_eq!(id.timestamp_millis(), 1700000000123);
    }

    #[test]
    fn test_to_tensor_words() {
        let id = Snid::from_bytes([1u8; 16]);
        let (hi, lo) = id.to_tensor_words();
        assert_ne!(hi, 0);
        assert_ne!(lo, 0);
    }

    #[test]
    fn test_sequence() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let seq = id.sequence();
        assert!(seq <= 0x3FFF);
    }

    #[test]
    fn test_machine_or_shard() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let machine = id.machine_or_shard();
        assert!(machine <= 0x00FF_FFFF);
    }

    #[test]
    fn test_time_bin() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let bin = id.time_bin(3600000); // 1 hour
        assert!(bin > 0);
    }

    #[test]
    fn test_time_bin_zero_resolution() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let bin = id.time_bin(0);
        assert_eq!(bin, 1700000000123);
    }

    #[test]
    fn test_is_ghosted_default() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        assert!(!id.is_ghosted());
    }

    #[test]
    fn test_with_ghost_bit_enable() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let ghosted = id.with_ghost_bit(true);
        assert!(ghosted.is_ghosted());
    }

    #[test]
    fn test_with_ghost_bit_disable() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let ghosted = id.with_ghost_bit(true);
        let unghosted = ghosted.with_ghost_bit(false);
        assert!(!unghosted.is_ghosted());
    }

    #[test]
    fn test_from_hash_with_timestamp() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        assert_eq!(id.timestamp_millis(), 1700000000123);
        assert_ne!(id.0, [0u8; 16]);
    }

    #[test]
    fn test_from_hash_with_timestamp_empty_hash() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"");
        assert_eq!(id.timestamp_millis(), 1700000000123);
    }

    #[test]
    fn test_h3_cell_non_spatial() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        assert_eq!(id.h3_cell(), None);
    }

    #[test]
    fn test_from_spatial_parts() {
        let cell = 0x8c2a1072b59ffff;
        let entropy = 0x1234567890ABCDEF;
        let id = Snid::from_spatial_parts(cell, entropy);
        assert_eq!(id.h3_cell(), Some(cell));
    }

    #[test]
    fn test_h3_feature_vector_non_spatial() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        assert!(id.h3_feature_vector().is_empty());
    }

    #[test]
    fn test_h3_feature_vector_spatial() {
        let cell = 0x8c2a1072b59ffff;
        let entropy = 0x1234567890ABCDEF;
        let id = Snid::from_spatial_parts(cell, entropy);
        let features = id.h3_feature_vector();
        assert_eq!(features, vec![cell]);
    }

    #[cfg(feature = "data")]
    #[test]
    fn test_generate_binary_batch() {
        let batch = Snid::generate_binary_batch(10);
        assert_eq!(batch.len(), 160);
    }

    #[cfg(feature = "data")]
    #[test]
    fn test_generate_tensor_batch() {
        let batch = Snid::generate_tensor_batch(10);
        assert_eq!(batch.len(), 20);
    }

    #[cfg(feature = "data")]
    #[test]
    fn test_generate_tensor_batch_be_bytes() {
        let batch = Snid::generate_tensor_batch_be_bytes(10);
        assert_eq!(batch.len(), 160);
    }

    #[test]
    fn test_roundtrip_bytes() {
        let id1 = Snid::new_fast();
        let bytes = id1.to_bytes();
        let id2 = Snid::from_bytes(bytes);
        assert_eq!(id1, id2);
    }

    #[test]
    fn test_roundtrip_wire() {
        let id1 = Snid::new_fast();
        let wire = id1.to_wire("MAT").unwrap();
        let (id2, atom) = Snid::parse_wire(&wire).unwrap();
        assert_eq!(id1, id2);
        assert_eq!(atom, "MAT");
    }

    #[test]
    fn test_version_bits() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let version = (id.0[6] >> 4) & 0x0F;
        assert_eq!(version, 0x7);
    }

    #[test]
    fn test_variant_bits() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let variant = (id.0[8] >> 6) & 0b11;
        assert_eq!(variant, 0b10);
    }
}

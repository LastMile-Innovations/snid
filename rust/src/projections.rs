//! LLM/tensor projections, time binning, and ghost bit.

use crate::core::Snid;
use crate::error::Error;

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct LlmFormatV1 {
    pub atom: String,
    pub timestamp_millis: i64,
    pub machine_or_shard: u32,
    pub sequence: u16,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct LlmFormatV2 {
    pub kind: String,
    pub atom: String,
    pub timestamp_millis: Option<i64>,
    pub spatial_anchor: Option<u64>,
    pub machine_or_shard: Option<u32>,
    pub sequence: Option<u16>,
    pub ghosted: bool,
}

impl Snid {
    pub fn to_llm_format(self, atom: &str) -> Result<LlmFormatV1, Error> {
        let atom = Self::canonical_atom(atom).ok_or(Error::InvalidAtom)?;
        Ok(LlmFormatV1 {
            atom: atom.to_string(),
            timestamp_millis: self.timestamp_millis(),
            machine_or_shard: self.machine_or_shard(),
            sequence: self.sequence(),
        })
    }

    pub fn to_llm_format_v2(self, atom: &str) -> Result<LlmFormatV2, Error> {
        let atom = Self::canonical_atom(atom).ok_or(Error::InvalidAtom)?;
        if let Some(cell) = self.h3_cell() {
            return Ok(LlmFormatV2 {
                kind: "sgid".to_string(),
                atom: atom.to_string(),
                timestamp_millis: None,
                spatial_anchor: Some(cell),
                machine_or_shard: None,
                sequence: None,
                ghosted: self.is_ghosted(),
            });
        }
        Ok(LlmFormatV2 {
            kind: "snid".to_string(),
            atom: atom.to_string(),
            timestamp_millis: Some(self.timestamp_millis()),
            spatial_anchor: None,
            machine_or_shard: Some(self.machine_or_shard()),
            sequence: Some(self.sequence()),
            ghosted: self.is_ghosted(),
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_to_llm_format_valid() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let llm = id.to_llm_format("MAT").unwrap();
        assert_eq!(llm.atom, "MAT");
        assert_eq!(llm.timestamp_millis, 1700000000123);
        assert!(llm.machine_or_shard <= 0x00FF_FFFF);
        assert!(llm.sequence <= 0x3FFF);
    }

    #[test]
    fn test_to_llm_format_invalid_atom() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let result = id.to_llm_format("XXX");
        assert!(matches!(result, Err(Error::InvalidAtom)));
    }

    #[test]
    fn test_to_llm_format_v2_snid() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let llm = id.to_llm_format_v2("MAT").unwrap();
        assert_eq!(llm.kind, "snid");
        assert_eq!(llm.atom, "MAT");
        assert_eq!(llm.timestamp_millis, Some(1700000000123));
        assert_eq!(llm.spatial_anchor, None);
        assert!(llm.machine_or_shard.is_some());
        assert!(llm.sequence.is_some());
        assert!(!llm.ghosted);
    }

    #[test]
    fn test_to_llm_format_v2_sgid() {
        let cell = 0x8c2a1072b59ffff;
        let entropy = 0x1234567890ABCDEF;
        let id = Snid::from_spatial_parts(cell, entropy);
        let llm = id.to_llm_format_v2("LOC").unwrap();
        assert_eq!(llm.kind, "sgid");
        assert_eq!(llm.atom, "LOC");
        assert_eq!(llm.timestamp_millis, None);
        assert_eq!(llm.spatial_anchor, Some(cell));
        assert_eq!(llm.machine_or_shard, None);
        assert_eq!(llm.sequence, None);
    }

    #[test]
    fn test_to_llm_format_v2_invalid_atom() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let result = id.to_llm_format_v2("XXX");
        assert!(matches!(result, Err(Error::InvalidAtom)));
    }

    #[test]
    fn test_to_llm_format_v2_ghosted() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let ghosted = id.with_ghost_bit(true);
        let llm = ghosted.to_llm_format_v2("MAT").unwrap();
        assert!(llm.ghosted);
    }

    #[test]
    fn test_llm_format_v1_equality() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let llm1 = id.to_llm_format("MAT").unwrap();
        let llm2 = id.to_llm_format("MAT").unwrap();
        assert_eq!(llm1, llm2);
    }

    #[test]
    fn test_llm_format_v2_equality() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let llm1 = id.to_llm_format_v2("MAT").unwrap();
        let llm2 = id.to_llm_format_v2("MAT").unwrap();
        assert_eq!(llm1, llm2);
    }
}

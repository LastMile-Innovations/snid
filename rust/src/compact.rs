//! Compact identifiers: ShortId, TraceId.

use crate::core::Snid;
use crate::encoding::encode_payload;
use crate::error::Error;
use crate::helpers::hex_encode_fast;
use getrandom::getrandom;
use std::fmt;
use std::str::FromStr;

#[derive(Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Hash, Debug)]
#[repr(transparent)]
pub struct ShortId(pub [u8; 8]);

impl ShortId {
    pub fn new() -> Self {
        let mut raw = [0u8; 8];
        getrandom(&mut raw).unwrap();
        Self(raw)
    }

    pub fn from_bytes(bytes: [u8; 8]) -> Self {
        Self(bytes)
    }

    pub fn to_string(&self, atom: &str) -> String {
        let mut padded = [0u8; 16];
        padded[..8].copy_from_slice(&self.0);
        format!("{}:{}", atom, encode_payload(padded))
    }
}

impl FromStr for ShortId {
    type Err = Error;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        let (id, _) = Snid::parse_wire(s)?;
        let mut out = [0u8; 8];
        out.copy_from_slice(&id.0[..8]);
        Ok(Self(out))
    }
}

impl fmt::Display for ShortId {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        let mut padded = [0u8; 16];
        padded[..8].copy_from_slice(&self.0);
        let payload = encode_payload(padded);
        write!(f, "MAT:{}", payload)
    }
}

#[derive(Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Hash, Debug)]
#[repr(transparent)]
pub struct TraceId(pub [u8; 16]);

impl TraceId {
    pub fn new() -> Self {
        Self(Snid::new_fast().0)
    }

    pub fn from_bytes(bytes: [u8; 16]) -> Self {
        Self(bytes)
    }

    pub fn to_hex(&self) -> String {
        hex_encode_fast(&self.0)
    }

    pub fn to_trace_parent(&self, span_id: [u8; 8]) -> String {
        let trace_hex = self.to_hex();
        let span_hex = hex_encode_fast(&span_id);
        format!("00-{}-{}-01", trace_hex, span_hex)
    }
}

impl FromStr for TraceId {
    type Err = Error;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        let id = Snid::from_hex(s)?;
        Ok(Self(id.0))
    }
}

impl fmt::Display for TraceId {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", hex_encode_fast(&self.0))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_short_id_from_str() {
        let id1 = ShortId::new();
        let wire = id1.to_string("MAT");
        let id2: ShortId = wire.parse().unwrap();
        // ShortId only stores first 8 bytes, so we compare those
        assert_eq!(id1.0, id2.0);
    }

    #[test]
    fn test_short_id_display() {
        let id = ShortId::new();
        let display = format!("{}", id);
        assert!(display.starts_with("MAT:"));
    }

    #[test]
    fn test_trace_id_from_str() {
        let id1 = TraceId::new();
        let hex = id1.to_hex();
        let id2: TraceId = hex.parse().unwrap();
        assert_eq!(id1.0, id2.0);
    }

    #[test]
    fn test_trace_id_display() {
        let id = TraceId::new();
        let display = format!("{}", id);
        assert_eq!(display.len(), 32); // 16 bytes = 32 hex chars
    }

    #[test]
    fn test_short_id_ord() {
        let id1 = ShortId::from_bytes([1u8; 8]);
        let id2 = ShortId::from_bytes([2u8; 8]);
        assert!(id1 < id2);
    }

    #[test]
    fn test_trace_id_ord() {
        let id1 = TraceId::from_bytes([1u8; 16]);
        let id2 = TraceId::from_bytes([2u8; 16]);
        assert!(id1 < id2);
    }
}

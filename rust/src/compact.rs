//! Compact identifiers: ShortId, TraceId.

use crate::core::Snid;
use crate::encoding::encode_payload;
use crate::error::Error;
use crate::helpers::{hex_encode_fast, hex_encode_to};
use getrandom::fill;
use std::fmt;
use std::str::FromStr;

#[derive(Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Hash, Debug)]
#[repr(transparent)]
pub struct ShortId(pub [u8; 8]);

impl ShortId {
    pub fn new() -> Self {
        Self::try_new().expect("os rng")
    }

    pub fn try_new() -> Result<Self, Error> {
        let mut raw = [0u8; 8];
        fill(&mut raw)?;
        Ok(Self(raw))
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

impl Default for ShortId {
    fn default() -> Self {
        Self::new()
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
        let mut out = [0u8; 55];
        self.write_traceparent(span_id, &mut out).to_owned()
    }

    pub fn write_traceparent<'a>(&self, span_id: [u8; 8], out: &'a mut [u8; 55]) -> &'a str {
        out[..3].copy_from_slice(b"00-");
        hex_encode_to(&self.0, &mut out[3..35]);
        out[35] = b'-';
        hex_encode_to(&span_id, &mut out[36..52]);
        out[52..].copy_from_slice(b"-01");
        unsafe { std::str::from_utf8_unchecked(out) }
    }
}

impl Default for TraceId {
    fn default() -> Self {
        Self::new()
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
    fn test_trace_id_write_traceparent_matches_to_trace_parent() {
        let id = TraceId::from_bytes([1u8; 16]);
        let span = [2u8; 8];
        let mut out = [0u8; 55];
        let written = id.write_traceparent(span, &mut out);
        assert_eq!(written, id.to_trace_parent(span));
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

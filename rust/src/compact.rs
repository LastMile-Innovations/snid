//! Compact identifiers: ShortId, TraceId.

use crate::core::Snid;
use crate::encoding::encode_payload;
use crate::helpers::hex_encode_fast;
use getrandom::getrandom;

#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub struct ShortId(pub [u8; 8]);

impl ShortId {
    pub fn new() -> Self {
        let mut raw = [0u8; 8];
        getrandom(&mut raw).unwrap();
        Self(raw)
    }

    pub fn to_string(&self, atom: &str) -> String {
        let mut padded = [0u8; 16];
        padded[..8].copy_from_slice(&self.0);
        format!("{}:{}", atom, encode_payload(padded))
    }
}

#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub struct TraceId(pub [u8; 16]);

impl TraceId {
    pub fn new() -> Self {
        Self(Snid::new_fast().0)
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

//! Access Key ID (AKID) dual-part credentials.

use crate::core::Snid;
use crate::encoding::{decode_base58_value, encode_payload_to};
use crate::error::Error;
use crate::helpers::fnv1a;

#[derive(Clone, Debug)]
pub struct Akid {
    pub public: Snid,
    pub secret: String,
}

impl Akid {
    pub fn new_public() -> Self {
        Self {
            public: Snid::new_fast(),
            secret: String::new(),
        }
    }

    pub fn encode_secret(&self) -> Result<String, Error> {
        if self.secret.is_empty() {
            return Ok(String::new());
        }
        let hash = fnv1a(&self.secret);
        let checksum = (hash % 58) as u8;
        let alphabet = b"123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz";
        let mut out = String::new();
        for byte in self.secret.as_bytes() {
            out.push(alphabet[(byte % 58) as usize] as char);
        }
        out.push(alphabet[checksum as usize] as char);
        Ok(out)
    }

    pub fn new_secret(secret: &str) -> Self {
        Self {
            public: Snid::new_fast(),
            secret: secret.to_string(),
        }
    }

    pub fn verify_secret_checksum(encoded: &str) -> Result<(), Error> {
        if encoded.is_empty() {
            return Ok(());
        }
        let bytes = encoded.as_bytes();
        let checksum_byte = bytes[bytes.len() - 1];
        let checksum_pos = decode_base58_value(checksum_byte).ok_or(Error::InvalidPayload)?;
        let mut hash: u32 = 2166136261;
        for &byte in &bytes[..bytes.len() - 1] {
            let pos = decode_base58_value(byte).ok_or(Error::InvalidPayload)? as u32;
            hash ^= pos;
            hash = hash.wrapping_mul(16777619);
        }
        if (hash % 58) as u8 != checksum_pos {
            return Err(Error::ChecksumMismatch);
        }
        Ok(())
    }

    pub fn parse(wire: &str) -> Result<(Self, String), Error> {
        let wire = wire.trim();
        if !wire.starts_with("KEY:") {
            return Err(Error::InvalidFormat);
        }
        let (public_part, secret_part) = wire.split_once('_').ok_or(Error::InvalidFormat)?;
        let payload = public_part
            .strip_prefix("KEY:")
            .ok_or(Error::InvalidFormat)?;
        let public = Snid(crate::encoding::decode_payload(payload)?);
        Self::verify_secret_checksum(secret_part)?;
        Ok((
            Self {
                public,
                secret: secret_part.to_string(),
            },
            "KEY".to_string(),
        ))
    }

    pub fn format(public: Snid, secret: &str) -> String {
        let mut out = String::with_capacity(4 + 24 + 1 + secret.trim().len());
        let mut payload = [0u8; 24];
        out.push_str("KEY:");
        out.push_str(encode_payload_to(public.0, &mut payload));
        out.push('_');
        out.push_str(secret.trim());
        out
    }
}

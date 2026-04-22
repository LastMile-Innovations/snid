//! Access Key ID (AKID) dual-part credentials.

use crate::core::Snid;
use crate::encoding::decode_base58_value;
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
        let parts: Vec<&str> = wire.splitn(2, '_').collect();
        if parts.len() != 2 {
            return Err(Error::InvalidFormat);
        }
        let (public, atom) = Snid::parse_wire(parts[0])?;
        if atom != "KEY" {
            return Err(Error::InvalidAtom);
        }
        Self::verify_secret_checksum(parts[1])?;
        Ok((
            Self {
                public,
                secret: parts[1].to_string(),
            },
            atom,
        ))
    }

    pub fn format(public: Snid, secret: &str) -> String {
        format!(
            "KEY:{}_{}",
            public
                .to_wire("MAT")
                .unwrap()
                .split(':')
                .nth(1)
                .unwrap_or(""),
            secret.trim()
        )
    }
}

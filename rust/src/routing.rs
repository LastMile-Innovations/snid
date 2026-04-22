//! Routing types: GrantId, ScopeId, ShardId, AliasId.

use crate::core::Snid;
use crate::encoding::{decode_payload, encode_payload};
use crate::error::Error;
use crate::helpers::fnv1a;
use std::time::{SystemTime, UNIX_EPOCH};

#[derive(Clone, Debug)]
pub struct GrantId {
    pub id: Snid,
    pub atom: String,
    pub signature: [u8; 16],
    pub expires_at: Option<SystemTime>,
}

impl GrantId {
    pub fn new(atom: &str, ttl: Option<std::time::Duration>, _secret: &[u8]) -> Result<Self, Error> {
        let atom = Snid::canonical_atom(atom).ok_or(Error::InvalidAtom)?;
        let id = Snid::new_fast();
        let expires_at = ttl.map(|d| SystemTime::now() + d);

        // TODO: Implement AES-CBC-MAC signing
        let signature = [0u8; 16];

        Ok(Self {
            id,
            atom: atom.to_string(),
            signature,
            expires_at,
        })
    }

    pub fn verify(&self, _secret: &[u8]) -> bool {
        // TODO: Implement AES-CBC-MAC verification
        if let Some(exp) = self.expires_at {
            if SystemTime::now() > exp {
                return false;
            }
        }
        true
    }

    pub fn to_string(&self, atom: &str) -> String {
        let use_atom = if self.atom.is_empty() {
            Snid::canonical_atom(atom).unwrap_or(atom)
        } else {
            Snid::canonical_atom(&self.atom).unwrap_or(&self.atom)
        };

        let mut buf = String::new();
        buf.push_str(&self.id.to_wire(use_atom).unwrap());
        if let Some(exp) = self.expires_at {
            if let Ok(dur) = exp.duration_since(UNIX_EPOCH) {
                buf.push('@');
                buf.push_str(&dur.as_secs().to_string());
            }
        }
        buf.push('.');
        buf.push_str(&encode_payload(self.signature));
        buf
    }

    pub fn parse(s: &str, _secret: &[u8]) -> Result<(Self, String), Error> {
        let dot_idx = s.rfind('.').ok_or(Error::InvalidFormat)?;
        let sig_part = &s[dot_idx + 1..];
        let main_part = &s[..dot_idx];

        let mut exp = None;
        let mut id_part = main_part;

        if let Some(at_idx) = main_part.rfind('@') {
            id_part = &main_part[..at_idx];
            let ts: u64 = main_part[at_idx + 1..].parse().map_err(|_| Error::InvalidFormat)?;
            exp = Some(UNIX_EPOCH + std::time::Duration::from_secs(ts));
        }

        let (id, atom) = Snid::parse_wire(id_part)?;
        let signature = decode_payload(sig_part)?;

        let grant = Self {
            id,
            atom: atom.clone(),
            signature: signature.try_into().map_err(|_| Error::InvalidLength)?,
            expires_at: exp,
        };

        Ok((grant, atom))
    }
}

#[derive(Clone, Debug)]
pub struct ScopeId {
    pub id: Snid,
    pub scope: String,
}

impl ScopeId {
    pub fn new(_atom: &str, scope: &str) -> Self {
        let mut id = Snid::new_fast();
        let hash = fnv1a(scope);
        let hash_bytes = hash.to_be_bytes();
        id.0[10..14].copy_from_slice(&hash_bytes[..4]);
        Self {
            id,
            scope: scope.to_string(),
        }
    }

    pub fn new_with_hash(_atom: &str, scope: &str, hash: u32) -> Self {
        let mut id = Snid::new_fast();
        let hash_bytes = hash.to_be_bytes();
        id.0[10..14].copy_from_slice(&hash_bytes[..4]);
        Self {
            id,
            scope: scope.to_string(),
        }
    }

    pub fn hash_scope(s: &str) -> u32 {
        fnv1a(s)
    }

    pub fn to_string(&self, atom: &str) -> String {
        if self.scope.is_empty() {
            return self.id.to_wire(atom).unwrap();
        }
        format!("{}:{}.{}", atom, self.scope, encode_payload(self.id.0))
    }

    pub fn parse(s: &str) -> Result<(Self, String), Error> {
        let delim_idx = s.find(':').ok_or(Error::InvalidFormat)?;
        let dot_idx = s.rfind('.').ok_or(Error::InvalidFormat)?;

        if delim_idx >= dot_idx {
            let (id, atom) = Snid::parse_wire(s)?;
            return Ok((Self { id, scope: String::new() }, atom));
        }

        let atom = &s[..delim_idx];
        let scope = &s[delim_idx + 1..dot_idx];

        let id = Snid::from_hex(&encode_payload(
            decode_payload(&s[dot_idx + 1..])?
        ))?;

        Ok((Self {
            id,
            scope: scope.to_string(),
        }, atom.to_string()))
    }
}

#[derive(Clone, Debug)]
pub struct ShardId {
    pub id: Snid,
    pub shard_key: u16,
}

impl ShardId {
    pub fn new(_atom: &str, shard: u16) -> Self {
        let id = Snid::new_fast();
        Self {
            id,
            shard_key: shard,
        }
    }

    pub fn shard(&self, total: usize) -> usize {
        if total == 0 {
            return 0;
        }
        (self.shard_key as usize) % total
    }

    pub fn to_string(&self, atom: &str) -> String {
        format!("{}:{}#{}", atom, encode_payload(self.id.0), self.shard_key)
    }

    pub fn parse(s: &str) -> Result<(Self, String), Error> {
        let idx = s.rfind('#').ok_or(Error::InvalidFormat)?;
        let (id, atom) = Snid::parse_wire(&s[..idx])?;
        let shard_key: u16 = s[idx + 1..].parse().map_err(|_| Error::InvalidFormat)?;
        Ok((Self { id, shard_key }, atom))
    }
}

#[derive(Clone, Debug)]
pub struct AliasId {
    pub id: Snid,
    pub alias: String,
}

impl AliasId {
    pub fn new(_atom: &str, alias: &str) -> Self {
        Self {
            id: Snid::new_fast(),
            alias: crate::helpers::sanitize_alias(alias),
        }
    }

    pub fn to_string(&self, atom: &str) -> String {
        format!("{}:{}/{}", atom, self.alias, encode_payload(self.id.0))
    }

    pub fn parse(s: &str) -> Result<(Self, String), Error> {
        let colon_idx = s.find(':').ok_or(Error::InvalidFormat)?;
        let slash_idx = s.rfind('/').ok_or(Error::InvalidFormat)?;

        if colon_idx >= slash_idx {
            let (id, atom) = Snid::parse_wire(s)?;
            return Ok((Self { id, alias: String::new() }, atom));
        }

        let atom = &s[..colon_idx];
        let alias = &s[colon_idx + 1..slash_idx];
        let id = Snid::from_hex(&encode_payload(
            decode_payload(&s[slash_idx + 1..])?
        ))?;

        Ok((Self {
            id,
            alias: alias.to_string(),
        }, atom.to_string()))
    }
}

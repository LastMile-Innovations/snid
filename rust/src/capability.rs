//! Capability ID (KID) with MAC-based verification.

use crate::core::Snid;
use crate::error::Error;
use hmac::{Hmac, Mac};
use sha2::Sha256;

type HmacSha256 = Hmac<Sha256>;

#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub struct Kid(pub [u8; 32]);

impl Kid {
    pub fn from_parts(
        head: Snid,
        actor: Snid,
        resource: &[u8],
        capability: &[u8],
        key: &[u8],
    ) -> Result<Self, Error> {
        if key.is_empty() {
            return Err(Error::InvalidKey);
        }
        let mut mac = <HmacSha256 as hmac::Mac>::new_from_slice(key).map_err(|_| Error::InvalidKey)?;
        mac.update(&head.0);
        mac.update(&actor.0);
        mac.update(resource);
        mac.update(capability);
        let sum = mac.finalize().into_bytes();
        let mut out = [0u8; 32];
        out[..16].copy_from_slice(&head.0);
        out[16..].copy_from_slice(&sum[..16]);
        Ok(Self(out))
    }

    pub fn head(&self) -> Snid {
        let mut out = [0u8; 16];
        out.copy_from_slice(&self.0[..16]);
        Snid(out)
    }

    pub fn verify(&self, actor: Snid, resource: &[u8], capability: &[u8], key: &[u8]) -> bool {
        Self::from_parts(self.head(), actor, resource, capability, key)
            .map(|expected| expected == *self)
            .unwrap_or(false)
    }

    pub fn to_tensor256_words(&self) -> (i64, i64, i64, i64) {
        (
            i64::from_be_bytes(self.0[0..8].try_into().unwrap()),
            i64::from_be_bytes(self.0[8..16].try_into().unwrap()),
            i64::from_be_bytes(self.0[16..24].try_into().unwrap()),
            i64::from_be_bytes(self.0[24..32].try_into().unwrap()),
        )
    }
}

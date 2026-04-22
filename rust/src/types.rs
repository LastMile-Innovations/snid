//! Extended identifier families: Nid, Lid, Wid, Xid, Kid, Eid, Bid.

use crate::core::Snid;
use crate::encoding::{decode_base32_32, encode_base32_32_lower_to, encode_payload_to};
use crate::error::Error;
use hmac::{Hmac, Mac};
use sha2::Sha256;
use subtle::ConstantTimeEq;

type HmacSha256 = Hmac<Sha256>;

#[derive(Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Hash, Debug)]
#[repr(transparent)]
pub struct Nid(pub [u8; 32]);

impl Nid {
    #[inline(always)]
    pub fn from_parts(head: Snid, semantic_hash: [u8; 16]) -> Self {
        let mut out = [0u8; 32];
        out[..16].copy_from_slice(&head.0);
        out[16..].copy_from_slice(&semantic_hash);
        Self(out)
    }

    #[inline(always)]
    pub fn head(&self) -> Snid {
        let mut out = [0u8; 16];
        out.copy_from_slice(&self.0[..16]);
        Snid(out)
    }

    #[inline(always)]
    pub fn semantic_hash(&self) -> [u8; 16] {
        let mut out = [0u8; 16];
        out.copy_from_slice(&self.0[16..]);
        out
    }

    #[inline(always)]
    pub fn quantize(vector: &[f32]) -> [u8; 16] {
        let mut out = [0u8; 16];
        for (i, value) in vector.iter().take(128).enumerate() {
            if *value > 0.0 {
                out[i / 8] |= 1 << (7 - (i % 8));
            }
        }
        out
    }

    #[inline(always)]
    pub fn hamming_distance(&self, other: &Nid) -> u32 {
        // Direct byte comparison without allocations for performance
        let mut distance = 0u32;
        for i in 16..32 {
            distance += (self.0[i] ^ other.0[i]).count_ones();
        }
        distance
    }

    #[inline(always)]
    pub fn to_tensor256_words(self) -> (i64, i64, i64, i64) {
        (
            i64::from_be_bytes(self.0[0..8].try_into().unwrap()),
            i64::from_be_bytes(self.0[8..16].try_into().unwrap()),
            i64::from_be_bytes(self.0[16..24].try_into().unwrap()),
            i64::from_be_bytes(self.0[24..32].try_into().unwrap()),
        )
    }

    /// Generate a batch of NIDs from a single head with varying semantic hashes
    /// Reduces allocations by pre-allocating the output vector
    pub fn batch_from_head(head: Snid, semantic_hashes: &[[u8; 16]]) -> Vec<Self> {
        let mut out = Vec::with_capacity(semantic_hashes.len());
        for semantic_hash in semantic_hashes {
            out.push(Self::from_parts(head, *semantic_hash));
        }
        out
    }
}

#[derive(Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Hash, Debug)]
#[repr(transparent)]
pub struct Lid(pub [u8; 32]);

impl Lid {
    pub fn from_parts(
        head: Snid,
        prev: [u8; 32],
        payload: &[u8],
        key: &[u8],
    ) -> Result<Self, Error> {
        if key.is_empty() {
            return Err(Error::InvalidKey);
        }
        let mut mac =
            <HmacSha256 as hmac::Mac>::new_from_slice(key).map_err(|_| Error::InvalidKey)?;
        mac.update(&head.0);
        mac.update(&prev);
        mac.update(payload);
        let mut out = [0u8; 32];
        out[..16].copy_from_slice(&head.0);
        out[16..].copy_from_slice(&mac.finalize().into_bytes()[..16]);
        Ok(Self(out))
    }

    pub fn head(&self) -> Snid {
        let mut head = [0u8; 16];
        head.copy_from_slice(&self.0[..16]);
        Snid(head)
    }

    pub fn to_tensor256_words(self) -> (i64, i64, i64, i64) {
        (
            i64::from_be_bytes(self.0[0..8].try_into().unwrap()),
            i64::from_be_bytes(self.0[8..16].try_into().unwrap()),
            i64::from_be_bytes(self.0[16..24].try_into().unwrap()),
            i64::from_be_bytes(self.0[24..32].try_into().unwrap()),
        )
    }
}

#[derive(Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Hash, Debug)]
#[repr(transparent)]
pub struct Wid(pub [u8; 32]);

impl Wid {
    pub fn from_parts(head: Snid, scenario_hash: [u8; 16]) -> Self {
        let mut out = [0u8; 32];
        out[..16].copy_from_slice(&head.0);
        out[16..].copy_from_slice(&scenario_hash);
        Self(out)
    }

    pub fn head(&self) -> Snid {
        let mut out = [0u8; 16];
        out.copy_from_slice(&self.0[..16]);
        Snid(out)
    }

    pub fn to_tensor256_words(self) -> (i64, i64, i64, i64) {
        (
            i64::from_be_bytes(self.0[0..8].try_into().unwrap()),
            i64::from_be_bytes(self.0[8..16].try_into().unwrap()),
            i64::from_be_bytes(self.0[16..24].try_into().unwrap()),
            i64::from_be_bytes(self.0[24..32].try_into().unwrap()),
        )
    }
}

#[derive(Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Hash, Debug)]
#[repr(transparent)]
pub struct Xid(pub [u8; 32]);

impl Xid {
    pub fn from_parts(head: Snid, edge_hash: [u8; 16]) -> Self {
        let mut out = [0u8; 32];
        out[..16].copy_from_slice(&head.0);
        out[16..].copy_from_slice(&edge_hash);
        Self(out)
    }

    pub fn head(&self) -> Snid {
        let mut out = [0u8; 16];
        out.copy_from_slice(&self.0[..16]);
        Snid(out)
    }

    pub fn to_tensor256_words(self) -> (i64, i64, i64, i64) {
        (
            i64::from_be_bytes(self.0[0..8].try_into().unwrap()),
            i64::from_be_bytes(self.0[8..16].try_into().unwrap()),
            i64::from_be_bytes(self.0[16..24].try_into().unwrap()),
            i64::from_be_bytes(self.0[24..32].try_into().unwrap()),
        )
    }
}

#[derive(Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Hash, Debug)]
#[repr(transparent)]
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
        let mut mac =
            <HmacSha256 as hmac::Mac>::new_from_slice(key).map_err(|_| Error::InvalidKey)?;
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
            .map(|expected| expected.0.ct_eq(&self.0).into())
            .unwrap_or(false)
    }

    pub fn to_tensor256_words(self) -> (i64, i64, i64, i64) {
        (
            i64::from_be_bytes(self.0[0..8].try_into().unwrap()),
            i64::from_be_bytes(self.0[8..16].try_into().unwrap()),
            i64::from_be_bytes(self.0[16..24].try_into().unwrap()),
            i64::from_be_bytes(self.0[24..32].try_into().unwrap()),
        )
    }
}

#[derive(Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Hash, Debug)]
#[repr(transparent)]
pub struct Eid(pub u64);

impl Eid {
    pub fn from_parts(unix_millis: u64, counter: u16) -> Self {
        Self((unix_millis << 16) | counter as u64)
    }

    pub fn to_bytes(self) -> [u8; 8] {
        self.0.to_be_bytes()
    }

    pub fn counter(self) -> u16 {
        self.0 as u16
    }

    pub fn timestamp_millis(self) -> u64 {
        self.0 >> 16
    }
}

#[derive(Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Hash, Debug)]
pub struct Bid {
    pub topology: Snid,
    pub content: [u8; 32],
}

impl Bid {
    pub fn from_parts(topology: Snid, content: [u8; 32]) -> Self {
        Self { topology, content }
    }

    pub fn wire(&self) -> Result<String, Error> {
        let mut out = [0u8; 81];
        Ok(self.write_wire(&mut out)?.to_owned())
    }

    pub fn write_wire<'a>(&self, out: &'a mut [u8; 81]) -> Result<&'a str, Error> {
        let mut payload = [0u8; 24];
        let payload = encode_payload_to(self.topology.0, &mut payload);
        let content_offset = 4 + payload.len() + 1;

        out[..4].copy_from_slice(b"CAS:");
        out[4..4 + payload.len()].copy_from_slice(payload.as_bytes());
        out[4 + payload.len()] = b':';
        let content_len = {
            let content = encode_base32_32_lower_to(
                &self.content,
                (&mut out[content_offset..content_offset + 52])
                    .try_into()
                    .expect("fixed content buffer"),
            );
            content.len()
        };
        unsafe {
            Ok(std::str::from_utf8_unchecked(
                &out[..content_offset + content_len],
            ))
        }
    }

    pub fn append_wire(&self, out: &mut Vec<u8>) -> Result<(), Error> {
        let mut encoded = [0u8; 81];
        let encoded = self.write_wire(&mut encoded)?;
        out.extend_from_slice(encoded.as_bytes());
        Ok(())
    }

    pub fn parse_wire(value: &str) -> Result<Self, Error> {
        use crate::encoding::decode_payload;
        let rest = value.strip_prefix("CAS:").ok_or(Error::InvalidFormat)?;
        let sep = rest.find(':').ok_or(Error::InvalidFormat)?;
        let topology = Snid(decode_payload(&rest[..sep])?);
        let out = decode_base32_32(&rest[sep + 1..]).map_err(|_| Error::InvalidContentHash)?;
        Ok(Self {
            topology,
            content: out,
        })
    }

    pub fn r2_key(&self) -> String {
        let mut out = [0u8; 52];
        self.write_r2_key(&mut out).to_owned()
    }

    pub fn write_r2_key<'a>(&self, out: &'a mut [u8; 52]) -> &'a str {
        encode_base32_32_lower_to(&self.content, out)
    }

    pub fn append_r2_key(&self, out: &mut Vec<u8>) {
        let mut encoded = [0u8; 52];
        let encoded = self.write_r2_key(&mut encoded);
        out.extend_from_slice(encoded.as_bytes());
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_nid_from_parts() {
        let head = Snid::from_bytes([1u8; 16]);
        let semantic = [2u8; 16];
        let nid = Nid::from_parts(head, semantic);
        assert_eq!(nid.head(), head);
        assert_eq!(nid.semantic_hash(), semantic);
    }

    #[test]
    fn test_nid_head() {
        let head = Snid::from_bytes([1u8; 16]);
        let semantic = [2u8; 16];
        let nid = Nid::from_parts(head, semantic);
        assert_eq!(nid.head(), head);
    }

    #[test]
    fn test_nid_semantic_hash() {
        let head = Snid::from_bytes([1u8; 16]);
        let semantic = [2u8; 16];
        let nid = Nid::from_parts(head, semantic);
        assert_eq!(nid.semantic_hash(), semantic);
    }

    #[test]
    fn test_nid_quantize() {
        let vector = vec![1.0f32, -1.0, 0.5, 0.0];
        let quantized = Nid::quantize(&vector);
        assert_ne!(quantized, [0u8; 16]);
    }

    #[test]
    fn test_nid_quantize_empty() {
        let vector: Vec<f32> = vec![];
        let quantized = Nid::quantize(&vector);
        assert_eq!(quantized, [0u8; 16]);
    }

    #[test]
    fn test_nid_hamming_distance() {
        let head = Snid::from_bytes([1u8; 16]);
        let semantic1 = [0u8; 16];
        let semantic2 = [0xFFu8; 16];
        let nid1 = Nid::from_parts(head, semantic1);
        let nid2 = Nid::from_parts(head, semantic2);
        let distance = nid1.hamming_distance(&nid2);
        assert_eq!(distance, 128); // All bits different
    }

    #[test]
    fn test_nid_hamming_distance_self() {
        let head = Snid::from_bytes([1u8; 16]);
        let semantic = [2u8; 16];
        let nid = Nid::from_parts(head, semantic);
        assert_eq!(nid.hamming_distance(&nid), 0);
    }

    #[test]
    fn test_nid_to_tensor256_words() {
        let head = Snid::from_bytes([1u8; 16]);
        let semantic = [2u8; 16];
        let nid = Nid::from_parts(head, semantic);
        let (w0, w1, w2, w3) = nid.to_tensor256_words();
        assert_ne!(w0, 0);
        assert_ne!(w1, 0);
        assert_ne!(w2, 0);
        assert_ne!(w3, 0);
    }

    #[test]
    fn test_lid_from_parts() {
        let head = Snid::from_bytes([1u8; 16]);
        let prev = [2u8; 32];
        let payload = b"test";
        let key = b"0123456789abcdef";
        let lid = Lid::from_parts(head, prev, payload, key).unwrap();
        assert_eq!(lid.head(), head);
    }

    #[test]
    fn test_lid_from_parts_empty_key() {
        let head = Snid::from_bytes([1u8; 16]);
        let prev = [2u8; 32];
        let payload = b"test";
        let key = b"";
        let result = Lid::from_parts(head, prev, payload, key);
        assert!(matches!(result, Err(Error::InvalidKey)));
    }

    #[test]
    fn test_lid_head() {
        let head = Snid::from_bytes([1u8; 16]);
        let prev = [2u8; 32];
        let payload = b"test";
        let key = b"0123456789abcdef";
        let lid = Lid::from_parts(head, prev, payload, key).unwrap();
        assert_eq!(lid.head(), head);
    }

    #[test]
    fn test_lid_to_tensor256_words() {
        let head = Snid::from_bytes([1u8; 16]);
        let prev = [2u8; 32];
        let payload = b"test";
        let key = b"0123456789abcdef";
        let lid = Lid::from_parts(head, prev, payload, key).unwrap();
        let (w0, w1, w2, w3) = lid.to_tensor256_words();
        assert_ne!(w0, 0);
        assert_ne!(w1, 0);
        assert_ne!(w2, 0);
        assert_ne!(w3, 0);
    }

    #[test]
    fn test_wid_from_parts() {
        let head = Snid::from_bytes([1u8; 16]);
        let scenario = [2u8; 16];
        let wid = Wid::from_parts(head, scenario);
        assert_eq!(wid.head(), head);
    }

    #[test]
    fn test_wid_head() {
        let head = Snid::from_bytes([1u8; 16]);
        let scenario = [2u8; 16];
        let wid = Wid::from_parts(head, scenario);
        assert_eq!(wid.head(), head);
    }

    #[test]
    fn test_wid_to_tensor256_words() {
        let head = Snid::from_bytes([1u8; 16]);
        let scenario = [2u8; 16];
        let wid = Wid::from_parts(head, scenario);
        let (w0, w1, w2, w3) = wid.to_tensor256_words();
        assert_ne!(w0, 0);
        assert_ne!(w1, 0);
        assert_ne!(w2, 0);
        assert_ne!(w3, 0);
    }

    #[test]
    fn test_xid_from_parts() {
        let head = Snid::from_bytes([1u8; 16]);
        let edge = [2u8; 16];
        let xid = Xid::from_parts(head, edge);
        assert_eq!(xid.head(), head);
    }

    #[test]
    fn test_xid_head() {
        let head = Snid::from_bytes([1u8; 16]);
        let edge = [2u8; 16];
        let xid = Xid::from_parts(head, edge);
        assert_eq!(xid.head(), head);
    }

    #[test]
    fn test_xid_to_tensor256_words() {
        let head = Snid::from_bytes([1u8; 16]);
        let edge = [2u8; 16];
        let xid = Xid::from_parts(head, edge);
        let (w0, w1, w2, w3) = xid.to_tensor256_words();
        assert_ne!(w0, 0);
        assert_ne!(w1, 0);
        assert_ne!(w2, 0);
        assert_ne!(w3, 0);
    }

    #[test]
    fn test_kid_from_parts() {
        let head = Snid::from_bytes([1u8; 16]);
        let actor = Snid::from_bytes([2u8; 16]);
        let resource = b"resource";
        let capability = b"read";
        let key = b"0123456789abcdef";
        let kid = Kid::from_parts(head, actor, resource, capability, key).unwrap();
        assert_eq!(kid.head(), head);
    }

    #[test]
    fn test_kid_from_parts_empty_key() {
        let head = Snid::from_bytes([1u8; 16]);
        let actor = Snid::from_bytes([2u8; 16]);
        let resource = b"resource";
        let capability = b"read";
        let key = b"";
        let result = Kid::from_parts(head, actor, resource, capability, key);
        assert!(matches!(result, Err(Error::InvalidKey)));
    }

    #[test]
    fn test_kid_verify() {
        let head = Snid::from_bytes([1u8; 16]);
        let actor = Snid::from_bytes([2u8; 16]);
        let resource = b"resource";
        let capability = b"read";
        let key = b"0123456789abcdef";
        let kid = Kid::from_parts(head, actor, resource, capability, key).unwrap();
        assert!(kid.verify(actor, resource, capability, key));
    }

    #[test]
    fn test_kid_verify_wrong_capability() {
        let head = Snid::from_bytes([1u8; 16]);
        let actor = Snid::from_bytes([2u8; 16]);
        let resource = b"resource";
        let capability = b"read";
        let key = b"0123456789abcdef";
        let kid = Kid::from_parts(head, actor, resource, capability, key).unwrap();
        assert!(!kid.verify(actor, resource, b"write", key));
    }

    #[test]
    fn test_kid_verify_wrong_key() {
        let head = Snid::from_bytes([1u8; 16]);
        let actor = Snid::from_bytes([2u8; 16]);
        let resource = b"resource";
        let capability = b"read";
        let key = b"0123456789abcdef";
        let kid = Kid::from_parts(head, actor, resource, capability, key).unwrap();
        assert!(!kid.verify(actor, resource, capability, b"different-key"));
    }

    #[test]
    fn test_kid_to_tensor256_words() {
        let head = Snid::from_bytes([1u8; 16]);
        let actor = Snid::from_bytes([2u8; 16]);
        let resource = b"resource";
        let capability = b"read";
        let key = b"0123456789abcdef";
        let kid = Kid::from_parts(head, actor, resource, capability, key).unwrap();
        let (w0, w1, w2, w3) = kid.to_tensor256_words();
        assert_ne!(w0, 0);
        assert_ne!(w1, 0);
        assert_ne!(w2, 0);
        assert_ne!(w3, 0);
    }

    #[test]
    fn test_eid_from_parts() {
        let unix_millis = 1700000000123;
        let counter = 0x00FF;
        let eid = Eid::from_parts(unix_millis, counter);
        assert_eq!(eid.timestamp_millis(), unix_millis);
        assert_eq!(eid.counter(), counter);
    }

    #[test]
    fn test_eid_to_bytes() {
        let unix_millis = 1700000000123;
        let counter = 0x00FF;
        let eid = Eid::from_parts(unix_millis, counter);
        let bytes = eid.to_bytes();
        assert_eq!(bytes.len(), 8);
    }

    #[test]
    fn test_eid_counter() {
        let unix_millis = 1700000000123;
        let counter = 0x00FF;
        let eid = Eid::from_parts(unix_millis, counter);
        assert_eq!(eid.counter(), counter);
    }

    #[test]
    fn test_eid_timestamp_millis() {
        let unix_millis = 1700000000123;
        let counter = 0x00FF;
        let eid = Eid::from_parts(unix_millis, counter);
        assert_eq!(eid.timestamp_millis(), unix_millis);
    }

    #[test]
    fn test_bid_from_parts() {
        let topology = Snid::from_bytes([1u8; 16]);
        let content = [2u8; 32];
        let bid = Bid::from_parts(topology, content);
        assert_eq!(bid.topology, topology);
        assert_eq!(bid.content, content);
    }

    #[test]
    fn test_bid_wire() {
        let topology = Snid::from_bytes([1u8; 16]);
        let content = [2u8; 32];
        let bid = Bid::from_parts(topology, content);
        let wire = bid.wire().unwrap();
        assert!(wire.starts_with("CAS:"));
    }

    #[test]
    fn test_bid_write_wire_matches_wire() {
        let topology = Snid::from_bytes([1u8; 16]);
        let content = [2u8; 32];
        let bid = Bid::from_parts(topology, content);
        let mut out = [0u8; 81];
        let written = bid.write_wire(&mut out).unwrap();
        assert_eq!(written, bid.wire().unwrap());

        let mut appended = Vec::new();
        bid.append_wire(&mut appended).unwrap();
        assert_eq!(appended, written.as_bytes());
    }

    #[test]
    fn test_bid_parse_wire() {
        let topology = Snid::from_bytes([1u8; 16]);
        let content = [2u8; 32];
        let bid = Bid::from_parts(topology, content);
        let wire = bid.wire().unwrap();
        let parsed = Bid::parse_wire(&wire).unwrap();
        assert_eq!(parsed.topology, topology);
        assert_eq!(parsed.content, content);
    }

    #[test]
    fn test_bid_parse_wire_invalid_format() {
        let result = Bid::parse_wire("invalid");
        assert!(matches!(result, Err(Error::InvalidFormat)));
    }

    #[test]
    fn test_bid_r2_key() {
        let topology = Snid::from_bytes([1u8; 16]);
        let content = [2u8; 32];
        let bid = Bid::from_parts(topology, content);
        let r2_key = bid.r2_key();
        assert!(!r2_key.is_empty());
    }

    #[test]
    fn test_bid_write_r2_key_matches_r2_key() {
        let topology = Snid::from_bytes([1u8; 16]);
        let content = [2u8; 32];
        let bid = Bid::from_parts(topology, content);
        let mut out = [0u8; 52];
        let written = bid.write_r2_key(&mut out);
        assert_eq!(written, bid.r2_key());

        let mut appended = Vec::new();
        bid.append_r2_key(&mut appended);
        assert_eq!(appended, written.as_bytes());
    }
}

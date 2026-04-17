use getrandom::getrandom;
use hmac::{Hmac, Mac};
use serde::Deserialize;
use sha2::Sha256;
use std::cell::RefCell;
use std::fmt;
use std::process;
use std::time::{SystemTime, UNIX_EPOCH};

type HmacSha256 = Hmac<Sha256>;

pub const BASE58_ALPHABET: &[u8; 58] =
    b"123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz";

const CRC8_TABLE: [u8; 256] = [
    0x00, 0x07, 0x0E, 0x09, 0x1C, 0x1B, 0x12, 0x15, 0x38, 0x3F, 0x36, 0x31, 0x24, 0x23, 0x2A, 0x2D,
    0x70, 0x77, 0x7E, 0x79, 0x6C, 0x6B, 0x62, 0x65, 0x48, 0x4F, 0x46, 0x41, 0x54, 0x53, 0x5A, 0x5D,
    0xE0, 0xE7, 0xEE, 0xE9, 0xFC, 0xFB, 0xF2, 0xF5, 0xD8, 0xDF, 0xD6, 0xD1, 0xC4, 0xC3, 0xCA, 0xCD,
    0x90, 0x97, 0x9E, 0x99, 0x8C, 0x8B, 0x82, 0x85, 0xA8, 0xAF, 0xA6, 0xA1, 0xB4, 0xB3, 0xBA, 0xBD,
    0xC7, 0xC0, 0xC9, 0xCE, 0xDB, 0xDC, 0xD5, 0xD2, 0xFF, 0xF8, 0xF1, 0xF6, 0xE3, 0xE4, 0xED, 0xEA,
    0xB7, 0xB0, 0xB9, 0xBE, 0xAB, 0xAC, 0xA5, 0xA2, 0x8F, 0x88, 0x81, 0x86, 0x93, 0x94, 0x9D, 0x9A,
    0x27, 0x20, 0x29, 0x2E, 0x3B, 0x3C, 0x35, 0x32, 0x1F, 0x18, 0x11, 0x16, 0x03, 0x04, 0x0D, 0x0A,
    0x57, 0x50, 0x59, 0x5E, 0x4B, 0x4C, 0x45, 0x42, 0x6F, 0x68, 0x61, 0x66, 0x73, 0x74, 0x7D, 0x7A,
    0x89, 0x8E, 0x87, 0x80, 0x95, 0x92, 0x9B, 0x9C, 0xB1, 0xB6, 0xBF, 0xB8, 0xAD, 0xAA, 0xA3, 0xA4,
    0xF9, 0xFE, 0xF7, 0xF0, 0xE5, 0xE2, 0xEB, 0xEC, 0xC1, 0xC6, 0xCF, 0xC8, 0xDD, 0xDA, 0xD3, 0xD4,
    0x69, 0x6E, 0x67, 0x60, 0x75, 0x72, 0x7B, 0x7C, 0x51, 0x56, 0x5F, 0x58, 0x4D, 0x4A, 0x43, 0x44,
    0x19, 0x1E, 0x17, 0x10, 0x05, 0x02, 0x0B, 0x0C, 0x21, 0x26, 0x2F, 0x28, 0x3D, 0x3A, 0x33, 0x34,
    0x4E, 0x49, 0x40, 0x47, 0x52, 0x55, 0x5C, 0x5B, 0x76, 0x71, 0x78, 0x7F, 0x6A, 0x6D, 0x64, 0x63,
    0x3E, 0x39, 0x30, 0x37, 0x22, 0x25, 0x2C, 0x2B, 0x06, 0x01, 0x08, 0x0F, 0x1A, 0x1D, 0x14, 0x13,
    0xAE, 0xA9, 0xA0, 0xA7, 0xB2, 0xB5, 0xBC, 0xBB, 0x96, 0x91, 0x98, 0x9F, 0x8A, 0x8D, 0x84, 0x83,
    0xDE, 0xD9, 0xD0, 0xD7, 0xC2, 0xC5, 0xCC, 0xCB, 0xE6, 0xE1, 0xE8, 0xEF, 0xFA, 0xFD, 0xF4, 0xF3,
];

#[derive(Debug)]
pub enum Error {
    InvalidLength,
    InvalidFormat,
    InvalidAtom,
    InvalidPayload,
    ChecksumMismatch,
    InvalidContentHash,
    InvalidKey,
    Hex(hex::FromHexError),
    Json(serde_json::Error),
}

impl From<hex::FromHexError> for Error {
    fn from(value: hex::FromHexError) -> Self {
        Self::Hex(value)
    }
}

impl From<serde_json::Error> for Error {
    fn from(value: serde_json::Error) -> Self {
        Self::Json(value)
    }
}

#[derive(Clone, Copy, PartialEq, Eq, Hash)]
pub struct Snid(pub [u8; 16]);

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

impl fmt::Debug for Snid {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "Snid({})", hex::encode(self.0))
    }
}

impl Snid {
    pub fn new_fast() -> Self {
        GENERATOR.with(|cell| cell.borrow_mut().next())
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
        GENERATOR.with(|cell| {
            let state = &mut *cell.borrow_mut();
            for _ in 0..count {
                out.extend_from_slice(&state.next().0);
            }
        });
        out
    }

    #[cfg(feature = "data")]
    pub fn generate_tensor_batch(count: usize) -> Vec<i64> {
        let mut out = Vec::with_capacity(count.saturating_mul(2));
        GENERATOR.with(|cell| {
            let state = &mut *cell.borrow_mut();
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
        GENERATOR.with(|cell| {
            let state = &mut *cell.borrow_mut();
            for _ in 0..count {
                let (hi, lo) = state.next().to_tensor_words();
                out.extend_from_slice(&hi.to_be_bytes());
                out.extend_from_slice(&lo.to_be_bytes());
            }
        });
        out
    }
}

thread_local! {
    static GENERATOR: RefCell<GeneratorState> = RefCell::new(GeneratorState::init());
}

struct GeneratorState {
    last_ms: u64,
    sequence: u16,
    machine_id: u32,
    state: [u64; 4],
    pid: u32,
}

impl GeneratorState {
    fn init() -> Self {
        let mut seed = [0u8; 8];
        getrandom(&mut seed).expect("os rng");
        let mut z = u64::from_le_bytes(seed);
        let pid = process::id();
        let machine_id = (splitmix64(&mut z) as u32 ^ pid) & 0x00FF_FFFF;
        Self {
            last_ms: current_time_ms(),
            sequence: 0,
            machine_id,
            state: [
                splitmix64(&mut z),
                splitmix64(&mut z),
                splitmix64(&mut z),
                splitmix64(&mut z),
            ],
            pid,
        }
    }

    fn next(&mut self) -> Snid {
        if self.pid != process::id() {
            *self = Self::init();
        }

        let mut ms = current_time_ms();
        if ms > self.last_ms {
            self.last_ms = ms;
            self.sequence = 0;
        } else {
            self.sequence = self.sequence.wrapping_add(1);
            if self.sequence > 0x3FFF {
                self.last_ms += 1;
                ms = self.last_ms;
                self.sequence = 0;
            } else {
                ms = self.last_ms;
            }
        }

        let result = self.state[1].wrapping_mul(5).rotate_left(7).wrapping_mul(9);
        let t = self.state[1] << 17;
        self.state[2] ^= self.state[0];
        self.state[3] ^= self.state[1];
        self.state[1] ^= self.state[2];
        self.state[0] ^= self.state[3];
        self.state[2] ^= t;
        self.state[3] = self.state[3].rotate_left(45);

        let seq = self.sequence as u64;
        let hi = (ms << 16) | 0x7000 | (seq >> 2);
        let lo = 0x8000_0000_0000_0000
            | ((seq & 0x03) << 60)
            | ((self.machine_id as u64 & 0x00FF_FFFF) << 36)
            | ((result >> 28) & 0xFFFF_FFFFF);
        let mut out = [0u8; 16];
        out[..8].copy_from_slice(&hi.to_be_bytes());
        out[8..].copy_from_slice(&lo.to_be_bytes());
        Snid(out)
    }
}

fn current_time_ms() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .expect("clock")
        .as_millis() as u64
}

fn splitmix64(seed: &mut u64) -> u64 {
    *seed = seed.wrapping_add(0x9e3779b97f4a7c15);
    let mut z = *seed;
    z = (z ^ (z >> 30)).wrapping_mul(0xbf58476d1ce4e5b9);
    z = (z ^ (z >> 27)).wrapping_mul(0x94d049bb133111eb);
    z ^ (z >> 31)
}

fn expand_hash_material(hash: &[u8]) -> [u8; 16] {
    let mut out = [0u8; 16];
    if hash.is_empty() {
        return out;
    }
    for i in 0..16 {
        out[i] = hash[i % hash.len()];
    }
    out
}

#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub struct Nid(pub [u8; 32]);

impl Nid {
    pub fn from_parts(head: Snid, semantic_hash: [u8; 16]) -> Self {
        let mut out = [0u8; 32];
        out[..16].copy_from_slice(&head.0);
        out[16..].copy_from_slice(&semantic_hash);
        Self(out)
    }

    pub fn head(&self) -> Snid {
        let mut out = [0u8; 16];
        out.copy_from_slice(&self.0[..16]);
        Snid(out)
    }

    pub fn semantic_hash(&self) -> [u8; 16] {
        let mut out = [0u8; 16];
        out.copy_from_slice(&self.0[16..]);
        out
    }

    pub fn quantize(vector: &[f32]) -> [u8; 16] {
        let mut out = [0u8; 16];
        for (i, value) in vector.iter().take(128).enumerate() {
            if *value > 0.0 {
                out[i / 8] |= 1 << (7 - (i % 8));
            }
        }
        out
    }

    pub fn hamming_distance(&self, other: &Nid) -> u32 {
        self.semantic_hash()
            .iter()
            .zip(other.semantic_hash())
            .map(|(a, b)| (a ^ b).count_ones())
            .sum()
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

#[derive(Clone, Copy, PartialEq, Eq, Debug)]
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
        let mut mac = HmacSha256::new_from_slice(key).map_err(|_| Error::InvalidKey)?;
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

    pub fn to_tensor256_words(&self) -> (i64, i64, i64, i64) {
        (
            i64::from_be_bytes(self.0[0..8].try_into().unwrap()),
            i64::from_be_bytes(self.0[8..16].try_into().unwrap()),
            i64::from_be_bytes(self.0[16..24].try_into().unwrap()),
            i64::from_be_bytes(self.0[24..32].try_into().unwrap()),
        )
    }
}

#[derive(Clone, Copy, PartialEq, Eq, Debug)]
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

    pub fn to_tensor256_words(&self) -> (i64, i64, i64, i64) {
        (
            i64::from_be_bytes(self.0[0..8].try_into().unwrap()),
            i64::from_be_bytes(self.0[8..16].try_into().unwrap()),
            i64::from_be_bytes(self.0[16..24].try_into().unwrap()),
            i64::from_be_bytes(self.0[24..32].try_into().unwrap()),
        )
    }
}

#[derive(Clone, Copy, PartialEq, Eq, Debug)]
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

    pub fn to_tensor256_words(&self) -> (i64, i64, i64, i64) {
        (
            i64::from_be_bytes(self.0[0..8].try_into().unwrap()),
            i64::from_be_bytes(self.0[8..16].try_into().unwrap()),
            i64::from_be_bytes(self.0[16..24].try_into().unwrap()),
            i64::from_be_bytes(self.0[24..32].try_into().unwrap()),
        )
    }
}

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
        let mut mac = HmacSha256::new_from_slice(key).map_err(|_| Error::InvalidKey)?;
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

#[derive(Clone, Copy, PartialEq, Eq, Debug)]
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

#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub struct Bid {
    pub topology: Snid,
    pub content: [u8; 32],
}

impl Bid {
    pub fn from_parts(topology: Snid, content: [u8; 32]) -> Self {
        Self { topology, content }
    }

    pub fn wire(self) -> Result<String, Error> {
        let payload = encode_payload(self.topology.0);
        let content = base32::encode(base32::Alphabet::Rfc4648 { padding: false }, &self.content)
            .to_lowercase();
        Ok(format!("CAS:{payload}:{content}"))
    }

    pub fn parse_wire(value: &str) -> Result<Self, Error> {
        let parts: Vec<_> = value.split(':').collect();
        if parts.len() != 3 || parts[0] != "CAS" {
            return Err(Error::InvalidFormat);
        }
        let topology = Snid(decode_payload(parts[1])?);
        let content = base32::decode(
            base32::Alphabet::Rfc4648 { padding: false },
            &parts[2].to_uppercase(),
        )
        .ok_or(Error::InvalidContentHash)?;
        if content.len() != 32 {
            return Err(Error::InvalidContentHash);
        }
        let mut out = [0u8; 32];
        out.copy_from_slice(&content);
        Ok(Self {
            topology,
            content: out,
        })
    }

    pub fn r2_key(self) -> String {
        base32::encode(base32::Alphabet::Rfc4648 { padding: false }, &self.content).to_lowercase()
    }
}

pub fn crc8(data: &[u8]) -> u8 {
    let mut crc = 0u8;
    for byte in data {
        crc = CRC8_TABLE[(crc ^ byte) as usize];
    }
    crc
}

pub fn encode_payload(bytes: [u8; 16]) -> String {
    let mut buf = [0u8; 24];
    let mut idx = 23usize;
    let checksum = crc8(&bytes);
    buf[idx] = BASE58_ALPHABET[(checksum % 58) as usize];
    idx -= 1;

    let mut hi = u64::from_be_bytes(bytes[..8].try_into().unwrap());
    let mut lo = u64::from_be_bytes(bytes[8..].try_into().unwrap());
    while hi > 0 || lo > 0 {
        let qhi = hi / 58;
        let rhi = hi - qhi * 58;
        let dividend = ((rhi as u128) << 64) | lo as u128;
        let qlo = (dividend / 58) as u64;
        let rem = (dividend % 58) as usize;
        hi = qhi;
        lo = qlo;
        buf[idx] = BASE58_ALPHABET[rem];
        if idx == 0 {
            break;
        }
        idx -= 1;
    }

    for byte in bytes {
        if byte == 0 {
            buf[idx] = b'1';
            if idx == 0 {
                break;
            }
            idx -= 1;
        } else {
            break;
        }
    }
    String::from_utf8(buf[idx + 1..].to_vec()).unwrap()
}

pub fn decode_payload(payload: &str) -> Result<[u8; 16], Error> {
    if payload.len() < 2 {
        return Err(Error::InvalidPayload);
    }
    let (body, checksum_part) = payload.split_at(payload.len() - 1);
    let mut out = [0u8; 16];
    for ch in body.bytes() {
        let value = BASE58_ALPHABET
            .iter()
            .position(|candidate| *candidate == ch)
            .ok_or(Error::InvalidPayload)? as u32;
        let mut carry = value;
        for byte in out.iter_mut().rev() {
            let res = (*byte as u32) * 58 + carry;
            *byte = res as u8;
            carry = res >> 8;
        }
        if carry > 0 {
            return Err(Error::InvalidPayload);
        }
    }
    let expected = BASE58_ALPHABET[(crc8(&out) % 58) as usize] as char;
    if checksum_part.chars().next() != Some(expected) {
        return Err(Error::ChecksumMismatch);
    }
    Ok(out)
}

pub fn encode_fixed64_pair(hi: i64, lo: i64) -> [u8; 16] {
    let mut out = [0u8; 16];
    out[..8].copy_from_slice(&hi.to_be_bytes());
    out[8..].copy_from_slice(&lo.to_be_bytes());
    out
}

pub fn decode_fixed64_pair(raw: &[u8]) -> Result<(i64, i64), Error> {
    if raw.len() != 16 {
        return Err(Error::InvalidLength);
    }
    Ok((
        i64::from_be_bytes(raw[..8].try_into().unwrap()),
        i64::from_be_bytes(raw[8..16].try_into().unwrap()),
    ))
}

fn split_wire(value: &str) -> Result<(&str, &str, char), Error> {
    let bytes = value.as_bytes();
    for idx in 0..bytes.len().min(9) {
        if bytes[idx] == b':' || bytes[idx] == b'_' {
            if idx < 2 {
                return Err(Error::InvalidFormat);
            }
            return Ok((&value[..idx], &value[idx + 1..], bytes[idx] as char));
        }
    }
    Err(Error::InvalidFormat)
}

#[derive(Deserialize)]
pub struct VectorFile {
    pub core: Vec<CoreVector>,
    pub spatial: SpatialVector,
    pub neural: NeuralVector,
    pub ledger: LedgerVector,
    pub world: WorldVector,
    pub edge: EdgeVector,
    pub capability: CapabilityVector,
    pub ephemeral: EIDVector,
    pub bid: BIDVector,
    pub compatibility: CompatibilityVector,
    pub negative: NegativeVector,
}

#[derive(Deserialize)]
pub struct CoreVector {
    pub atom: String,
    pub bytes_hex: String,
    pub wire: String,
    pub underscore_wire: String,
    pub timestamp_millis: i64,
    pub tensor_hi: i64,
    pub tensor_lo: i64,
    pub llm_format: LLMFormatVector,
    pub llm_format_v2: LLMFormatVectorV2,
    pub time_bin_hour: i64,
    pub ghosted_bytes_hex: String,
}

#[derive(Deserialize)]
pub struct LLMFormatVector {
    pub atom: String,
    pub timestamp_millis: i64,
    pub machine_or_shard: u32,
    pub sequence: u16,
}

#[derive(Deserialize)]
pub struct LLMFormatVectorV2 {
    pub kind: String,
    pub atom: String,
    pub timestamp_millis: Option<i64>,
    pub spatial_anchor: Option<u64>,
    pub machine_or_shard: Option<u32>,
    pub sequence: Option<u16>,
    pub ghosted: bool,
}

#[derive(Deserialize)]
pub struct SpatialVector {
    pub atom: String,
    pub bytes_hex: String,
    pub wire: String,
    pub h3_cell_hex: String,
}

#[derive(Deserialize)]
pub struct NeuralVector {
    pub head_hex: String,
    pub bytes_hex: String,
    pub semantic_hex: String,
    pub hamming_to_zero: i32,
}

#[derive(Deserialize)]
pub struct LedgerVector {
    pub head_hex: String,
    pub bytes_hex: String,
    pub prev_hex: String,
    pub payload_hex: String,
    pub key_hex: String,
    pub blake3_hex: String,
}

#[derive(Deserialize)]
pub struct WorldVector {
    pub head_hex: String,
    pub bytes_hex: String,
    pub scenario_hex: String,
    pub tensor_words: [i64; 4],
}

#[derive(Deserialize)]
pub struct EdgeVector {
    pub head_hex: String,
    pub bytes_hex: String,
    pub edge_hex: String,
    pub tensor_words: [i64; 4],
}

#[derive(Deserialize)]
pub struct CapabilityVector {
    pub head_hex: String,
    pub actor_hex: String,
    pub resource_hex: String,
    pub capability_hex: String,
    pub key_hex: String,
    pub bytes_hex: String,
    pub tensor_words: [i64; 4],
}

#[derive(Deserialize)]
pub struct EIDVector {
    pub bytes_hex: String,
    pub timestamp_millis: u64,
    pub counter: u16,
}

#[derive(Deserialize)]
pub struct BIDVector {
    pub topology_hex: String,
    pub content_hex: String,
    pub wire: String,
    pub r2_key: String,
    pub neo4j_id: String,
}

#[derive(Deserialize)]
pub struct CompatibilityVector {
    pub bytes_hex: String,
    pub wire: String,
}

#[derive(Deserialize)]
pub struct NegativeVector {
    pub invalid_atom_wire: String,
    pub invalid_binary_hex: String,
    pub invalid_wire_checksum: String,
    pub invalid_adapter_hex: String,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn validates_vectors() {
        let raw = std::fs::read_to_string("../conformance/vectors.json")
            .or_else(|_| std::fs::read_to_string("../SNID/conformance/vectors.json"))
            .expect("vectors.json");
        let vectors: VectorFile = serde_json::from_str(&raw).expect("parse vectors");

        for case in vectors.core {
            let expected = Snid::from_hex(&case.bytes_hex).unwrap();
            let (parsed, atom) = Snid::parse_wire(&case.wire).unwrap();
            assert_eq!(parsed, expected);
            assert_eq!(atom, case.atom);
            assert_eq!(expected.to_wire(&case.atom).unwrap(), case.wire);
            assert_eq!(expected.timestamp_millis(), case.timestamp_millis);
            assert_eq!(expected.to_tensor_words(), (case.tensor_hi, case.tensor_lo));
            let llm = expected.to_llm_format(&case.atom).unwrap();
            assert_eq!(llm.atom, case.llm_format.atom);
            assert_eq!(llm.timestamp_millis, case.llm_format.timestamp_millis);
            assert_eq!(llm.machine_or_shard, case.llm_format.machine_or_shard);
            assert_eq!(llm.sequence, case.llm_format.sequence);
            let llm_v2 = expected.to_llm_format_v2(&case.atom).unwrap();
            assert_eq!(llm_v2.kind, case.llm_format_v2.kind);
            assert_eq!(llm_v2.atom, case.llm_format_v2.atom);
            assert_eq!(llm_v2.timestamp_millis, case.llm_format_v2.timestamp_millis);
            assert_eq!(llm_v2.machine_or_shard, case.llm_format_v2.machine_or_shard);
            assert_eq!(llm_v2.sequence, case.llm_format_v2.sequence);
            assert_eq!(llm_v2.ghosted, case.llm_format_v2.ghosted);
            assert_eq!(expected.time_bin(3_600_000), case.time_bin_hour);
            assert_eq!(
                hex::encode(expected.with_ghost_bit(true).0),
                case.ghosted_bytes_hex
            );
            let (underscore, _) = Snid::parse_wire(&case.underscore_wire).unwrap();
            assert_eq!(underscore, expected);
        }

        let sgid = Snid::from_hex(&vectors.spatial.bytes_hex).unwrap();
        assert_eq!(
            sgid.to_wire(&vectors.spatial.atom).unwrap(),
            vectors.spatial.wire
        );
        assert_eq!(
            format!("{:x}", sgid.h3_cell().unwrap()),
            vectors.spatial.h3_cell_hex
        );

        let head = Snid::from_hex(&vectors.neural.head_hex).unwrap();
        let semantic: [u8; 16] = hex::decode(vectors.neural.semantic_hex)
            .unwrap()
            .try_into()
            .unwrap();
        let nid = Nid::from_parts(head, semantic);
        assert_eq!(hex::encode(nid.0), vectors.neural.bytes_hex);
        assert_eq!(
            nid.hamming_distance(&Nid([0u8; 32])) as i32,
            vectors.neural.hamming_to_zero
        );

        let head = Snid::from_hex(&vectors.ledger.head_hex).unwrap();
        let prev: [u8; 32] = hex::decode(vectors.ledger.prev_hex)
            .unwrap()
            .try_into()
            .unwrap();
        let payload = hex::decode(vectors.ledger.payload_hex).unwrap();
        let key = hex::decode(vectors.ledger.key_hex).unwrap();
        let lid = Lid::from_parts(head, prev, &payload, &key).unwrap();
        assert_eq!(hex::encode(lid.0), vectors.ledger.bytes_hex);

        let world_head = Snid::from_hex(&vectors.world.head_hex).unwrap();
        let scenario: [u8; 16] = hex::decode(vectors.world.scenario_hex)
            .unwrap()
            .try_into()
            .unwrap();
        let wid = Wid::from_parts(world_head, scenario);
        assert_eq!(hex::encode(wid.0), vectors.world.bytes_hex);
        assert_eq!(
            wid.to_tensor256_words(),
            (
                vectors.world.tensor_words[0],
                vectors.world.tensor_words[1],
                vectors.world.tensor_words[2],
                vectors.world.tensor_words[3]
            )
        );

        let edge_head = Snid::from_hex(&vectors.edge.head_hex).unwrap();
        let edge_hash: [u8; 16] = hex::decode(vectors.edge.edge_hex)
            .unwrap()
            .try_into()
            .unwrap();
        let xid = Xid::from_parts(edge_head, edge_hash);
        assert_eq!(hex::encode(xid.0), vectors.edge.bytes_hex);
        assert_eq!(
            xid.to_tensor256_words(),
            (
                vectors.edge.tensor_words[0],
                vectors.edge.tensor_words[1],
                vectors.edge.tensor_words[2],
                vectors.edge.tensor_words[3]
            )
        );

        let kid_head = Snid::from_hex(&vectors.capability.head_hex).unwrap();
        let actor = Snid::from_hex(&vectors.capability.actor_hex).unwrap();
        let resource = hex::decode(vectors.capability.resource_hex).unwrap();
        let capability = hex::decode(vectors.capability.capability_hex).unwrap();
        let key = hex::decode(vectors.capability.key_hex).unwrap();
        let kid = Kid::from_parts(kid_head, actor, &resource, &capability, &key).unwrap();
        assert_eq!(hex::encode(kid.0), vectors.capability.bytes_hex);
        assert!(kid.verify(actor, &resource, &capability, &key));
        assert_eq!(
            kid.to_tensor256_words(),
            (
                vectors.capability.tensor_words[0],
                vectors.capability.tensor_words[1],
                vectors.capability.tensor_words[2],
                vectors.capability.tensor_words[3]
            )
        );

        let eid_bytes: [u8; 8] = hex::decode(vectors.ephemeral.bytes_hex)
            .unwrap()
            .try_into()
            .unwrap();
        let eid = Eid(u64::from_be_bytes(eid_bytes));
        assert_eq!(eid.timestamp_millis(), vectors.ephemeral.timestamp_millis);
        assert_eq!(eid.counter(), vectors.ephemeral.counter);

        let topology = Snid::from_hex(&vectors.bid.topology_hex).unwrap();
        let content: [u8; 32] = hex::decode(vectors.bid.content_hex)
            .unwrap()
            .try_into()
            .unwrap();
        let bid = Bid::from_parts(topology, content);
        assert_eq!(bid.wire().unwrap(), vectors.bid.wire);
        assert_eq!(bid.r2_key(), vectors.bid.r2_key);
        assert_eq!(bid.topology.to_wire("MAT").unwrap(), vectors.bid.neo4j_id);
        assert_eq!(Bid::parse_wire(&vectors.bid.wire).unwrap(), bid);

        let compat = Snid::from_hex(&vectors.compatibility.bytes_hex).unwrap();
        assert_eq!(compat.to_wire("MAT").unwrap(), vectors.compatibility.wire);
        let (parsed, _) = Snid::parse_wire(&vectors.compatibility.wire).unwrap();
        assert_eq!(parsed, compat);

        assert!(Snid::parse_wire(&vectors.negative.invalid_atom_wire).is_err());
        assert!(Snid::from_hex(&vectors.negative.invalid_binary_hex).is_err());
        assert!(Snid::parse_wire(&vectors.negative.invalid_wire_checksum).is_err());
        assert!(Snid::from_hex(&vectors.negative.invalid_adapter_hex).is_err());
    }
}

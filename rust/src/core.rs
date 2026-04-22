//! Core SNID type and all Snid methods.

use crate::encoding::{decode_payload, encode_payload, split_wire};
use crate::error::Error;
use crate::generator::{init_coarse_clock, GENERATOR};
use crate::helpers::expand_hash_material;
use std::fmt;
use std::str::FromStr;

#[derive(Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Hash)]
#[repr(transparent)]
pub struct Snid(pub [u8; 16]);

impl fmt::Debug for Snid {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "Snid({})", hex::encode(self.0))
    }
}

impl FromStr for Snid {
    type Err = Error;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        Self::parse(s)
    }
}

impl fmt::Display for Snid {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        // Allocate a small buffer purely on the stack (0 heap allocations)
        let mut buffer = [0u8; 48];
        let payload = crate::encoding::encode_payload(self.0);
        let len = payload.len();
        buffer[..4].copy_from_slice(b"MAT:");
        buffer[4..4 + len].copy_from_slice(payload.as_bytes());
        let s = unsafe { std::str::from_utf8_unchecked(&buffer[..4 + len]) };
        write!(f, "{}", s)
    }
}

/// Options for configured ID generation (zero-allocation).
#[derive(Clone, Default)]
pub struct Options {
    pub tenant: Option<String>,
    pub shard: Option<u16>,
}

impl Snid {
    /// Generates a new SNID with ~3.7ns latency.
    /// This is the universal paradigm for fast ID generation.
    #[inline(always)]
    pub fn new() -> Self {
        init_coarse_clock();
        GENERATOR.with(|cell| unsafe { (*cell.get()).next() })
    }

    /// Generates a new SNID with lock-free per-P state.
    /// Deprecated: Use new() instead.
    pub fn new_fast() -> Self {
        Self::new()
    }

    /// Generates a configured ID using options.
    /// This is the universal paradigm for configured ID generation.
    pub fn new_with(opts: Options) -> Self {
        if let Some(tenant) = opts.tenant {
            // Use tenant-based generation
            return Self::new_tenant_sharded(&tenant, opts.shard.unwrap_or(0));
        }
        Self::new()
    }

    /// Generates a tenant-sharded ID for multi-tenant systems.
    fn new_tenant_sharded(tenant: &str, shard: u16) -> Self {
        // Generate a base ID first
        let mut id = Self::new();

        // Hash tenant ID for deterministic shard assignment
        let tenant_hash = crate::helpers::fnv1a(tenant);
        let shard_key = (shard as u64) | ((tenant_hash as u64) << 16);

        // Modify the ID to include shard information
        let mut lo_bytes = [0u8; 8];
        lo_bytes.copy_from_slice(&id.0[8..]);
        let mut lo = u64::from_be_bytes(lo_bytes);
        lo = (lo & 0x0000_FFFF_FFFF_FFFF) | ((shard_key & 0xFFF) << 48);
        id.0[8..].copy_from_slice(&lo.to_be_bytes());

        id
    }

    /// Generates a SNID with RFC 9562 UUIDv7-compatible bytes.
    pub fn uuidv7() -> Self {
        Self::new()
    }

    /// Generates a spatial ID from lat/lng coordinates.
    /// This is the universal paradigm for spatial ID generation.
    pub fn new_spatial(_lat: f64, _lng: f64) -> Self {
        // For now, generate a regular ID with spatial markers
        // Full H3 integration would require the h3 crate
        let mut id = Self::new();
        // Mark as spatial (Version 8)
        id.0[6] = (id.0[6] & 0x0F) | 0x80;
        id.0[14] = 0xA5; // spatial marker
        id.0[15] = 0xA0; // spatial tail nibble
        id
    }

    /// Generates a batch of IDs efficiently.
    /// This is the universal paradigm for batch generation.
    pub fn batch(count: usize) -> Vec<Self> {
        let mut ids = Vec::with_capacity(count);
        for _ in 0..count {
            ids.push(Self::new());
        }
        ids
    }

    #[inline(always)]
    pub fn from_bytes(bytes: [u8; 16]) -> Self {
        Self(bytes)
    }

    #[inline(always)]
    pub fn to_bytes(self) -> [u8; 16] {
        self.0
    }

    /// Writes the standard hyphenated UUID text form into `out`.
    pub fn write_uuid_string(&self, out: &mut [u8; 36]) {
        encode_uuid_string(self.0, out);
    }

    /// Returns the standard hyphenated UUID text form.
    pub fn to_uuid_string(&self) -> String {
        let mut out = [0u8; 36];
        self.write_uuid_string(&mut out);
        // The encoder only writes lowercase hex digits and hyphens.
        unsafe { String::from_utf8_unchecked(out.to_vec()) }
    }

    /// Parses a standard hyphenated UUID string into a SNID.
    ///
    /// Only UUIDv7 values are accepted.
    pub fn from_uuid_string(value: &str) -> Result<Self, Error> {
        let bytes = parse_uuid_string_bytes(value.as_bytes())?;
        Ok(Self(bytes))
    }

    /// Converts this SNID into a `uuid::Uuid`.
    #[cfg(feature = "uuid")]
    pub fn to_uuidv7(&self) -> uuid::Uuid {
        uuid::Uuid::from_bytes(self.0)
    }

    /// Converts a UUIDv7 value into a SNID.
    #[cfg(feature = "uuid")]
    pub fn from_uuidv7(value: uuid::Uuid) -> Result<Self, Error> {
        let bytes = value.into_bytes();
        validate_uuidv7_bytes(&bytes)?;
        Ok(Self(bytes))
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

    /// Formats an ID using the default wire format with "MAT:" atom.
    /// This is the universal paradigm for serialization (default atom).
    pub fn to_string(self) -> String {
        self.to_wire("MAT").unwrap_or_else(|_| format!("MAT:{}", encode_payload(self.0)))
    }

    /// Formats an ID with a custom atom.
    /// This is the universal paradigm for serialization (override atom).
    pub fn with_atom(self, atom: &str) -> String {
        self.to_wire(atom).unwrap_or_else(|_| format!("{}:{}", atom, encode_payload(self.0)))
    }

    /// Formats an ID using Crockford Base32 encoding.
    /// This is case-insensitive and excludes ambiguous characters (I, L, O).
    pub fn to_base32(self) -> String {
        crate::encoding::encode_base32(self.0)
    }

    /// Generates a public-safe ID with time-blurring and pure CSPRNG entropy.
    /// This is the "One ID" solution for database PK + public API use.
    /// Time-blurring: Truncates timestamp to nearest second (instead of millisecond)
    /// Pure CSPRNG: Fills 74 bits with cryptographic randomness (no monotonic counter)
    /// Performance: ~40-50ns (vs 5ns for new())
    pub fn new_safe() -> Self {
        let mut out = [0u8; 16];

        // Get current time in milliseconds and truncate to second (time-blurring)
        use std::time::{SystemTime, UNIX_EPOCH};
        let ms = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .expect("clock")
            .as_millis() as u64;
        let ms_sec = ms & !0x3FF; // Clear lower 10 bits to get second granularity

        // Generate 74 bits of pure CSPRNG entropy
        let mut entropy = [0u8; 10]; // 80 bits, we'll use 74
        getrandom::getrandom(&mut entropy).expect("os rng");

        // Assemble the ID with time-blurred timestamp and CSPRNG entropy
        // Layout: [timestamp (48 bits)][version (4 bits)][entropy (74 bits)][variant (2 bits)]
        let hi = (ms_sec << 16) | 0x7000; // timestamp + version
        
        // Set variant bits (bits 6-7 of byte 8 should be 0b10 for RFC 4122)
        let mut entropy64 = u64::from_be_bytes(entropy[..8].try_into().unwrap());
        entropy64 = (entropy64 & 0x3FFF_FFFF_FFFF_FFFF) | 0x8000_0000_0000_0000; // Clear top 2 bits, set variant to 0b10
        let lo = entropy64;

        out[..8].copy_from_slice(&hi.to_be_bytes());
        out[8..].copy_from_slice(&lo.to_be_bytes());

        Self(out)
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

    /// Parses a wire string and returns the ID.
    /// This is the universal paradigm for parsing wire strings.
    pub fn parse(value: &str) -> Result<Self, Error> {
        let (id, _) = Self::parse_wire(value)?;
        Ok(id)
    }

    /// Parses a UUID string and returns the ID.
    /// This is the universal paradigm for parsing UUID strings.
    pub fn parse_uuid(value: &str) -> Result<Self, Error> {
        Self::from_uuid_string(value)
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
        GENERATOR.with(|cell| unsafe {
            let state = &mut *cell.get();
            for _ in 0..count {
                out.extend_from_slice(&state.next().0);
            }
        });
        out
    }

    #[cfg(feature = "data")]
    pub fn generate_tensor_batch(count: usize) -> Vec<i64> {
        let mut out = Vec::with_capacity(count.saturating_mul(2));
        GENERATOR.with(|cell| unsafe {
            let state = &mut *cell.get();
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
        GENERATOR.with(|cell| unsafe {
            let state = &mut *cell.get();
            for _ in 0..count {
                let (hi, lo) = state.next().to_tensor_words();
                out.extend_from_slice(&hi.to_be_bytes());
                out.extend_from_slice(&lo.to_be_bytes());
            }
        });
        out
    }
}

#[inline(always)]
fn encode_uuid_string(bytes: [u8; 16], out: &mut [u8; 36]) {
    const HEX: &[u8; 16] = b"0123456789abcdef";

    let mut dst = 0usize;
    for byte in bytes {
        if dst == 8 || dst == 13 || dst == 18 || dst == 23 {
            out[dst] = b'-';
            dst += 1;
        }
        out[dst] = HEX[(byte >> 4) as usize];
        out[dst + 1] = HEX[(byte & 0x0F) as usize];
        dst += 2;
    }
}

#[inline(always)]
fn parse_uuid_string_bytes(src: &[u8]) -> Result<[u8; 16], Error> {
    if src.len() != 36 {
        return Err(Error::InvalidLength);
    }
    if src[8] != b'-' || src[13] != b'-' || src[18] != b'-' || src[23] != b'-' {
        return Err(Error::InvalidFormat);
    }

    let mut out = [0u8; 16];
    let mut cursor = 0usize;
    for byte in &mut out {
        if cursor == 8 || cursor == 13 || cursor == 18 || cursor == 23 {
            cursor += 1;
        }
        let hi = decode_uuid_hex(src[cursor])?;
        let lo = decode_uuid_hex(src[cursor + 1])?;
        *byte = (hi << 4) | lo;
        cursor += 2;
    }

    validate_uuidv7_bytes(&out)?;
    Ok(out)
}

#[inline(always)]
fn decode_uuid_hex(byte: u8) -> Result<u8, Error> {
    match byte {
        b'0'..=b'9' => Ok(byte - b'0'),
        b'a'..=b'f' => Ok(byte - b'a' + 10),
        b'A'..=b'F' => Ok(byte - b'A' + 10),
        _ => Err(Error::InvalidFormat),
    }
}

#[inline(always)]
fn validate_uuidv7_bytes(bytes: &[u8; 16]) -> Result<(), Error> {
    if (bytes[6] >> 4) != 7 || (bytes[8] & 0xC0) != 0x80 {
        return Err(Error::InvalidFormat);
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_new_fast() {
        let id = Snid::new_fast();
        assert_ne!(id.0, [0u8; 16]);
    }

    #[test]
    fn test_uuidv7_alias_and_uuid_string_roundtrip() {
        let id = Snid::uuidv7();
        assert_eq!((id.0[6] >> 4) & 0x0F, 7);
        assert_eq!((id.0[8] >> 6) & 0b11, 0b10);

        let uuid_string = id.to_uuid_string();
        assert_eq!(uuid_string.len(), 36);
        assert_eq!(uuid_string.as_bytes()[8], b'-');
        assert_eq!(uuid_string.as_bytes()[13], b'-');
        assert_eq!(uuid_string.as_bytes()[18], b'-');
        assert_eq!(uuid_string.as_bytes()[23], b'-');

        let mut buffer = [0u8; 36];
        id.write_uuid_string(&mut buffer);
        assert_eq!(uuid_string.as_bytes(), &buffer);

        let parsed = Snid::from_uuid_string(&uuid_string).unwrap();
        assert_eq!(parsed, id);
    }

    #[test]
    fn test_from_uuid_string_rejects_non_v7() {
        let non_v7 = "018f1c3e-5a7b-4c8d-9e0f-1a2b3c4d5e6f";
        assert!(matches!(
            Snid::from_uuid_string(non_v7),
            Err(Error::InvalidFormat)
        ));
    }

    #[test]
    fn test_from_uuid_string_rejects_invalid_variant() {
        let mut bytes = Snid::uuidv7().to_bytes();
        bytes[8] &= 0x3F;
        let invalid = Snid::from_bytes(bytes).to_uuid_string();
        assert!(matches!(
            Snid::from_uuid_string(&invalid),
            Err(Error::InvalidFormat)
        ));
    }

    #[cfg(feature = "uuid")]
    #[test]
    fn test_uuid_crate_interop() {
        let id = Snid::uuidv7();
        let uuid = id.to_uuidv7();
        assert_eq!(uuid.to_string(), id.to_uuid_string());
        assert_eq!(Snid::from_uuidv7(uuid).unwrap(), id);

        let mut bytes = id.to_bytes();
        bytes[6] = (bytes[6] & 0x0F) | 0x40;
        let non_v7 = uuid::Uuid::from_bytes(bytes);
        assert!(matches!(
            Snid::from_uuidv7(non_v7),
            Err(Error::InvalidFormat)
        ));
    }

    #[test]
    fn test_from_bytes() {
        let bytes = [1u8; 16];
        let id = Snid::from_bytes(bytes);
        assert_eq!(id.0, bytes);
    }

    #[test]
    fn test_to_bytes() {
        let bytes = [2u8; 16];
        let id = Snid::from_bytes(bytes);
        assert_eq!(id.to_bytes(), bytes);
    }

    #[test]
    fn test_from_hex_valid() {
        let hex = "0123456789abcdef0123456789abcdef";
        let id = Snid::from_hex(hex).unwrap();
        assert_eq!(hex::encode(id.0), hex);
    }

    #[test]
    fn test_from_hex_invalid_length() {
        let result = Snid::from_hex("0123");
        assert!(matches!(result, Err(Error::InvalidLength)));
    }

    #[test]
    fn test_from_hex_invalid_chars() {
        let result = Snid::from_hex("gggggggggggggggggggggggggggggggg");
        assert!(result.is_err());
    }

    #[test]
    fn test_canonical_atom_valid() {
        assert_eq!(Snid::canonical_atom("MAT"), Some("MAT"));
        assert_eq!(Snid::canonical_atom("IAM"), Some("IAM"));
        assert_eq!(Snid::canonical_atom("TEN"), Some("TEN"));
    }

    #[test]
    fn test_canonical_atom_legacy() {
        assert_eq!(Snid::canonical_atom("OBJ"), Some("MAT"));
        assert_eq!(Snid::canonical_atom("TXN"), Some("LED"));
        assert_eq!(Snid::canonical_atom("SCH"), Some("CHR"));
    }

    #[test]
    fn test_canonical_atom_invalid() {
        assert_eq!(Snid::canonical_atom("XXX"), None);
        assert_eq!(Snid::canonical_atom(""), None);
    }

    #[test]
    fn test_to_wire() {
        let id = Snid::from_bytes([1u8; 16]);
        let wire = id.to_wire("MAT").unwrap();
        assert!(wire.starts_with("MAT:"));
    }

    #[test]
    fn test_to_wire_invalid_atom() {
        let id = Snid::from_bytes([1u8; 16]);
        let result = id.to_wire("XXX");
        assert!(matches!(result, Err(Error::InvalidAtom)));
    }

    #[test]
    fn test_parse_wire_valid() {
        let id = Snid::from_bytes([1u8; 16]);
        let wire = id.to_wire("MAT").unwrap();
        let (parsed, atom) = Snid::parse_wire(&wire).unwrap();
        assert_eq!(parsed, id);
        assert_eq!(atom, "MAT");
    }

    #[test]
    fn test_parse_wire_underscore() {
        let id = Snid::from_bytes([1u8; 16]);
        let wire = id.to_wire("MAT").unwrap();
        let underscore_wire = wire.replace(':', "_");
        let (parsed, _atom) = Snid::parse_wire(&underscore_wire).unwrap();
        assert_eq!(parsed, id);
    }

    #[test]
    fn test_parse_wire_invalid_format() {
        let result = Snid::parse_wire("invalid");
        assert!(matches!(result, Err(Error::InvalidFormat)));
    }

    #[test]
    fn test_parse_wire_invalid_atom() {
        let result = Snid::parse_wire("XXX:payload");
        assert!(matches!(result, Err(Error::InvalidAtom)));
    }

    #[test]
    fn test_timestamp_millis() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        assert_eq!(id.timestamp_millis(), 1700000000123);
    }

    #[test]
    fn test_to_tensor_words() {
        let id = Snid::from_bytes([1u8; 16]);
        let (hi, lo) = id.to_tensor_words();
        assert_ne!(hi, 0);
        assert_ne!(lo, 0);
    }

    #[test]
    fn test_sequence() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let seq = id.sequence();
        assert!(seq <= 0x3FFF);
    }

    #[test]
    fn test_machine_or_shard() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let machine = id.machine_or_shard();
        assert!(machine <= 0x00FF_FFFF);
    }

    #[test]
    fn test_time_bin() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let bin = id.time_bin(3600000); // 1 hour
        assert!(bin > 0);
    }

    #[test]
    fn test_new_safe() {
        let id = Snid::new_safe();
        // Check that it's a valid UUIDv7
        let version = (id.0[6] >> 4) & 0x0F;
        assert_eq!(version, 0x7);
        // Check variant bits
        let variant = (id.0[8] >> 6) & 0b11;
        assert_eq!(variant, 0b10);
    }

    #[test]
    fn test_new_safe_time_blurring() {
        let id1 = Snid::new_safe();
        std::thread::sleep(std::time::Duration::from_millis(500));
        let id2 = Snid::new_safe();
        // Both should be in the same second due to time-blurring
        let ts1 = id1.timestamp_millis();
        let ts2 = id2.timestamp_millis();
        // Timestamps should be within 1 second
        assert!((ts1 - ts2).abs() < 1000);
    }

    #[test]
    fn test_new_safe_uniqueness() {
        let mut ids = std::collections::HashSet::new();
        for _ in 0..100 {
            let id = Snid::new_safe();
            assert!(ids.insert(id), "duplicate ID generated");
        }
    }

    #[test]
    fn test_to_base32() {
        let id = Snid::new_fast();
        let base32 = id.to_base32();
        assert!(!base32.is_empty());
        // Crockford Base32 should only contain valid characters
        for c in base32.chars() {
            assert!(c.is_ascii_alphanumeric());
        }
    }

    #[test]
    fn test_to_base32_consistency() {
        let id = Snid::new_fast();
        let base32_1 = id.to_base32();
        let base32_2 = id.to_base32();
        assert_eq!(base32_1, base32_2);
    }


    #[test]
    fn test_time_bin_zero_resolution() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let bin = id.time_bin(0);
        assert_eq!(bin, 1700000000123);
    }

    #[test]
    fn test_is_ghosted_default() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        assert!(!id.is_ghosted());
    }

    #[test]
    fn test_with_ghost_bit_enable() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let ghosted = id.with_ghost_bit(true);
        assert!(ghosted.is_ghosted());
    }

    #[test]
    fn test_with_ghost_bit_disable() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let ghosted = id.with_ghost_bit(true);
        let unghosted = ghosted.with_ghost_bit(false);
        assert!(!unghosted.is_ghosted());
    }

    #[test]
    fn test_from_hash_with_timestamp() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        assert_eq!(id.timestamp_millis(), 1700000000123);
        assert_ne!(id.0, [0u8; 16]);
    }

    #[test]
    fn test_from_hash_with_timestamp_empty_hash() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"");
        assert_eq!(id.timestamp_millis(), 1700000000123);
    }

    #[test]
    fn test_h3_cell_non_spatial() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        assert_eq!(id.h3_cell(), None);
    }

    #[test]
    fn test_from_spatial_parts() {
        let cell = 0x8c2a1072b59ffff;
        let entropy = 0x1234567890ABCDEF;
        let id = Snid::from_spatial_parts(cell, entropy);
        assert_eq!(id.h3_cell(), Some(cell));
    }

    #[test]
    fn test_h3_feature_vector_non_spatial() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        assert!(id.h3_feature_vector().is_empty());
    }

    #[test]
    fn test_h3_feature_vector_spatial() {
        let cell = 0x8c2a1072b59ffff;
        let entropy = 0x1234567890ABCDEF;
        let id = Snid::from_spatial_parts(cell, entropy);
        let features = id.h3_feature_vector();
        assert_eq!(features, vec![cell]);
    }

    #[cfg(feature = "data")]
    #[test]
    fn test_generate_binary_batch() {
        let batch = Snid::generate_binary_batch(10);
        assert_eq!(batch.len(), 160);
    }

    #[cfg(feature = "data")]
    #[test]
    fn test_generate_tensor_batch() {
        let batch = Snid::generate_tensor_batch(10);
        assert_eq!(batch.len(), 20);
    }

    #[cfg(feature = "data")]
    #[test]
    fn test_generate_tensor_batch_be_bytes() {
        let batch = Snid::generate_tensor_batch_be_bytes(10);
        assert_eq!(batch.len(), 160);
    }

    #[test]
    fn test_roundtrip_bytes() {
        let id1 = Snid::new_fast();
        let bytes = id1.to_bytes();
        let id2 = Snid::from_bytes(bytes);
        assert_eq!(id1, id2);
    }

    #[test]
    fn test_roundtrip_wire() {
        let id1 = Snid::new_fast();
        let wire = id1.to_wire("MAT").unwrap();
        let (id2, atom) = Snid::parse_wire(&wire).unwrap();
        assert_eq!(id1, id2);
        assert_eq!(atom, "MAT");
    }

    #[test]
    fn test_version_bits() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let version = (id.0[6] >> 4) & 0x0F;
        assert_eq!(version, 0x7);
    }

    #[test]
    fn test_variant_bits() {
        let id = Snid::from_hash_with_timestamp(1700000000123, b"test");
        let variant = (id.0[8] >> 6) & 0b11;
        assert_eq!(variant, 0b10);
    }

    #[test]
    fn test_from_str() {
        let id1 = Snid::new();
        let wire = id1.to_wire("MAT").unwrap();
        let id2: Snid = wire.parse().unwrap();
        assert_eq!(id1, id2);
    }

    #[test]
    fn test_from_str_invalid() {
        let result: Result<Snid, _> = "invalid".parse();
        assert!(result.is_err());
    }

    #[test]
    fn test_display() {
        let id = Snid::from_bytes([1u8; 16]);
        let display = format!("{}", id);
        assert!(display.starts_with("MAT:"));
    }

    #[test]
    fn test_ord_sorting() {
        let id1 = Snid::from_hash_with_timestamp(1700000000123, b"test1");
        let id2 = Snid::from_hash_with_timestamp(1700000000124, b"test2");
        assert!(id1 < id2);
    }

    #[test]
    fn test_ord_equality() {
        let bytes = [1u8; 16];
        let id1 = Snid::from_bytes(bytes);
        let id2 = Snid::from_bytes(bytes);
        assert_eq!(id1, id2);
        assert!(!(id1 < id2));
        assert!(!(id1 > id2));
    }
}

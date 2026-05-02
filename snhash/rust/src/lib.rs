//! SNHASH: content integrity, HASHID, evidence manifests, CAS, and API proof helpers.
//!
//! SNHASH identifies bytes. SNID identifies topology and time. The optional
//! `snid-bridge` feature attaches BLAKE3 content digests to SNID `Bid` values.

use data_encoding::BASE32_NOPAD;
use hmac::{Hmac, KeyInit, Mac};
use serde::de::{self, Deserializer};
use serde::ser::{SerializeStruct, Serializer};
use serde::{Deserialize, Serialize};
use sha1::Sha1;
use sha2::{Digest as ShaDigest, Sha256};
use sha3::digest::{ExtendableOutput, Update, XofReader};
use sha3::{Sha3_256, Shake256};
use std::fmt;
use std::fs;
use std::hash::Hasher as _;
use std::io::{self, Read};
use std::path::{Path, PathBuf};
use twox_hash::XxHash3_64;

type HmacSha256 = Hmac<Sha256>;

const WIRE_VERSION: &str = "h1";
const HASH_ID_VERSION: &str = "hid1";
const MANIFEST_VERSION: &str = "snhash-manifest-v1";

/// SNHASH errors.
#[derive(Debug)]
pub enum Error {
    UnknownAlgorithm,
    InvalidDigest,
    InvalidWire,
    InvalidHashId,
    DigestNotFound,
    InvalidKey,
    Io(io::Error),
    Json(serde_json::Error),
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Error::UnknownAlgorithm => write!(f, "unknown algorithm"),
            Error::InvalidDigest => write!(f, "invalid digest"),
            Error::InvalidWire => write!(f, "invalid wire format"),
            Error::InvalidHashId => write!(f, "invalid hash id"),
            Error::DigestNotFound => write!(f, "digest not found"),
            Error::InvalidKey => write!(f, "invalid key"),
            Error::Io(err) => write!(f, "io error: {err}"),
            Error::Json(err) => write!(f, "json error: {err}"),
        }
    }
}

impl std::error::Error for Error {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            Error::Io(err) => Some(err),
            Error::Json(err) => Some(err),
            _ => None,
        }
    }
}

impl From<io::Error> for Error {
    fn from(value: io::Error) -> Self {
        Self::Io(value)
    }
}

impl From<serde_json::Error> for Error {
    fn from(value: serde_json::Error) -> Self {
        Self::Json(value)
    }
}

/// Supported digest algorithms.
#[allow(non_camel_case_types)]
#[derive(Clone, Copy, Debug, PartialEq, Eq, PartialOrd, Ord, Hash)]
pub enum Algorithm {
    Blake3_256,
    Blake3Keyed_256,
    Sha256,
    HmacSha256,
    Sha3_256,
    Shake256_256,
    Xxh3_64,
    Md5,
    Sha1,
}

impl Algorithm {
    /// Returns the canonical wire name.
    pub fn as_str(self) -> &'static str {
        match self {
            Algorithm::Blake3_256 => "blake3-256",
            Algorithm::Blake3Keyed_256 => "blake3-keyed-256",
            Algorithm::Sha256 => "sha-256",
            Algorithm::HmacSha256 => "hmac-sha-256",
            Algorithm::Sha3_256 => "sha3-256",
            Algorithm::Shake256_256 => "shake256-256",
            Algorithm::Xxh3_64 => "xxh3-64",
            Algorithm::Md5 => "md5",
            Algorithm::Sha1 => "sha-1",
        }
    }

    /// Returns true when a digest can be used for current integrity verification.
    pub fn trusted(self) -> bool {
        matches!(
            self,
            Algorithm::Blake3_256
                | Algorithm::Blake3Keyed_256
                | Algorithm::Sha256
                | Algorithm::HmacSha256
                | Algorithm::Sha3_256
                | Algorithm::Shake256_256
        )
    }

    fn validate_size(self, size: usize) -> Result<(), Error> {
        let ok = match self {
            Algorithm::Blake3_256
            | Algorithm::Blake3Keyed_256
            | Algorithm::Sha256
            | Algorithm::HmacSha256
            | Algorithm::Sha3_256
            | Algorithm::Shake256_256 => size == 32,
            Algorithm::Xxh3_64 => size == 8,
            Algorithm::Md5 => size == 16,
            Algorithm::Sha1 => size == 20,
        };
        if ok {
            Ok(())
        } else {
            Err(Error::InvalidDigest)
        }
    }
}

impl fmt::Display for Algorithm {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.write_str(self.as_str())
    }
}

impl std::str::FromStr for Algorithm {
    type Err = Error;

    fn from_str(value: &str) -> Result<Self, Self::Err> {
        match value {
            "blake3-256" => Ok(Algorithm::Blake3_256),
            "blake3-keyed-256" => Ok(Algorithm::Blake3Keyed_256),
            "sha-256" => Ok(Algorithm::Sha256),
            "hmac-sha-256" => Ok(Algorithm::HmacSha256),
            "sha3-256" => Ok(Algorithm::Sha3_256),
            "shake256-256" => Ok(Algorithm::Shake256_256),
            "xxh3-64" => Ok(Algorithm::Xxh3_64),
            "md5" => Ok(Algorithm::Md5),
            "sha-1" => Ok(Algorithm::Sha1),
            _ => Err(Error::UnknownAlgorithm),
        }
    }
}

impl Serialize for Algorithm {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
        serializer.serialize_str(self.as_str())
    }
}

impl<'de> Deserialize<'de> for Algorithm {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: Deserializer<'de>,
    {
        let value = String::deserialize(deserializer)?;
        value.parse().map_err(de::Error::custom)
    }
}

/// Digest profile.
#[derive(Clone, Copy, Debug, PartialEq, Eq, PartialOrd, Ord, Hash, Serialize, Deserialize)]
pub enum Profile {
    FastCAS,
    Evidence,
    API,
    FIPS,
    Cache,
}

impl Profile {
    /// Returns the profile algorithms.
    pub fn algorithms(self) -> &'static [Algorithm] {
        match self {
            Profile::FastCAS => &[Algorithm::Blake3_256],
            Profile::Evidence => &[Algorithm::Blake3_256, Algorithm::Sha256],
            Profile::API => &[Algorithm::Sha256, Algorithm::Blake3_256],
            Profile::FIPS => &[
                Algorithm::Sha256,
                Algorithm::Sha3_256,
                Algorithm::Shake256_256,
            ],
            Profile::Cache => &[Algorithm::Xxh3_64, Algorithm::Blake3_256],
        }
    }
}

impl Default for Profile {
    fn default() -> Self {
        Self::FastCAS
    }
}

/// Self-describing digest.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Digest {
    pub algorithm: Algorithm,
    pub size: usize,
    pub bytes: Vec<u8>,
}

impl Digest {
    /// Creates a validated digest.
    pub fn new(algorithm: Algorithm, bytes: impl AsRef<[u8]>) -> Result<Self, Error> {
        let bytes = bytes.as_ref();
        algorithm.validate_size(bytes.len())?;
        Ok(Self {
            algorithm,
            size: bytes.len(),
            bytes: bytes.to_vec(),
        })
    }

    /// Lower-case hex.
    pub fn hex(&self) -> String {
        hex::encode(&self.bytes)
    }

    /// Lower-case RFC 4648 base32 without padding.
    pub fn base32(&self) -> String {
        BASE32_NOPAD.encode(&self.bytes).to_ascii_lowercase()
    }

    /// Digest wire: h1:<algorithm>:<base32-lower-no-padding>.
    pub fn wire(&self) -> String {
        format!("{WIRE_VERSION}:{}:{}", self.algorithm, self.base32())
    }

    /// True when algorithm is trusted for current verification.
    pub fn trusted(&self) -> bool {
        self.algorithm.trusted()
    }

    /// Hashes bytes and checks equality.
    pub fn verify(&self, data: &[u8]) -> bool {
        let Ok(manifest) = hash_bytes_with_algorithms(data, &[self.algorithm]) else {
            return false;
        };
        manifest
            .digest(self.algorithm)
            .is_some_and(|got| equal_bytes(&got.bytes, &self.bytes))
    }

    /// Hashes a reader and checks equality.
    pub fn verify_reader<R: Read>(&self, reader: R) -> Result<bool, Error> {
        let manifest = hash_reader_with_algorithms(reader, &[self.algorithm])?;
        Ok(manifest
            .digest(self.algorithm)
            .is_some_and(|got| equal_bytes(&got.bytes, &self.bytes)))
    }
}

impl Serialize for Digest {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
        let mut state = serializer.serialize_struct("Digest", 6)?;
        state.serialize_field("algorithm", &self.algorithm)?;
        state.serialize_field("size", &self.size)?;
        state.serialize_field("hex", &self.hex())?;
        state.serialize_field("base32", &self.base32())?;
        state.serialize_field("wire", &self.wire())?;
        state.serialize_field("trusted", &self.trusted())?;
        state.end()
    }
}

impl<'de> Deserialize<'de> for Digest {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: Deserializer<'de>,
    {
        #[derive(Deserialize)]
        struct DigestJson {
            algorithm: Option<Algorithm>,
            size: Option<usize>,
            hex: Option<String>,
            wire: Option<String>,
        }

        let value = DigestJson::deserialize(deserializer)?;
        if let Some(wire) = value.wire {
            return parse_digest_wire(&wire).map_err(de::Error::custom);
        }
        let algorithm = value
            .algorithm
            .ok_or_else(|| de::Error::custom("missing algorithm"))?;
        let hex = value.hex.ok_or_else(|| de::Error::custom("missing hex"))?;
        let bytes = hex::decode(hex).map_err(de::Error::custom)?;
        let digest = Digest::new(algorithm, bytes).map_err(de::Error::custom)?;
        if value.size.is_some_and(|size| size != digest.size) {
            return Err(de::Error::custom("size mismatch"));
        }
        Ok(digest)
    }
}

/// Parses digest wire.
pub fn parse_digest_wire(wire: &str) -> Result<Digest, Error> {
    let mut parts = wire.split(':');
    if parts.next() != Some(WIRE_VERSION) {
        return Err(Error::InvalidWire);
    }
    let algorithm: Algorithm = parts.next().ok_or(Error::InvalidWire)?.parse()?;
    let encoded = parts.next().ok_or(Error::InvalidWire)?;
    if parts.next().is_some() {
        return Err(Error::InvalidWire);
    }
    let bytes = BASE32_NOPAD
        .decode(encoded.to_ascii_uppercase().as_bytes())
        .map_err(|_| Error::InvalidWire)?;
    Digest::new(algorithm, bytes)
}

/// Compact content identifier.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct HashId {
    pub profile: Profile,
    pub size: u64,
    pub root: Digest,
}

impl HashId {
    /// Creates a HASHID from a trusted digest.
    pub fn new(profile: Profile, size: u64, root: Digest) -> Result<Self, Error> {
        if !root.trusted() {
            return Err(Error::InvalidHashId);
        }
        Ok(Self {
            profile,
            size,
            root: Digest::new(root.algorithm, &root.bytes)?,
        })
    }

    /// HASHID wire: hid1:<profile>:<size>:<algorithm>:<base32>.
    pub fn wire(&self) -> String {
        format!(
            "{HASH_ID_VERSION}:{:?}:{}:{}:{}",
            self.profile,
            self.size,
            self.root.algorithm,
            self.root.base32()
        )
    }

    /// Checks the HASHID against bytes.
    pub fn verify(&self, data: &[u8]) -> bool {
        data.len() as u64 == self.size && self.root.verify(data)
    }
}

impl fmt::Display for HashId {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.write_str(&self.wire())
    }
}

/// Parses HASHID wire.
pub fn parse_hash_id(wire: &str) -> Result<HashId, Error> {
    let mut parts = wire.split(':');
    if parts.next() != Some(HASH_ID_VERSION) {
        return Err(Error::InvalidHashId);
    }
    let profile = match parts.next().ok_or(Error::InvalidHashId)? {
        "FastCAS" => Profile::FastCAS,
        "Evidence" => Profile::Evidence,
        "API" => Profile::API,
        "FIPS" => Profile::FIPS,
        "Cache" => Profile::Cache,
        _ => return Err(Error::InvalidHashId),
    };
    let size = parts
        .next()
        .ok_or(Error::InvalidHashId)?
        .parse::<u64>()
        .map_err(|_| Error::InvalidHashId)?;
    let algorithm = parts.next().ok_or(Error::InvalidHashId)?;
    let encoded = parts.next().ok_or(Error::InvalidHashId)?;
    if parts.next().is_some() {
        return Err(Error::InvalidHashId);
    }
    let root = parse_digest_wire(&format!("{WIRE_VERSION}:{algorithm}:{encoded}"))?;
    HashId::new(profile, size, root)
}

/// Hashing options.
#[derive(Clone, Debug)]
pub struct Options {
    pub profile: Profile,
    pub algorithms: Vec<Algorithm>,
    pub media_type: String,
    pub canonicalization: String,
    pub source_name: Option<String>,
    pub custody: Vec<CustodyEvent>,
    pub chunk_size: Option<usize>,
    pub created_at: String,
}

impl Default for Options {
    fn default() -> Self {
        Self {
            profile: Profile::FastCAS,
            algorithms: Vec::new(),
            media_type: "application/octet-stream".to_owned(),
            canonicalization: "raw-bytes".to_owned(),
            source_name: None,
            custody: Vec::new(),
            chunk_size: None,
            created_at: "1970-01-01T00:00:00Z".to_owned(),
        }
    }
}

impl Options {
    /// Creates options for a profile.
    pub fn profile(profile: Profile) -> Self {
        Self {
            profile,
            ..Self::default()
        }
    }

    /// Sets exact algorithms.
    pub fn algorithms(mut self, algorithms: impl Into<Vec<Algorithm>>) -> Self {
        self.algorithms = algorithms.into();
        self
    }

    /// Sets chunk size. Zero disables chunks.
    pub fn chunk_size(mut self, chunk_size: usize) -> Self {
        self.chunk_size = (chunk_size > 0).then_some(chunk_size);
        self
    }

    /// Sets creation time as an RFC3339 string.
    pub fn created_at(mut self, created_at: impl Into<String>) -> Self {
        self.created_at = created_at.into();
        self
    }

    /// Sets source name.
    pub fn source_name(mut self, source_name: impl Into<String>) -> Self {
        self.source_name = Some(source_name.into());
        self
    }

    /// Sets media type.
    pub fn media_type(mut self, media_type: impl Into<String>) -> Self {
        self.media_type = media_type.into();
        self
    }

    /// Sets canonicalization contract.
    pub fn canonicalization(mut self, canonicalization: impl Into<String>) -> Self {
        self.canonicalization = canonicalization.into();
        self
    }

    /// Appends a custody event.
    pub fn custody(mut self, event: CustodyEvent) -> Self {
        self.custody.push(event);
        self
    }

    fn normalized_algorithms(&self) -> Vec<Algorithm> {
        let mut algorithms = if self.algorithms.is_empty() {
            self.profile.algorithms().to_vec()
        } else {
            self.algorithms.clone()
        };
        normalize_algorithms(&mut algorithms);
        algorithms
    }
}

/// Manifest proof envelope.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct Manifest {
    pub version: String,
    pub profile: Profile,
    pub size: u64,
    pub media_type: String,
    pub canonicalization: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub source_name: Option<String>,
    pub digests: Vec<Digest>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub tree_root: Option<Digest>,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub chunks: Vec<Chunk>,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub entries: Vec<Entry>,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub custody: Vec<CustodyEvent>,
    pub created_at: String,
}

impl Manifest {
    /// Returns a digest by algorithm.
    pub fn digest(&self, algorithm: Algorithm) -> Option<&Digest> {
        self.digests
            .iter()
            .find(|digest| digest.algorithm == algorithm)
    }

    /// Returns the BLAKE3 digest.
    pub fn blake3(&self) -> Result<&Digest, Error> {
        self.digest(Algorithm::Blake3_256)
            .ok_or(Error::DigestNotFound)
    }

    /// Returns the SHA-256 digest.
    pub fn sha256(&self) -> Result<&Digest, Error> {
        self.digest(Algorithm::Sha256).ok_or(Error::DigestNotFound)
    }

    /// Derives HASHID from the strongest available digest.
    pub fn hash_id(&self) -> Result<HashId, Error> {
        for algorithm in [
            Algorithm::Blake3_256,
            Algorithm::Sha256,
            Algorithm::Sha3_256,
            Algorithm::Shake256_256,
        ] {
            if let Some(digest) = self.digest(algorithm) {
                return HashId::new(self.profile, self.size, digest.clone());
            }
        }
        if let Some(tree_root) = &self.tree_root {
            if tree_root.trusted() {
                return HashId::new(self.profile, self.size, tree_root.clone());
            }
        }
        Err(Error::DigestNotFound)
    }

    /// Canonical JSON bytes.
    pub fn canonical_json(&self) -> Result<Vec<u8>, Error> {
        Ok(serde_json::to_vec(&self.normalized())?)
    }

    /// BLAKE3-256 over canonical manifest JSON.
    pub fn proof_root(&self) -> Result<Digest, Error> {
        let canonical = self.canonical_json()?;
        Digest::new(Algorithm::Blake3_256, blake3::hash(&canonical).as_bytes())
    }

    /// Verifies all trusted digests against bytes.
    pub fn verify_bytes(&self, data: &[u8]) -> bool {
        self.digests
            .iter()
            .filter(|digest| digest.trusted())
            .all(|digest| digest.verify(data))
            && self.digests.iter().any(|digest| digest.trusted())
    }

    /// Verifies all trusted digests against a reader.
    pub fn verify_reader<R: Read>(&self, reader: R) -> Result<bool, Error> {
        let algorithms = trusted_algorithms(&self.digests);
        if algorithms.is_empty() {
            return Ok(false);
        }
        let actual = hash_reader_with_algorithms(reader, &algorithms)?;
        if actual.size != self.size {
            return Ok(false);
        }
        Ok(self
            .digests
            .iter()
            .filter(|digest| digest.trusted())
            .all(|expected| {
                actual
                    .digest(expected.algorithm)
                    .is_some_and(|got| equal_bytes(&got.bytes, &expected.bytes))
            }))
    }

    /// Verifies one chunk against per-chunk trusted digests.
    pub fn verify_chunk(&self, index: usize, data: &[u8]) -> bool {
        let Some(chunk) = self.chunks.get(index) else {
            return false;
        };
        if chunk.size != data.len() as u64 {
            return false;
        }
        let algorithms = trusted_algorithms(&chunk.digests);
        if algorithms.is_empty() {
            return false;
        }
        let Ok(actual) = hash_bytes_with_algorithms(data, &algorithms) else {
            return false;
        };
        chunk
            .digests
            .iter()
            .filter(|digest| digest.trusted())
            .all(|expected| {
                actual
                    .digest(expected.algorithm)
                    .is_some_and(|got| equal_bytes(&got.bytes, &expected.bytes))
            })
    }

    /// Writes canonical JSON to disk.
    pub fn write_file(&self, path: impl AsRef<Path>) -> Result<(), Error> {
        fs::write(path, self.canonical_json()?)?;
        Ok(())
    }

    fn normalized(&self) -> Self {
        let mut out = self.clone();
        if out.version.is_empty() {
            out.version = MANIFEST_VERSION.to_owned();
        }
        normalize_digests(&mut out.digests);
        for chunk in &mut out.chunks {
            normalize_digests(&mut chunk.digests);
        }
        out.chunks
            .sort_by_key(|chunk| (chunk.index, chunk.offset, chunk.size));
        for entry in &mut out.entries {
            normalize_digests(&mut entry.digests);
        }
        out.entries.sort_by(|a, b| a.path.cmp(&b.path));
        out
    }
}

/// Per-chunk digest metadata.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct Chunk {
    pub index: usize,
    pub offset: u64,
    pub size: u64,
    pub digests: Vec<Digest>,
}

/// Directory manifest entry.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct Entry {
    pub path: String,
    pub kind: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub mode: Option<String>,
    pub size: u64,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub digests: Vec<Digest>,
}

/// Optional chain-of-custody event.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct CustodyEvent {
    pub at: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub actor: Option<String>,
    pub action: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub note: Option<String>,
}

/// Hashes bytes with default FastCAS settings.
pub fn hash_bytes(data: &[u8]) -> Result<Manifest, Error> {
    hash_bytes_with_options(data, Options::default())
}

/// Hashes bytes with a profile.
pub fn hash_bytes_profile(data: &[u8], profile: Profile) -> Result<Manifest, Error> {
    hash_bytes_with_options(data, Options::profile(profile))
}

/// Hashes bytes with exact algorithms.
pub fn hash_bytes_with_algorithms(
    data: &[u8],
    algorithms: &[Algorithm],
) -> Result<Manifest, Error> {
    hash_bytes_with_options(data, Options::default().algorithms(algorithms.to_vec()))
}

/// Hashes bytes with options.
pub fn hash_bytes_with_options(data: &[u8], options: Options) -> Result<Manifest, Error> {
    let algorithms = options.normalized_algorithms();
    if algorithms == [Algorithm::Blake3_256] && options.chunk_size.is_none() {
        let digest = Digest::new(Algorithm::Blake3_256, blake3::hash(data).as_bytes())?;
        return manifest_from_parts(data.len() as u64, vec![digest], Vec::new(), None, options);
    }
    if algorithms == [Algorithm::Xxh3_64] && options.chunk_size.is_none() {
        let digest = Digest::new(Algorithm::Xxh3_64, xxh3_64_bytes(data))?;
        return manifest_from_parts(data.len() as u64, vec![digest], Vec::new(), None, options);
    }
    if algorithms == [Algorithm::Blake3_256, Algorithm::Xxh3_64] && options.chunk_size.is_none() {
        let digests = vec![
            Digest::new(Algorithm::Blake3_256, blake3::hash(data).as_bytes())?,
            Digest::new(Algorithm::Xxh3_64, xxh3_64_bytes(data))?,
        ];
        return manifest_from_parts(data.len() as u64, digests, Vec::new(), None, options);
    }

    let mut states = new_states(&algorithms);
    write_states(&mut states, data);
    let digests = digests_from_states(states)?;
    let chunks = if let Some(chunk_size) = options.chunk_size {
        hash_chunks(data, chunk_size, &algorithms)?
    } else {
        Vec::new()
    };
    let tree_root = tree_root_from_chunks(&chunks)?;
    manifest_from_parts(data.len() as u64, digests, chunks, tree_root, options)
}

/// Hashes a reader with default FastCAS settings.
pub fn hash_reader<R: Read>(reader: R) -> Result<Manifest, Error> {
    hash_reader_with_options(reader, Options::default())
}

/// Hashes a reader with exact algorithms.
pub fn hash_reader_with_algorithms<R: Read>(
    reader: R,
    algorithms: &[Algorithm],
) -> Result<Manifest, Error> {
    hash_reader_with_options(reader, Options::default().algorithms(algorithms.to_vec()))
}

/// Hashes a reader with options.
pub fn hash_reader_with_options<R: Read>(
    mut reader: R,
    options: Options,
) -> Result<Manifest, Error> {
    let algorithms = options.normalized_algorithms();
    let mut states = new_states(&algorithms);
    let mut chunks = Vec::new();
    let mut total = 0u64;

    if let Some(chunk_size) = options.chunk_size {
        let mut buf = vec![0u8; chunk_size.max(1)];
        loop {
            let n = reader.read(&mut buf)?;
            if n == 0 {
                break;
            }
            let part = &buf[..n];
            write_states(&mut states, part);
            chunks.push(hash_chunk(chunks.len(), total, part, &algorithms)?);
            total += n as u64;
        }
    } else {
        let mut buf = [0u8; 128 * 1024];
        loop {
            let n = reader.read(&mut buf)?;
            if n == 0 {
                break;
            }
            write_states(&mut states, &buf[..n]);
            total += n as u64;
        }
    }

    let digests = digests_from_states(states)?;
    let tree_root = tree_root_from_chunks(&chunks)?;
    manifest_from_parts(total, digests, chunks, tree_root, options)
}

/// Hashes a file.
pub fn hash_file(path: impl AsRef<Path>, mut options: Options) -> Result<Manifest, Error> {
    let path = path.as_ref();
    if options.source_name.is_none() {
        options.source_name = Some(path.to_string_lossy().into_owned());
    }
    let file = fs::File::open(path)?;
    hash_reader_with_options(file, options)
}

/// Hashes a directory deterministically.
pub fn hash_dir(root: impl AsRef<Path>, mut options: Options) -> Result<Manifest, Error> {
    let root = root.as_ref();
    if options.source_name.is_none() {
        options.source_name = Some(root.to_string_lossy().into_owned());
    }
    if options.canonicalization == "raw-bytes" {
        options.canonicalization = "snhash-dir-v1:path-sort-mode-size-digest".to_owned();
    }
    let algorithms = options.normalized_algorithms();
    let mut entries = Vec::new();
    collect_dir_entries(root, root, &algorithms, &options, &mut entries)?;
    entries.sort_by(|a, b| a.path.cmp(&b.path));
    let canonical = directory_canonical_bytes(&entries);
    let mut manifest = hash_bytes_with_options(
        &canonical,
        Options {
            media_type: "application/vnd.snhash.directory+json".to_owned(),
            ..options
        },
    )?;
    manifest.size = canonical.len() as u64;
    manifest.entries = entries;
    Ok(manifest)
}

/// Reads a manifest JSON file.
pub fn read_manifest_file(path: impl AsRef<Path>) -> Result<Manifest, Error> {
    let data = fs::read(path)?;
    Ok(serde_json::from_slice::<Manifest>(&data)?.normalized())
}

/// Parses manifest JSON.
pub fn parse_manifest_json(data: &[u8]) -> Result<Manifest, Error> {
    Ok(serde_json::from_slice::<Manifest>(data)?.normalized())
}

/// HMAC-SHA256 for public API compatibility.
pub fn hmac_sha256(key: &[u8], data: &[u8]) -> Result<Digest, Error> {
    let mut mac = HmacSha256::new_from_slice(key).map_err(|_| Error::InvalidKey)?;
    Mac::update(&mut mac, data);
    Digest::new(Algorithm::HmacSha256, mac.finalize().into_bytes())
}

/// Verifies HMAC-SHA256.
pub fn verify_hmac_sha256(key: &[u8], data: &[u8], expected: &Digest) -> bool {
    expected.algorithm == Algorithm::HmacSha256
        && hmac_sha256(key, data)
            .map(|actual| equal_bytes(&actual.bytes, &expected.bytes))
            .unwrap_or(false)
}

/// Keyed BLAKE3 for fast internal service calls.
pub fn keyed_blake3(key: &[u8; 32], data: &[u8]) -> Result<Digest, Error> {
    let mut hasher = blake3::Hasher::new_keyed(key);
    hasher.update(data);
    Digest::new(Algorithm::Blake3Keyed_256, hasher.finalize().as_bytes())
}

/// Verifies keyed BLAKE3.
pub fn verify_keyed_blake3(key: &[u8; 32], data: &[u8], expected: &Digest) -> bool {
    expected.algorithm == Algorithm::Blake3Keyed_256
        && keyed_blake3(key, data)
            .map(|actual| equal_bytes(&actual.bytes, &expected.bytes))
            .unwrap_or(false)
}

#[cfg(feature = "snid-bridge")]
/// Converts a BLAKE3 digest into a SNID BID.
pub fn new_bid_from_digest(digest: &Digest) -> Result<snid::Bid, Error> {
    if digest.algorithm != Algorithm::Blake3_256 || digest.bytes.len() != 32 {
        return Err(Error::DigestNotFound);
    }
    let mut content = [0u8; 32];
    content.copy_from_slice(&digest.bytes);
    Ok(snid::Bid::from_parts(snid::Snid::new_fast(), content))
}

#[cfg(feature = "snid-bridge")]
/// Attaches a SNID BID to a manifest BLAKE3 content digest.
pub fn attach_snid(manifest: &Manifest) -> Result<snid::Bid, Error> {
    new_bid_from_digest(manifest.blake3()?)
}

enum DigestState {
    Blake3(blake3::Hasher),
    Sha256(Sha256),
    Sha3_256(Sha3_256),
    Shake256(Shake256),
    Xxh3_64(XxHash3_64),
    Md5(md5::Md5),
    Sha1(Sha1),
}

impl DigestState {
    fn new(algorithm: Algorithm) -> Result<Self, Error> {
        Ok(match algorithm {
            Algorithm::Blake3_256 => Self::Blake3(blake3::Hasher::new()),
            Algorithm::Sha256 => Self::Sha256(Sha256::new()),
            Algorithm::Sha3_256 => Self::Sha3_256(Sha3_256::new()),
            Algorithm::Shake256_256 => Self::Shake256(Shake256::default()),
            Algorithm::Xxh3_64 => Self::Xxh3_64(XxHash3_64::default()),
            Algorithm::Md5 => Self::Md5(md5::Md5::new()),
            Algorithm::Sha1 => Self::Sha1(Sha1::new()),
            Algorithm::Blake3Keyed_256 | Algorithm::HmacSha256 => {
                return Err(Error::UnknownAlgorithm);
            }
        })
    }

    fn update(&mut self, data: &[u8]) {
        match self {
            Self::Blake3(hasher) => {
                hasher.update(data);
            }
            Self::Sha256(hasher) => ShaDigest::update(hasher, data),
            Self::Sha3_256(hasher) => ShaDigest::update(hasher, data),
            Self::Shake256(hasher) => Update::update(hasher, data),
            Self::Xxh3_64(hasher) => hasher.write(data),
            Self::Md5(hasher) => ShaDigest::update(hasher, data),
            Self::Sha1(hasher) => ShaDigest::update(hasher, data),
        }
    }

    fn finalize(self, algorithm: Algorithm) -> Result<Digest, Error> {
        match self {
            Self::Blake3(hasher) => Digest::new(algorithm, hasher.finalize().as_bytes()),
            Self::Sha256(hasher) => Digest::new(algorithm, hasher.finalize()),
            Self::Sha3_256(hasher) => Digest::new(algorithm, hasher.finalize()),
            Self::Shake256(hasher) => {
                let mut reader = hasher.finalize_xof();
                let mut out = [0u8; 32];
                reader.read(&mut out);
                Digest::new(algorithm, out)
            }
            Self::Xxh3_64(hasher) => Digest::new(algorithm, hasher.finish().to_be_bytes()),
            Self::Md5(hasher) => Digest::new(algorithm, hasher.finalize()),
            Self::Sha1(hasher) => Digest::new(algorithm, hasher.finalize()),
        }
    }
}

fn new_states(algorithms: &[Algorithm]) -> Vec<(Algorithm, DigestState)> {
    algorithms
        .iter()
        .filter_map(|algorithm| {
            DigestState::new(*algorithm)
                .ok()
                .map(|state| (*algorithm, state))
        })
        .collect()
}

fn write_states(states: &mut [(Algorithm, DigestState)], data: &[u8]) {
    for (_, state) in states {
        state.update(data);
    }
}

fn digests_from_states(states: Vec<(Algorithm, DigestState)>) -> Result<Vec<Digest>, Error> {
    let mut digests = Vec::with_capacity(states.len());
    for (algorithm, state) in states {
        digests.push(state.finalize(algorithm)?);
    }
    normalize_digests(&mut digests);
    Ok(digests)
}

fn hash_chunks(
    data: &[u8],
    chunk_size: usize,
    algorithms: &[Algorithm],
) -> Result<Vec<Chunk>, Error> {
    let mut chunks = Vec::new();
    for (index, part) in data.chunks(chunk_size.max(1)).enumerate() {
        chunks.push(hash_chunk(
            index,
            (index * chunk_size) as u64,
            part,
            algorithms,
        )?);
    }
    Ok(chunks)
}

fn hash_chunk(
    index: usize,
    offset: u64,
    data: &[u8],
    algorithms: &[Algorithm],
) -> Result<Chunk, Error> {
    let mut states = new_states(algorithms);
    write_states(&mut states, data);
    Ok(Chunk {
        index,
        offset,
        size: data.len() as u64,
        digests: digests_from_states(states)?,
    })
}

fn tree_root_from_chunks(chunks: &[Chunk]) -> Result<Option<Digest>, Error> {
    if chunks.is_empty() {
        return Ok(None);
    }
    let mut hasher = blake3::Hasher::new();
    hasher.update(b"snhash-tree-v1\0");
    for chunk in chunks {
        hasher.update(&(chunk.index as u64).to_be_bytes());
        hasher.update(&chunk.offset.to_be_bytes());
        hasher.update(&chunk.size.to_be_bytes());
        let mut digests = chunk.digests.clone();
        normalize_digests(&mut digests);
        for digest in digests {
            hasher.update(digest.wire().as_bytes());
            hasher.update(&[0]);
        }
    }
    Ok(Some(Digest::new(
        Algorithm::Blake3_256,
        hasher.finalize().as_bytes(),
    )?))
}

fn manifest_from_parts(
    size: u64,
    mut digests: Vec<Digest>,
    chunks: Vec<Chunk>,
    tree_root: Option<Digest>,
    options: Options,
) -> Result<Manifest, Error> {
    normalize_digests(&mut digests);
    Ok(Manifest {
        version: MANIFEST_VERSION.to_owned(),
        profile: options.profile,
        size,
        media_type: options.media_type,
        canonicalization: options.canonicalization,
        source_name: options.source_name,
        digests,
        tree_root,
        chunks,
        entries: Vec::new(),
        custody: options.custody,
        created_at: options.created_at,
    })
}

fn collect_dir_entries(
    root: &Path,
    current: &Path,
    algorithms: &[Algorithm],
    options: &Options,
    entries: &mut Vec<Entry>,
) -> Result<(), Error> {
    let mut paths = fs::read_dir(current)?
        .map(|entry| entry.map(|entry| entry.path()))
        .collect::<Result<Vec<PathBuf>, io::Error>>()?;
    paths.sort();
    for path in paths {
        let metadata = fs::symlink_metadata(&path)?;
        let rel = path
            .strip_prefix(root)
            .unwrap_or(&path)
            .to_string_lossy()
            .replace('\\', "/");
        if metadata.is_dir() {
            entries.push(Entry {
                path: rel,
                kind: "directory".to_owned(),
                mode: Some(format_mode(&metadata)),
                size: 0,
                digests: Vec::new(),
            });
            collect_dir_entries(root, &path, algorithms, options, entries)?;
        } else if metadata.file_type().is_symlink() {
            let target = fs::read_link(&path)?.to_string_lossy().into_owned();
            let manifest = hash_bytes_with_options(
                target.as_bytes(),
                Options {
                    profile: options.profile,
                    algorithms: algorithms.to_vec(),
                    media_type: "text/plain".to_owned(),
                    canonicalization: "symlink-target".to_owned(),
                    created_at: options.created_at.clone(),
                    ..Options::default()
                },
            )?;
            entries.push(Entry {
                path: rel,
                kind: "symlink".to_owned(),
                mode: Some(format_mode(&metadata)),
                size: target.len() as u64,
                digests: manifest.digests,
            });
        } else if metadata.is_file() {
            let manifest = hash_file(
                &path,
                Options {
                    profile: options.profile,
                    algorithms: algorithms.to_vec(),
                    created_at: options.created_at.clone(),
                    ..Options::default()
                },
            )?;
            entries.push(Entry {
                path: rel,
                kind: "file".to_owned(),
                mode: Some(format_mode(&metadata)),
                size: metadata.len(),
                digests: manifest.digests,
            });
        }
    }
    Ok(())
}

#[cfg(unix)]
fn format_mode(metadata: &fs::Metadata) -> String {
    use std::os::unix::fs::PermissionsExt;
    format!("{:o}", metadata.permissions().mode() & 0o7777)
}

#[cfg(not(unix))]
fn format_mode(metadata: &fs::Metadata) -> String {
    if metadata.permissions().readonly() {
        "readonly".to_owned()
    } else {
        "readwrite".to_owned()
    }
}

fn directory_canonical_bytes(entries: &[Entry]) -> Vec<u8> {
    let mut out = Vec::new();
    for entry in entries {
        out.extend_from_slice(entry.path.as_bytes());
        out.push(0);
        out.extend_from_slice(entry.kind.as_bytes());
        out.push(0);
        if let Some(mode) = &entry.mode {
            out.extend_from_slice(mode.as_bytes());
        }
        out.push(0);
        let mut digests = entry.digests.clone();
        normalize_digests(&mut digests);
        for digest in digests {
            out.extend_from_slice(digest.wire().as_bytes());
            out.push(0);
        }
        out.push(b'\n');
    }
    out
}

fn trusted_algorithms(digests: &[Digest]) -> Vec<Algorithm> {
    let mut algorithms = digests
        .iter()
        .filter(|digest| digest.trusted())
        .map(|digest| digest.algorithm)
        .collect::<Vec<_>>();
    normalize_algorithms(&mut algorithms);
    algorithms
}

fn normalize_algorithms(algorithms: &mut Vec<Algorithm>) {
    algorithms.sort();
    algorithms.dedup();
}

fn normalize_digests(digests: &mut Vec<Digest>) {
    digests.sort_by_key(|digest| digest.algorithm);
}

fn xxh3_64_bytes(data: &[u8]) -> [u8; 8] {
    XxHash3_64::oneshot(data).to_be_bytes()
}

fn equal_bytes(a: &[u8], b: &[u8]) -> bool {
    if a.len() != b.len() {
        return false;
    }
    a.iter()
        .zip(b.iter())
        .fold(0u8, |acc, (a, b)| acc | (a ^ b))
        == 0
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Cursor;

    const FIXED_TIME: &str = "2026-05-01T12:00:00Z";

    #[test]
    fn evidence_vectors_match_known_hashes() {
        let manifest = hash_bytes_with_options(
            b"abc",
            Options::profile(Profile::Evidence).created_at(FIXED_TIME),
        )
        .unwrap();
        assert_eq!(manifest.size, 3);
        assert_eq!(
            manifest.blake3().unwrap().hex(),
            "6437b3ac38465133ffb63b75273a8db548c558465d79db03fd359c6cd5bd9d85"
        );
        assert_eq!(
            manifest.sha256().unwrap().hex(),
            "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
        );
        assert!(manifest.verify_bytes(b"abc"));
        assert!(!manifest.verify_bytes(b"abd"));
    }

    #[test]
    fn digest_wire_round_trips() {
        let digest = hash_bytes_profile(b"abc", Profile::Evidence)
            .unwrap()
            .sha256()
            .unwrap()
            .clone();
        let parsed = parse_digest_wire(&digest.wire()).unwrap();
        assert_eq!(parsed, digest);
    }

    #[test]
    fn cache_profile_keeps_xxh3_untrusted() {
        let manifest = hash_bytes_profile(b"cache payload", Profile::Cache).unwrap();
        let xxh3 = manifest.digest(Algorithm::Xxh3_64).unwrap();
        assert_eq!(xxh3.size, 8);
        assert!(!xxh3.trusted());
        assert!(manifest.digest(Algorithm::Blake3_256).is_some());
        assert!(manifest.verify_bytes(b"cache payload"));
    }

    #[test]
    fn hash_id_round_trips() {
        let manifest = hash_bytes_profile(b"hash id payload", Profile::Evidence).unwrap();
        let id = manifest.hash_id().unwrap();
        assert!(id.verify(b"hash id payload"));
        assert!(!id.verify(b"changed"));
        let parsed = parse_hash_id(&id.wire()).unwrap();
        assert_eq!(parsed.wire(), id.wire());
    }

    #[test]
    fn proof_root_is_deterministic() {
        let manifest = hash_bytes_profile(b"case evidence", Profile::Evidence).unwrap();
        let first = manifest.proof_root().unwrap();
        let second = manifest.proof_root().unwrap();
        assert_eq!(first, second);
        assert!(first.verify(&manifest.canonical_json().unwrap()));
    }

    #[test]
    fn chunk_verification_and_tree_root_work() {
        let manifest = hash_bytes_with_options(
            b"0123456789",
            Options::default().chunk_size(4).created_at(FIXED_TIME),
        )
        .unwrap();
        assert_eq!(manifest.chunks.len(), 3);
        assert!(manifest.tree_root.is_some());
        assert!(manifest.verify_chunk(0, b"0123"));
        assert!(!manifest.verify_chunk(0, b"0124"));
    }

    #[test]
    fn manifest_verify_reader_works() {
        let manifest = hash_bytes_profile(b"stream body", Profile::Evidence).unwrap();
        assert!(manifest.verify_reader(Cursor::new(b"stream body")).unwrap());
        assert!(
            !manifest
                .verify_reader(Cursor::new(b"stream body changed"))
                .unwrap()
        );
    }

    #[test]
    fn api_signatures_work() {
        let key = *b"01234567890123456789012345678901";
        let hmac = hmac_sha256(&key, b"request body").unwrap();
        assert!(verify_hmac_sha256(&key, b"request body", &hmac));
        assert!(!verify_hmac_sha256(&key, b"changed", &hmac));
        let keyed = keyed_blake3(&key, b"request body").unwrap();
        assert!(verify_keyed_blake3(&key, b"request body", &keyed));
        assert!(!verify_keyed_blake3(&key, b"changed", &keyed));
    }

    #[test]
    fn manifest_file_round_trips() {
        let temp = tempfile::tempdir().unwrap();
        let path = temp.path().join("manifest.json");
        let manifest = hash_bytes_profile(b"case evidence", Profile::Evidence).unwrap();
        manifest.write_file(&path).unwrap();
        let parsed = read_manifest_file(&path).unwrap();
        assert_eq!(
            parsed.canonical_json().unwrap(),
            manifest.canonical_json().unwrap()
        );
    }

    #[test]
    fn hash_dir_is_deterministic() {
        let temp = tempfile::tempdir().unwrap();
        fs::write(temp.path().join("b.txt"), b"bravo").unwrap();
        fs::create_dir(temp.path().join("sub")).unwrap();
        fs::write(temp.path().join("sub/a.txt"), b"alpha").unwrap();
        let first = hash_dir(temp.path(), Options::profile(Profile::Evidence)).unwrap();
        let second = hash_dir(temp.path(), Options::profile(Profile::Evidence)).unwrap();
        assert_eq!(first.proof_root().unwrap(), second.proof_root().unwrap());
        assert_eq!(first.entries.len(), 3);
    }

    #[cfg(feature = "snid-bridge")]
    #[test]
    fn snid_bridge_attaches_bid() {
        let manifest = hash_bytes_profile(b"content for bid", Profile::Evidence).unwrap();
        let bid = attach_snid(&manifest).unwrap();
        assert_eq!(bid.content.as_slice(), manifest.blake3().unwrap().bytes);
    }
}

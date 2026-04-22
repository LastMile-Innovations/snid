# Rust API Reference

Rust implementation API reference (API V2 - Universal Paradigms).

## Crate

```toml
[dependencies]
snid = "0.2"
```

## Core Philosophy

**"One Default, Infinite Extensibility"**

- **Decoupled Presentation**: Atoms are strictly a serialization concern
- **Strict Memory Tiers**: 16-byte (Snid) and 32-byte (Nid, Lid, Kid)
- **Zero-Allocation Hot Paths**: No heap allocations in generation
- **Universal Paradigms**: Consistent patterns across all languages

## Universal Paradigms

### Generation

```rust
let id = Snid::new();                    // Fastest path, ~5ns
let id = Snid::new_with(Options { tenant: Some("acme".to_string()), shard: Some(42) });
let id = Snid::new_spatial(lat, lng);    // Spatial IDs
let id = Snid::new_safe();               // Public-safe mode with time-blurring and CSPRNG entropy (~40-50ns)
```

### Batching

```rust
let batch = Snid::batch(1000);
```

### Parsing

```rust
let id = Snid::parse("MAT:2xXFhP...")?;
let id = Snid::parse_uuid("018f1c3e-...")?;
```

### Serialization

```rust
let wire = id.to_string();      // Default: "MAT:"
let wire = id.with_atom("IAM"); // Override: "IAM:"
let uuid = id.to_uuid_string(); // UUIDv7 format
let base32 = id.to_base32();    // Crockford Base32 (case-insensitive, human-friendly)
```

## Types

### Snid

Core 128-bit identifier type (Tier 1: 16-byte).

```rust
pub struct Snid(pub [u8; 16]);
```

### Options

Configuration for ID generation with zero-allocation (pass by value).

```rust
pub struct Options {
    pub tenant: Option<String>,
    pub shard: Option<u16>,
}
```

## Functions (Universal Paradigms)

### new

Generate a new SNID with ~5ns latency. This is the universal paradigm for fast ID generation.

```rust
pub fn new() -> Snid
```

### new_with

Generate a configured ID using options. This is the universal paradigm for configured ID generation.

```rust
pub fn new_with(opts: Options) -> Snid
```

### new_spatial

Generate a spatial ID from lat/lng coordinates. This is the universal paradigm for spatial ID generation.

```rust
pub fn new_spatial(lat: f64, lng: f64) -> Snid
```

### new_safe

Generate a public-safe ID with time-blurring and pure CSPRNG entropy. This is the "One ID" solution for database PK + public API use. Time-blurring truncates timestamp to nearest second (instead of millisecond). Pure CSPRNG fills 74 bits with cryptographic randomness (no monotonic counter). Performance: ~40-50ns (vs 5ns for new).

```rust
pub fn new_safe() -> Snid
```

### batch

Generate a batch of IDs efficiently. This is the universal paradigm for batch generation.

```rust
pub fn batch(count: usize) -> Vec<Snid>
```

### parse

Parse a wire string and return the ID. This is the universal paradigm for parsing wire strings.

```rust
pub fn parse(value: &str) -> Result<Snid, Error>
```

### parse_uuid

Parse a UUID string and return the ID. This is the universal paradigm for parsing UUID strings.

```rust
pub fn parse_uuid(value: &str) -> Result<Snid, Error>
```

## Functions (Legacy - Deprecated)

### new_fast

Generate a new SNID with lock-free per-P state. Deprecated: Use new() instead.

```rust
pub fn new_fast() -> Snid
```

### generate_batch

Generate a batch of SNIDs. Deprecated: Use batch() instead.

```rust
pub fn generate_batch(count: usize) -> Vec<Snid>
```

### parse_wire

Parse a wire string into an SNID. Deprecated: Use parse() instead.

```rust
pub fn parse_wire(s: &str) -> Result<(Snid, String), Error>
```

### from_bytes

Create SNID from bytes.

```rust
pub fn from_bytes(bytes: [u8; 16]) -> Snid
```

## Methods

### to_wire

Format SNID as wire string with atom.

```rust
pub fn to_wire(&self, atom: &str) -> String
```

### to_string

Format SNID using default "MAT:" atom. This is the universal paradigm for serialization (default atom).

```rust
pub fn to_string(&self) -> String
```

### with_atom

Format SNID with a custom atom. This is the universal paradigm for serialization (override atom).

```rust
pub fn with_atom(&self, atom: &str) -> String
```

### to_base32

Format SNID using Crockford Base32 encoding. This is case-insensitive and excludes ambiguous characters (I, L, O). Suitable for human-readable IDs and URLs.

```rust
pub fn to_base32(&self) -> String
```

### to_compact

Format SNID as compact wire string (no atom).

```rust
pub fn to_compact(&self) -> String
```

### to_uuid_string

Format SNID as standard UUID text.

```rust
pub fn to_uuid_string(&self) -> String
```

### as_bytes

Get SNID as byte slice.

```rust
pub fn to_bytes(self) -> [u8; 16]
```

### tensor128

Get tensor projection as hi/lo int64 pair.

```rust
pub fn to_tensor_words(self) -> (i64, i64)
```

## Extended ID Types

### SGID

Spatial ID type.

```rust
pub struct SGID([u8; 16]);

impl SGID {
    pub fn from_spatial_parts(h3_cell: u64, entropy: u64) -> SGID
    pub fn to_wire(&self, atom: &str) -> String
}
```

### NID

Neural ID type.

```rust
pub struct NID([u8; 32]);

impl NID {
    pub fn from_parts(base: SNID, semantic: [u8; 16]) -> NID
    pub fn tensor256(&self) -> (i64, i64, i64, i64)
}
```

### LID

Ledger ID type.

```rust
pub struct LID([u8; 32]);

impl LID {
    pub fn from_parts(head: SNID, prev: SNID, payload: SNID, key: &[u8]) -> Result<LID, Error>
}
```

### KID

Capability ID type.

```rust
pub struct KID([u8; 32]);

impl KID {
    pub fn for_capability(head: SNID, actor: &[u8], resource: &[u8], capability: &[u8], key: &[u8]) -> Result<KID, Error>
}
```

## Boundary Projections

### Tensor128

```rust
pub fn tensor128(&self) -> (i64, i64)
```

### Tensor256

```rust
pub fn tensor256(&self) -> (i64, i64, i64, i64)
```

## Errors

```rust
pub enum ParseError {
    InvalidString,
    ChecksumMismatch,
    InvalidLength,
    InvalidCharacter,
}
```

## Examples

### Basic Usage (Universal Paradigms)

```rust
use snid::Snid;

fn main() {
    // Generation
    let id = Snid::new();
    let wire = id.to_string();  // "MAT:..."
    println!("ID: {}", wire);

    // Custom atom
    let wire = id.with_atom("IAM");  // "IAM:..."
    println!("ID: {}", wire);

    // Configured generation
    let opts = snid::Options {
        tenant: Some("acme".to_string()),
        shard: Some(42),
    };
    let id = Snid::new_with(opts);
}
```

### Batch Generation

```rust
use snid::Snid;

fn main() {
    // Module-level batch (universal paradigm)
    let batch = Snid::batch(1000);
    for id in batch {
        println!("{}", id.to_string());
    }
}
```

### Parsing (Universal Paradigms)

```rust
use snid::Snid;

fn main() {
    // Parse wire string
    let wire = "MAT:2xXFhP9w7V4sKjBnG8mQpL";
    let id = Snid::parse(wire).expect("Failed to parse");
    println!("ID: {}", id);

    // Parse UUID
    let uuid_str = "018f1c3e-...";
    let id = Snid::parse_uuid(uuid_str).expect("Failed to parse");
    println!("ID: {}", id);
}
```

## Performance

- `new()`: ~5ns
- `generate_batch(1000)`: ~3μs
- `to_wire()`: ~60ns
- `parse_wire()`: ~120ns

## Features

### serde

Enable serde serialization:

```toml
[dependencies]
snid = { version = "0.2", features = ["serde"] }
```

```rust
use serde::{Serialize, Deserialize};

#[derive(Serialize, Deserialize)]
struct Item {
    id: SNID,
    name: String,
}
```

## See Also

- [Protocol Specification](../SPEC.md)
- [Basic Usage Guide](../guides/basic-usage.md)
- [Storage Contracts](../guides/storage-contracts.md)

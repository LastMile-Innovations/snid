# Rust API Reference

Rust implementation API reference.

## Crate

```toml
[dependencies]
snid = "0.2"
```

## Types

### SNID

Core 128-bit identifier type.

```rust
pub struct SNID([u8; 16]);
```

## Functions

### new

Generate a new SNID.

```rust
pub fn new() -> SNID
```

### generate_batch

Generate a batch of SNIDs.

```rust
pub fn generate_batch(count: usize) -> Vec<SNID>
```

### parse_wire

Parse a wire string into an SNID.

```rust
pub fn parse_wire(s: &str) -> Result<(SNID, String), ParseError>
```

### from_bytes

Create SNID from bytes.

```rust
pub fn from_bytes(bytes: &[u8]) -> Result<SNID, ParseError>
```

## Methods

### to_wire

Format SNID as wire string with atom.

```rust
pub fn to_wire(&self, atom: &str) -> String
```

### to_compact

Format SNID as compact wire string (no atom).

```rust
pub fn to_compact(&self) -> String
```

### to_uuid

Convert SNID to UUID.

```rust
pub fn to_uuid(&self) -> uuid::Uuid
```

### as_bytes

Get SNID as byte slice.

```rust
pub fn as_bytes(&self) -> &[u8; 16]
```

### tensor128

Get tensor projection as hi/lo int64 pair.

```rust
pub fn tensor128(&self) -> (i64, i64)
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

### Basic Usage

```rust
use snid::SNID;

fn main() {
    let id = SNID::new();
    let wire = id.to_wire("MAT");
    println!("ID: {}", wire);
}
```

### Batch Generation

```rust
use snid::SNID;

fn main() {
    let batch = SNID::generate_batch(1000);
    for id in batch {
        println!("{}", id.to_wire("MAT"));
    }
}
```

### Parsing

```rust
use snid::SNID;

fn main() {
    let wire = "MAT:2xXFhP9w7V4sKjBnG8mQpL";
    let (id, atom) = SNID::parse_wire(wire).expect("Failed to parse");
    println!("ID: {}, atom: {}", id, atom);
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

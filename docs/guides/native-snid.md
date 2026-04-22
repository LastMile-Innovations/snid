# Native SNID Guide

The native SNID protocol and its extended identifier families.

## Overview

Native SNID is the recommended choice for new projects. It provides superior performance, extended identifier families, AI/ML integration, and polyglot conformance across Go, Rust, and Python.

**Why choose Native SNID:**
- Best performance (4.106ns generation in the current local Go artifact)
- Extended ID families (SGID, NID, LID, KID, etc.)
- AI/ML integration (tensor projections, LLM formats)
- Polyglot conformance (byte-identical across languages)
- Storage contracts for all major databases

## Core SNID

The 128-bit core SNID is the foundation for all extended families.

### Generation

**Go:**
```go
import "github.com/LastMile-Innovations/snid"

id := snid.NewFast()
wire := id.String(snid.Matter)
// MAT:2xXFhP9w7V4sKjBnG8mQpL
```

**Rust:**
```rust
use snid::SNID;

let id = SNID::new();
let wire = id.to_wire("MAT");
```

**Python:**
```python
import snid

id = snid.SNID.new_fast()
wire = id.to_wire("MAT")
```

### Batch Generation

**Go:**
```go
batch := snid.NewBatch(snid.Matter, 1000)
for _, id := range batch {
    fmt.Println(id.String(snid.Matter))
}
```

**Rust:**
```rust
let batch = SNID::generate_batch(1000);
for id in batch {
    println!("{}", id.to_wire("MAT"));
}
```

**Python:**
```python
batch = snid.SNID.generate_batch(1000, backend="bytes")
```

## Extended ID Families

### SGID - Spatial ID

Location-aware identifiers using H3 geospatial encoding.

**Use cases:**
- Building tracking
- Sensor networks
- Location-based services
- Spatial queries

**Go:**
```go
sgid := snid.NewSpatial(37.7749, -122.4194) // San Francisco
wire := sgid.String(snid.Location)
// LOC:...
```

**Rust:**
```rust
let sgid = SGID::from_lat_lng(37.7749, -122.4194);
let wire = sgid.to_wire("LOC");
```

**Python:**
```python
sgid = snid.SGID.from_lat_lng(37.7749, -122.4194)
wire = sgid.to_wire("LOC")
```

See [Spatial IDs Guide](spatial-ids.md) for details.

### NID - Neural ID

Semantic identifiers for vector search and ML applications.

**Use cases:**
- Vector databases
- Semantic search
- ML pipelines
- Embedding-based systems

**Go:**
```go
base := snid.NewFast()
semantic := []byte{...} // 16-byte semantic tail
nid := snid.NewNeural(base, semantic)
```

**Rust:**
```rust
let base = SNID::new();
let semantic = [0u8; 16]; // 16-byte semantic tail
let nid = NID::from_parts(base, semantic);
```

**Python:**
```python
base = snid.SNID.new_fast()
semantic = bytes(16)  # 16-byte semantic tail
nid = snid.NID.from_parts(base, semantic)
```

See [Neural IDs Guide](neural-ids.md) for details.

### LID - Ledger ID

Verified identifiers for immutable logs and tamper-evident auditing.

**Use cases:**
- Blockchain
- Distributed ledgers
- Audit trails
- Tamper-evident storage

**Go:**
```go
head := snid.NewFast()
prev := snid.NewFast()
payload := snid.NewFast()
key := []byte("secret-key")
lid := snid.NewLID(head, prev, payload, key)
```

**Rust:**
```rust
let head = SNID::new();
let prev = SNID::new();
let payload = SNID::new();
let key = b"secret-key";
let lid = LID::from_parts(head, prev, payload, key)?;
```

**Python:**
```python
head = snid.SNID.new_fast()
prev = snid.SNID.new_fast()
payload = snid.SNID.new_fast()
key = b"secret-key"
lid = snid.LID.from_parts(head, prev, payload, key)
```

### KID - Capability ID

Self-verifying capability grants for authorization.

**Use cases:**
- API authorization
- Capability grants
- Edge cache validation
- Distributed authorization

**Go:**
```go
head := snid.NewFast()
actor := []byte("user-123")
resource := []byte("resource-456")
capability := []byte("read")
key := []byte("secret-key")
kid := snid.NewKID(head, actor, resource, capability, key)
```

**Rust:**
```rust
let head = SNID::new();
let actor = b"user-123";
let resource = b"resource-456";
let capability = b"read";
let key = b"secret-key";
let kid = KID::for_capability(head, actor, resource, capability, key)?;
```

**Python:**
```python
head = snid.SNID.new_fast()
actor = b"user-123"
resource = b"resource-456"
capability = b"read"
key = b"secret-key"
kid = snid.KID.for_capability(head, actor, resource, capability, key)
```

### AKID - Access Key ID

Dual-part public-plus-secret credentials for API keys.

**Use cases:**
- API keys
- Access tokens
- Credentials
- Secret management

**Go:**
```go
public := snid.NewAKIDPublic("tenant-123")
secret := snid.NewAKIDSecret()
```

**Rust:**
```rust
let public = AKIDPublic::new("tenant-123");
let secret = AKIDSecret::new();
```

**Python:**
```python
public = snid.AKID.public("tenant-123")
secret = snid.AKID.secret()
```

## Atoms

Atoms provide type-tagging at serialization time without changing the underlying ID.

**Canonical atoms:**
- `IAM` - Identity
- `TEN` - Tenant
- `MAT` - Matter
- `LOC` - Location
- `CHR` - Character
- `LED` - Ledger
- `LEG` - Legal
- `TRU` - Trust
- `KIN` - Kinetic
- `COG` - Cognition
- `SEM` - Semantic
- `SYS` - System
- `EVT` - Event
- `SES` - Session
- `KEY` - Key

**Go:**
```go
id.String(snid.Matter)  // MAT:...
id.String(snid.Location) // LOC:...
```

**Rust:**
```rust
id.to_wire("MAT")  // MAT:...
id.to_wire("LOC") // LOC:...
```

**Python:**
```python
id.to_wire("MAT")  # MAT:...
id.to_wire("LOC") # LOC:...
```

## Performance

Native SNID provides industry-leading performance:

| Operation | Go | Rust | Python |
|-----------|-----|------|--------|
| NewFast() | 4.106ns | ~5ns | ~15ns |
| NewBurst(1000) | 2.132μs | ~3μs | ~5μs (bytes) |
| String() | 106.5ns | ~60ns | ~80ns |
| AppendTo() | 94.42ns | n/a | n/a |
| FromString() | 173.4ns | ~120ns | ~150ns |

See [Performance Benchmarks](../performance/benchmarks.md) for details.

## Storage

Native SNID has canonical storage contracts for all major databases:

**PostgreSQL:**
```sql
CREATE TABLE items (
    id BYTEA PRIMARY KEY
);
```

**ClickHouse:**
```sql
CREATE TABLE items (
    id FixedString(16) PRIMARY KEY
) ENGINE = MergeTree()
ORDER BY id;
```

**MySQL:**
```sql
CREATE TABLE items (
    id BINARY(16) PRIMARY KEY
);
```

See [Storage Contracts Guide](storage-contracts.md) for details.

## AI/ML Integration

Native SNID provides AI/ML-friendly projections:

**Tensor128:**
```go
hi, lo := id.Tensor128()
```

**Tensor256 (for NID, LID, etc.):**
```go
w0, w1, w2, w3 := nid.Tensor256()
```

**LLMFormatV1:**
```go
llm := id.LLMFormatV1(atom)
```

See [AI/ML Integration Guide](ai-ml-integration.md) for details.

## When to Use Native SNID

Use Native SNID when:
- Building new projects
- Need maximum performance
- Require extended ID families
- Need AI/ML integration
- Want polyglot conformance
- Require spatial or semantic IDs
- Need verification capabilities

## Next Steps

- [Identifier Families Guide](identifier-families.md) - All ID families in detail
- [Spatial IDs Guide](spatial-ids.md) - SGID spatial IDs
- [Neural IDs Guide](neural-ids.md) - NID neural IDs
- [Storage Contracts Guide](storage-contracts.md) - Database integration
- [AI/ML Integration Guide](ai-ml-integration.md) - Tensor projections and LLM formats
- [Compatibility Modes Guide](compatibility-modes.md) - Drop-in replacements for other ID types

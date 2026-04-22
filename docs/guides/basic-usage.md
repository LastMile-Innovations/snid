# Basic Usage Guide

Common patterns and examples for using SNID in your applications (API V2 - Universal Paradigms).

## Generating IDs

### Single ID Generation (Universal Paradigm)

**Go:**
```go
id := snid.New()  // Fastest path, ~3.7ns
```

**Rust:**
```rust
let id = Snid::new();  // Fastest path, ~5ns
```

**Python:**
```python
id = snid.new()  # Fastest path, ~15ns
```

### Configured ID Generation

**Go:**
```go
id := snid.NewWith(snid.Options{Tenant: "acme", Shard: 42})
```

**Rust:**
```rust
let opts = snid::Options {
    tenant: Some("acme".to_string()),
    shard: Some(42),
};
let id = Snid::new_with(opts);
```

**Python:**
```python
id = snid.new_with(tenant="acme", shard=42)
```

### Batch Generation (Universal Paradigm)

**Go:**
```go
batch := snid.NewBatch(snid.Matter, 1000)
```

**Rust:**
```rust
let batch = Snid::batch(1000);
```

**Python:**
```python
# Python objects (default)
batch = snid.batch(1000, backend="snid")

# Raw bytes (fastest)
batch = snid.batch(1000, backend="bytes")

# NumPy arrays (requires snid[data])
batch = snid.batch(1000, backend="numpy")
```

### Public-Safe Mode Generation

For public-facing applications where you want to prevent ID enumeration and timestamp leakage:

**Go:**
```go
id := snid.NewSafe()  // ~40-50ns, time-blurred + CSPRNG entropy
```

**Rust:**
```rust
let id = Snid::new_safe();  // ~40-50ns, time-blurred + CSPRNG entropy
```

**Python:**
```python
id = snid.new_safe()  # ~40-50ns, time-blurred + CSPRNG entropy
```

This mode:
- Truncates timestamp to nearest second (time-blurring)
- Fills 74 bits with cryptographic randomness (no monotonic counter)
- Produces IDs safe for public APIs and URLs
- Slightly slower than standard generation but still essentially instant

## Wire Format

### Encoding to Wire String (Universal Paradigm)

**Go:**
```go
wire := id.StringDefault()  // Default: "MAT:..."
wire = id.WithAtom(snid.Identity)  // Override: "IAM:..."
```

**Rust:**
```rust
let wire = id.to_string();  // Default: "MAT:..."
let wire = id.with_atom("IAM");  // Override: "IAM:..."
```

**Python:**
```python
wire = id.string_default()  # Default: "MAT:..."
wire = id.with_atom("IAM")  # Override: "IAM:..."
```

### Crockford Base32 Encoding

For human-readable, case-insensitive IDs suitable for URLs and manual entry:

**Go:**
```go
base32 := id.StringBase32()  // Crockford Base32 (26 chars)
```

**Rust:**
```rust
let base32 = id.to_base32();  // Crockford Base32 (26 chars)
```

**Python:**
```python
base32 = id.to_base32()  # Crockford Base32 (26 chars)
```

Crockford Base32 features:
- Case-insensitive (A and a are the same)
- Excludes ambiguous characters (I, L, O to prevent confusion with 1 and 0)
- URL-safe and suitable for human-readable IDs
- Includes check digit for integrity validation

### Parsing from Wire String (Universal Paradigm)

**Go:**
```go
parsed, err := snid.Parse(wire)
```

**Rust:**
```rust
let parsed = Snid::parse(&wire)?;
```

**Python:**
```python
parsed = snid.parse(wire)
```

## Atoms

Atoms are type-tags applied at serialization time (decoupled from construction). Common atoms:

- `MAT` - Matter/objects
- `IAM` - Identity/users
- `TEN` - Tenants
- `LOC` - Location/spatial
- `LED` - Ledger/transactions
- `EVT` - Events
- `SES` - Sessions
- `KEY` - Keys/credentials

**Go:**
```go
id.WithAtom(snid.Matter)
id.WithAtom(snid.Identity)
id.WithAtom(snid.Location)
```

**Rust:**
```rust
id.with_atom("MAT")
id.with_atom("IAM")
id.with_atom("LOC")
```

**Python:**
```python
id.with_atom("MAT")
id.with_atom("IAM")
id.with_atom("LOC")
```

## Binary Storage

### Encoding to Binary

**Go:**
```go
bytes := id[:]
```

**Rust:**
```rust
let bytes = id.as_bytes();
```

**Python:**
```python
bytes = id.to_bytes()
```

### Decoding from Binary

**Go:**
```go
var id snid.ID
copy(id[:], bytes)
```

**Rust:**
```rust
let id = SNID::from_bytes(bytes)?;
```

**Python:**
```python
id = snid.SNID.from_bytes(bytes)
```

## UUID Compatibility

SNID is compatible with UUID v7. You can convert between formats:

**Go:**
```go
uuid := id.UUID()
```

**Rust:**
```rust
let uuid = id.to_uuid_string();
```

**Python:**
```python
uuid = id.to_uuid_string()
```

## Extended Identifier Families

### Spatial IDs (Universal Paradigm)

**Go:**
```go
sgid := snid.NewSpatial(37.7749, -122.4194) // San Francisco
```

**Rust:**
```rust
let sgid = Snid::new_spatial(37.7749, -122.4194);
```

**Python:**
```python
sgid = snid.new_spatial(37.7749, -122.4194)
```

### Neural IDs (NID)

**Go:**
```go
nid, err := snid.NewNeural(base_snid, semantic_hash)
```

**Rust:**
```rust
let nid = NID::from_parts(base_snid, semantic_hash);
```

**Python:**
```python
nid = snid.NID.from_parts(base_snid, semantic_hash)
```

### Ledger IDs (LID)

**Go:**
```go
lid, err := snid.NewLID(prev, payload, key)
```

**Rust:**
```rust
let lid = LID::from_parts(head, prev, payload, key)?;
```

**Python:**
```python
lid = snid.LID.from_parts(head, prev, payload, key)
```

## Database Integration

### SQL (PostgreSQL)

```sql
-- Store as UUID
CREATE TABLE items (
    id UUID PRIMARY KEY,
    name TEXT
);

-- Store as binary (recommended)
CREATE TABLE items (
    id BYTEA PRIMARY KEY,
    name TEXT
);
```

### Neo4j

```cypher
// Store as binary
CREATE (n:Item {id: $binary_id, name: $name})
```

See [Storage Contracts](storage-contracts.md) for more details.

## Error Handling (Universal Paradigm)

**Go:**
```go
id, err := snid.Parse(wire)
if err != nil {
    // Handle error
}
```

**Rust:**
```rust
let id = Snid::parse(&wire)?;
```

**Python:**
```python
try:
    parsed = snid.parse(wire)
except ValueError as e:
    # Handle error
```

## Next Steps

- [Batch Generation](batch-generation.md) - High-throughput patterns
- [Spatial IDs](spatial-ids.md) - H3 geospatial encoding
- [Neural IDs](neural-ids.md) - Semantic IDs for ML
- [Storage Contracts](storage-contracts.md) - Database integration

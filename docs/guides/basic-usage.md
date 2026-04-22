# Basic Usage Guide

Common patterns and examples for using SNID in your applications.

## Generating IDs

### Single ID Generation

**Go:**
```go
id := snid.NewFast()
```

**Rust:**
```rust
let id = SNID::new();
```

**Python:**
```python
id = snid.SNID.new_fast()
```

### Batch Generation

**Go:**
```go
batch := snid.NewBatch(snid.Matter, 1000)
```

**Rust:**
```rust
let batch = SNID::generate_batch(1000);
```

**Python:**
```python
# Raw bytes (fastest)
batch = snid.SNID.generate_batch(1000, backend="bytes")

# Tensor pairs
batch = snid.SNID.generate_batch(1000, backend="tensor")

# NumPy arrays (requires snid[data])
batch = snid.SNID.generate_batch(1000, backend="numpy")
```

## Wire Format

### Encoding to Wire String

**Go:**
```go
wire := id.String(snid.Matter)
// MAT:...
```

**Rust:**
```rust
let wire = id.to_wire("MAT");
// MAT:...
```

**Python:**
```python
wire = id.to_wire("MAT")
# MAT:...
```

### Parsing from Wire String

**Go:**
```go
parsed, atom, err := snid.FromString(wire)
```

**Rust:**
```rust
let (parsed, atom) = SNID::parse_wire(&wire)?;
```

**Python:**
```python
parsed, atom = snid.SNID.parse_wire(wire)
```

## Atoms

Atoms are type-tags applied at serialization time. Common atoms:

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
id.String(snid.Matter)
id.String(snid.Identity)
id.String(snid.Location)
```

**Rust:**
```rust
id.to_wire("MAT")
id.to_wire("IAM")
id.to_wire("LOC")
```

**Python:**
```python
id.to_wire("MAT")
id.to_wire("IAM")
id.to_wire("LOC")
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
let uuid = id.to_uuid();
```

**Python:**
```python
uuid = id.to_uuid()
```

## Extended Identifier Families

### Spatial IDs (SGID)

**Go:**
```go
sgid := snid.NewSpatial(37.7749, -122.4194) // San Francisco
```

**Rust:**
```rust
let sgid = SGID::from_spatial_parts(h3_cell, entropy);
```

**Python:**
```python
sgid = snid.SGID.from_spatial_parts(h3_cell, entropy)
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

## Error Handling

**Go:**
```go
id, atom, err := snid.FromString(wire)
if err != nil {
    // Handle error
}
```

**Rust:**
```rust
let (id, atom) = SNID::parse_wire(&wire)?;
```

**Python:**
```python
try:
    parsed, atom = snid.SNID.parse_wire(wire)
except ValueError as e:
    # Handle error
```

## Next Steps

- [Batch Generation](batch-generation.md) - High-throughput patterns
- [Spatial IDs](spatial-ids.md) - H3 geospatial encoding
- [Neural IDs](neural-ids.md) - Semantic IDs for ML
- [Storage Contracts](storage-contracts.md) - Database integration

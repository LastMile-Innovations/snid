# FAQ

Frequently asked questions about SNID.

## General

### What is SNID?

SNID is a modern polyglot sortable identifier protocol with UUID v7-compatible ordering, designed for distributed systems, AI pipelines, and high-scale infrastructure.

### Why use SNID instead of UUID?

SNID offers:
- **Better performance**: ~3.7ns vs ~50ns for UUID generation
- **Time-ordered**: Better for database indexing and sorting
- **Extended families**: SGID (spatial), NID (neural), LID (ledger), etc.
- **AI/ML support**: Tensor projections, LLM formats
- **Polyglot conformance**: Byte-identical across Go, Rust, Python

### Is SNID compatible with UUID v7?

Yes, SNID is byte-compatible with UUID v7. You can convert between formats:

```go
uuid := id.UUID()
snid := snid.ID(uuid)
```

### What are atoms?

Atoms are type-tags applied at serialization time (e.g., MAT for matter, IAM for identity). They provide context without changing the underlying ID.

## Installation

### How do I install SNID?

**Go:**
```bash
go get github.com/neighbor/snid
```

**Rust:**
```bash
cargo add snid
```

**Python:**
```bash
pip install snid
```

### What are the prerequisites?

- Go 1.24+
- Rust 1.70+
- Python 3.10+

## Usage

### How do I generate an ID?

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

### How do I generate IDs in bulk?

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
batch = snid.SNID.generate_batch(1000, backend="bytes")
```

### How do I parse a wire string?

**Go:**
```go
parsed, atom, err := snid.FromString("MAT:2xXFhP9w7V4sKjBnG8mQpL")
```

**Rust:**
```rust
let (parsed, atom) = SNID::parse_wire("MAT:2xXFhP9w7V4sKjBnG8mQpL")?;
```

**Python:**
```python
parsed, atom = snid.SNID.parse_wire("MAT:2xXFhP9w7V4sKjBnG8mQpL")
```

## Performance

### How fast is SNID?

- **Go**: ~3.7ns per ID (NewFast), ~1.7ns (TurboStreamer)
- **Rust**: ~5ns per ID
- **Python**: ~15ns per ID (native), ~5μs for 1000 batch (bytes)

### Which Python backend should I use?

- `bytes`: Fastest, for raw storage
- `tensor`: Fast, for tensor operations
- `numpy`: Zero-copy, for NumPy workflows
- `pyarrow`: Medium, for Arrow systems
- `polars`: Medium, for Polars workflows
- `snid`: Slowest, avoid in hot paths

## Conformance

### What is conformance testing?

Conformance testing ensures byte-identical behavior across Go, Rust, and Python implementations. It's the release gate for all changes.

### How do I run conformance tests?

```bash
just conformance
```

### When should I regenerate test vectors?

Regenerate vectors when:
- Protocol changes are made (byte layout, wire format)
- New identifier families are added
- New boundary projections are implemented
- Encoding/decoding logic changes

**Never commit `conformance/vectors.json` without regenerating from Go.**

## Extended ID Families

### What is SGID?

SGID (Spatial ID) provides location-aware identifiers using H3 geospatial encoding for building tracking, sensor networks, and location-based services.

### What is NID?

NID (Neural ID) provides semantic identifiers for vector search and ML applications, combining a time-ordered SNID head with a semantic tail.

### What is LID?

LID (Ledger ID) provides identifiers with HMAC verification tail for immutable logs and tamper-evident auditing.

### What is AKID?

AKID (Access Key ID) provides dual-part public-plus-secret credentials for API keys and access tokens.

## Database Integration

### How do I store SNIDs in PostgreSQL?

```sql
-- Recommended: BYTEA
CREATE TABLE items (
    id BYTEA PRIMARY KEY
);

-- Alternative: UUID
CREATE TABLE items (
    id UUID PRIMARY KEY
);
```

### How do I store SNIDs in MySQL?

```sql
CREATE TABLE items (
    id BINARY(16) PRIMARY KEY
);
```

### How do I store SNIDs in Neo4j?

```cypher
CREATE (n:Item {id: $binary_id})
```

## Migration

### How do I migrate from UUID?

See [From UUID](../migration/from-uuid.md) for detailed migration guide.

### How do I migrate from ULID?

See [From ULID](../migration/from-ulid.md) for detailed migration guide.

### How do I migrate from KSUID?

See [From KSUID](../migration/from-ksuid.md) for detailed migration guide.

## Contributing

### How do I contribute?

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for contribution guidelines.

### What is the release process?

See [Release Process](../../CONTRIBUTING.md#release-process) in CONTRIBUTING.md.

## Troubleshooting

### I'm getting a checksum mismatch error

This means the wire string was corrupted or modified. Ensure you're using the exact wire string returned by SNID.

### I'm getting an invalid character error

The wire string contains invalid Base58 characters. Ensure you're using the correct alphabet (no 0, O, I, l).

### Conformance tests are failing

1. Regenerate vectors: `just conformance`
2. Check if protocol changes were made
3. Verify encoding/decoding logic
4. Check all three implementations

### Python installation fails

Ensure you have Python 3.10+ and maturin installed:

```bash
pip install maturin
cd python
maturin develop
```

## Next Steps

- [Quick Start](../guides/quick-start.md) - Get started in 5 minutes
- [Basic Usage](../guides/basic-usage.md) - Common patterns
- [Common Errors](common-errors.md) - Error troubleshooting
- [Debugging](debugging.md) - Debugging tips

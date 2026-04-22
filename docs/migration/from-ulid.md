# Migrating from ULID to SNID

Guide for migrating from ULID to SNID.

## Overview

ULID and SNID both provide time-ordered identifiers, but SNID offers:

- UUID v7 compatibility
- Better performance (~3.7ns vs ~50ns for ULID)
- Extended identifier families
- AI/ML support (tensor projections)
- Polyglot conformance

## Key Differences

| Feature | ULID | SNID |
|---------|------|------|
| Byte layout | 16 bytes | 16 bytes (UUID v7 compatible) |
| Encoding | Base32 (Crockford) | Base58 + atom |
| Time precision | Milliseconds | Milliseconds |
| Performance | ~50ns | ~3.7ns |
| Sorting | Time-ordered | Time-ordered |
| Extended families | No | Yes (SGID, NID, LID, etc.) |
| Wire format | ULID string | Atom-prefixed wire string |

## Migration Strategy

### 1. Dual-Write Period

Write both ULID and SNID during migration:

```sql
ALTER TABLE items ADD COLUMN snid_id BYTEA;

-- Update application to write both
INSERT INTO items (ulid_id, snid_id, name) VALUES ($1, $2, $3);
```

### 2. Backfill Existing Data

Since ULID and SNID have different byte layouts, generate new SNIDs for existing data:

```sql
-- Generate new SNIDs for existing ULIDs
UPDATE items 
SET snid_id = generate_new_snid()
WHERE snid_id IS NULL;
```

**Go:**
```go
func backfillSNIDs(db *sql.DB) error {
    rows, err := db.Query("SELECT id FROM items WHERE snid_id IS NULL")
    if err != nil {
        return err
    }
    defer rows.Close()

    for rows.Next() {
        var id int
        if err := rows.Scan(&id); err != nil {
            return err
        }

        newSNID := snid.NewFast()
        _, err := db.Exec(
            "UPDATE items SET snid_id = $1 WHERE id = $2",
            newSNID[:], id,
        )
        if err != nil {
            return err
        }
    }
    return nil
}
```

### 3. Switch Reads

Update application to read from SNID:

```sql
-- Query by SNID
SELECT * FROM items WHERE snid_id = $1;
```

### 4. Drop ULID Column

```sql
-- After validation
ALTER TABLE items DROP COLUMN ulid_id;
```

## Code Migration

### Go

**Before (ULID):**
```go
import "github.com/oklog/ulid/v2"

id := ulid.Make()
```

**After (SNID):**
```go
import "github.com/LastMile-Innovations/snid"

id := snid.NewFast()
```

### Rust

**Before (ULID):**
```rust
use ulid::Ulid;

let id = Ulid::new();
```

**After (SNID):**
```rust
use snid::SNID;

let id = SNID::new();
```

### Python

**Before (ULID):**
```python
import ulid

id = ulid.new()
```

**After (SNID):**
```python
import snid

id = snid.SNID.new_fast()
```

## Wire Format Migration

### ULID String Format

ULID uses Crockford Base32:

```
01ARZ3NDEKTSV4RRFFQ69G5FAV
```

### SNID Wire Format

SNID uses Base58 with atom prefix:

```
MAT:2xXFhP9w7V4sKjBnG8mQpL
```

**Migration:**

```go
// Parse ULID string
ulid, err := ulid.ParseStrict("01ARZ3NDEKTSV4RRFFQ69G5FAV")

// Generate new SNID with same timestamp
snid := snid.NewFast() // Uses current time
```

## Database Migration

### PostgreSQL

```sql
-- Add SNID column
ALTER TABLE items ADD COLUMN snid_id BYTEA;

-- Create index
CREATE INDEX idx_items_snid_id ON items (snid_id);

-- Backfill with new SNIDs
UPDATE items SET snid_id = gen_random_bytes(16);

-- Switch primary key (after validation)
ALTER TABLE items DROP CONSTRAINT items_pkey;
ALTER TABLE items ADD PRIMARY KEY (snid_id);

-- Drop old column
ALTER TABLE items DROP COLUMN ulid_id;
```

### MySQL

```sql
-- Add SNID column
ALTER TABLE items ADD COLUMN snid_id BINARY(16);

-- Create index
CREATE INDEX idx_items_snid_id ON items (snid_id);

-- Backfill
UPDATE items SET snid_id = RANDOM_BYTES(16);

-- Switch primary key
ALTER TABLE items DROP PRIMARY KEY;
ALTER TABLE items ADD PRIMARY KEY (snid_id);

-- Drop old column
ALTER TABLE items DROP COLUMN ulid_id;
```

## API Migration

### REST API

**Before:**
```json
{
  "id": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
  "name": "item"
}
```

**After:**
```json
{
  "id": "MAT:2xXFhP9w7V4sKjBnG8mQpL",
  "name": "item"
}
```

## Compatibility

### Hybrid Support

If you need to support both during migration:

```go
func ParseID(s string) (snid.ID, error) {
    if len(s) == 26 {
        // ULID format (26 chars)
        ulid, err := ulid.ParseStrict(s)
        if err == nil {
            // Convert ULID to SNID (new ID with same time)
            // Note: Byte layouts differ, so we generate new SNID
            return snid.NewFast(), nil
        }
    }
    // SNID format
    return snid.FromString(s)
}
```

## Testing

### Validation

```sql
-- Verify all SNIDs are valid
SELECT COUNT(*) FROM items WHERE snid_id IS NULL;

-- Verify ordering
SELECT snid_id FROM items ORDER BY snid_id LIMIT 10;
```

### Performance Testing

```go
// Benchmark ULID vs SNID
func BenchmarkULID(b *testing.B) {
    for i := 0; i < b.N; i++ {
        ulid.Make()
    }
}

func BenchmarkSNID(b *testing.B) {
    for i := 0; i < b.N; i++ {
        snid.NewFast()
    }
}
```

## Rollback Plan

If migration fails:

```sql
-- Revert to ULID
ALTER TABLE items ADD COLUMN ulid_id CHAR(26);
UPDATE items SET ulid_id = generate_ulid();
ALTER TABLE items DROP COLUMN snid_id;
```

## Best Practices

1. **Generate new SNIDs** for existing data (byte layouts differ)
2. **Test thoroughly** in staging environment
3. **Use dual-write** during migration
4. **Monitor performance** after switch
5. **Update documentation** and APIs

## Advantages of SNID over ULID

- **Better performance**: ~3.7ns vs ~50ns
- **UUID v7 compatibility**: Interoperability with UUID systems
- **Extended families**: SGID, NID, LID, etc.
- **AI/ML support**: Tensor projections, LLM formats
- **Polyglot conformance**: Byte-identical across Go, Rust, Python

## Next Steps

- [From UUID](from-uuid.md) - Migrating from UUID
- [From KSUID](from-ksuid.md) - Migrating from KSUID
- [Basic Usage](../guides/basic-usage.md) - SNID usage patterns

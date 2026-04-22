# Migrating from UUID to SNID

Guide for migrating from UUID to SNID.

## Overview

SNID is compatible with UUID v7, making migration straightforward. SNID provides:

- UUID v7-compatible byte layout
- Time-ordered identifiers
- Better performance
- Extended identifier families

## Key Differences

| Feature | UUID v7 | SNID |
|---------|---------|------|
| Byte layout | Compatible | Compatible |
| Ordering | Time-ordered | Time-ordered |
| Performance | Standard | Optimized (~3.7ns) |
| Wire format | Hex string | Base58 + atom |
| Extended families | No | Yes (SGID, NID, LID, etc.) |
| AI/ML support | No | Yes (tensor projections) |

## Migration Strategy

### 1. Dual-Write Period

Write both UUID and SNID during migration:

```sql
ALTER TABLE items ADD COLUMN snid_id BYTEA;

-- Update application to write both
INSERT INTO items (uuid_id, snid_id, name) VALUES ($1, $2, $3);
```

### 2. Backfill Existing Data

```sql
-- Generate SNIDs for existing UUIDs
UPDATE items 
SET snid_id = convert_uuid_to_snid(uuid_id)
WHERE snid_id IS NULL;
```

**Go:**
```go
func convertUUIDToSNID(uuid uuid.UUID) snid.ID {
    return snid.ID(uuid)
}
```

**Python:**
```python
import snid
import uuid

def convert_uuid_to_snid(uuid_str):
    u = uuid.UUID(uuid_str)
    return snid.SNID.from_bytes(u.bytes)
```

### 3. Switch Reads

Update application to read from SNID:

```sql
-- Query by SNID
SELECT * FROM items WHERE snid_id = $1;
```

### 4. Drop UUID Column

```sql
-- After validation
ALTER TABLE items DROP COLUMN uuid_id;
```

## Code Migration

### Go

**Before (UUID):**
```go
import "github.com/google/uuid"

id := uuid.New()
```

**After (SNID):**
```go
import "github.com/neighbor/snid"

id := snid.NewFast()
```

### Rust

**Before (UUID):**
```rust
use uuid::Uuid;

let id = Uuid::new_v4();
```

**After (SNID):**
```rust
use snid::SNID;

let id = SNID::new();
```

### Python

**Before (UUID):**
```python
import uuid

id = uuid.uuid4()
```

**After (SNID):**
```python
import snid

id = snid.SNID.new_fast()
```

## Database Migration

### PostgreSQL

```sql
-- Add SNID column
ALTER TABLE items ADD COLUMN snid_id BYTEA;

-- Create index
CREATE INDEX idx_items_snid_id ON items (snid_id);

-- Backfill
UPDATE items SET snid_id = uuid_id::bytea;

-- Switch primary key (after validation)
ALTER TABLE items DROP CONSTRAINT items_pkey;
ALTER TABLE items ADD PRIMARY KEY (snid_id);

-- Drop old column
ALTER TABLE items DROP COLUMN uuid_id;
```

### MySQL

```sql
-- Add SNID column
ALTER TABLE items ADD COLUMN snid_id BINARY(16);

-- Create index
CREATE INDEX idx_items_snid_id ON items (snid_id);

-- Backfill
UPDATE items SET snid_id = uuid_id;

-- Switch primary key
ALTER TABLE items DROP PRIMARY KEY;
ALTER TABLE items ADD PRIMARY KEY (snid_id);

-- Drop old column
ALTER TABLE items DROP COLUMN uuid_id;
```

## API Migration

### REST API

**Before:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
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

Or keep binary for efficiency:
```json
{
  "id": "base64_encoded_bytes",
  "name": "item"
}
```

## Compatibility

### Drop-in UUIDv7 Replacement

SNID provides a true drop-in replacement for RFC 9562 UUIDv7. The core SNID type produces byte-for-byte identical output to reference UUIDv7 implementations (.NET 9, uuid crate v7, Python 3.14 uuid7, PostgreSQL uuid_generate_v7()).

#### Quick Start

**Go:**
```go
import "github.com/neighbor/snid"

// Generate UUIDv7-compatible ID
id := snid.NewUUIDv7()

// Convert to standard UUID string format
uuidStr := id.UUIDString() // "018f1c3e-5a7b-7c8d-9e0f-1a2b3c4d5e6f"

// Parse from UUID string
parsed, err := snid.ParseUUIDString("018f1c3e-5a7b-7c8d-9e0f-1a2b3c4d5e6f")

// Convert to/from github.com/google/uuid.UUID
uuidObj := id.ToUUIDv7()
fromUUID, err := snid.FromUUIDv7(uuidObj)
```

**Rust:**
```rust
use snid::Snid;

// Generate UUIDv7-compatible ID
let id = Snid::new();

// Format as UUID string
let uuid_str = id.to_uuid_string();

// Parse from UUID string
let parsed = Snid::from_uuid_string("018f1c3e-5a7b-7c8d-9e0f-1a2b3c4d5e6f")?;
```

**Python:**
```python
import snid

# Generate UUIDv7-compatible ID
id = snid.SNID.new_uuidv7()

# Format as UUID string
uuid_str = id.to_uuid_string()

# Parse from UUID string
parsed = snid.SNID.parse_uuid_string("018f1c3e-5a7b-7c8d-9e0f-1a2b3c4d5e6f")
```

#### Byte-for-Byte Compatibility

SNID produces identical bytes to reference UUIDv7 implementations:

```go
// SNID byte layout matches RFC 9562 UUIDv7
id := snid.NewUUIDv7()
bytes := id[:] // 16 bytes, UUIDv7-compatible

// Version nibble: 0b0111 (version 7)
version := (id[6] >> 4) & 0x0F // = 7

// Variant bits: 0b10
variant := (id[8] >> 6) & 0b11 // = 0b10
```

#### Migration Path for UUIDv7 Users

If you're already using UUIDv7, migration is trivial:

1. **Replace generator calls:**
   ```go
   // Before
   id := uuid.NewV7()

   // After
   id := snid.NewUUIDv7()
   ```

2. **Keep using UUID string format:**
   ```go
   // Both produce identical output
   uuidStr := id.UUIDString() // "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
   ```

3. **Use binary storage unchanged:**
   ```sql
   -- UUIDv7 and SNID use the same 16-byte storage
   CREATE TABLE items (id UUID PRIMARY KEY);
   ```

#### When to Use Native SNID vs UUIDv7 Mode

**Use `NewUUIDv7()` when:**
- Migrating from existing UUIDv7 codebases
- Need strict RFC 9562 compliance
- Interoperating with systems expecting standard UUID format
- Using databases with native UUID types

**Use native `NewFast()` when:**
- Want SNID wire format benefits (Base58 + atom)
- Need extended identifier families (SGID, NID, LID, etc.)
- Building new systems without UUID constraints
- Want AI/ML tensor projections

### UUID v7 Compatibility

SNID is byte-compatible with UUID v7:

```go
// Convert SNID to UUID
uuid := id.UUID()

// Convert UUID to SNID
snid := snid.ID(uuid)
```

### Legacy UUID Support

If you need to support legacy UUIDs:

```go
func ParseID(s string) (snid.ID, error) {
    if len(s) == 36 {
        // UUID format
        u, err := uuid.Parse(s)
        if err == nil {
            return snid.ID(u), nil
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
// Benchmark UUID vs SNID
func BenchmarkUUID(b *testing.B) {
    for i := 0; i < b.N; i++ {
        uuid.New()
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
-- Revert to UUID
ALTER TABLE items ADD COLUMN uuid_id UUID;
UPDATE items SET uuid_id = snid_id::uuid;
ALTER TABLE items DROP COLUMN snid_id;
```

## Best Practices

1. **Test thoroughly** in staging environment
2. **Use dual-write** during migration
3. **Monitor performance** after switch
4. **Keep rollback plan** ready
5. **Update documentation** and APIs

## Next Steps

- [From ULID](from-ulid.md) - Migrating from ULID
- [From KSUID](from-ksuid.md) - Migrating from KSUID
- [Basic Usage](../guides/basic-usage.md) - SNID usage patterns

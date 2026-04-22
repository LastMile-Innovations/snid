# Migrating from KSUID to SNID

Guide for migrating from KSUID to SNID.

## Overview

KSUID and SNID both provide time-ordered identifiers, but SNID offers:

- UUID v7 compatibility
- Better performance (~3.7ns vs ~100ns for KSUID)
- Extended identifier families
- AI/ML support (tensor projections)
- Polyglot conformance

## Key Differences

| Feature | KSUID | SNID |
|---------|-------|------|
| Byte layout | 20 bytes | 16 bytes (UUID v7 compatible) |
| Encoding | Base62 | Base58 + atom |
| Time precision | Seconds | Milliseconds |
| Performance | ~100ns | ~3.7ns |
| Sorting | Time-ordered | Time-ordered |
| Extended families | No | Yes (SGID, NID, LID, etc.) |

## Migration Strategy

Since KSUID is 20 bytes and SNID is 16 bytes, generate new SNIDs for existing data:

```sql
ALTER TABLE items ADD COLUMN snid_id BYTEA;
UPDATE items SET snid_id = generate_new_snid() WHERE snid_id IS NULL;
```

## Code Migration

**Go:**
```go
// Before
import "github.com/segmentio/ksuid"
id := ksuid.New()

// After
import "github.com/neighbor/snid"
id := snid.NewFast()
```

**Rust:**
```rust
// Before
use ksuid::Ksuid;
let id = Ksuid::new();

// After
use snid::SNID;
let id = SNID::new();
```

**Python:**
```python
# Before
import ksuid
id = ksuid.ksuid()

# After
import snid
id = snid.SNID.new_fast()
```

## Database Migration

```sql
-- Add SNID column
ALTER TABLE items ADD COLUMN snid_id BINARY(16);

-- Backfill with new SNIDs
UPDATE items SET snid_id = RANDOM_BYTES(16);

-- Switch primary key
ALTER TABLE items DROP PRIMARY KEY;
ALTER TABLE items ADD PRIMARY KEY (snid_id);

-- Drop old column
ALTER TABLE items DROP COLUMN ksuid_id;
```

## Advantages of SNID over KSUID

- **Better performance**: ~3.7ns vs ~100ns
- **UUID v7 compatibility**: Interoperability with UUID systems
- **Extended families**: SGID, NID, LID, etc.
- **AI/ML support**: Tensor projections, LLM formats
- **Polyglot conformance**: Byte-identical across Go, Rust, Python

## Next Steps

- [From UUID](from-uuid.md) - Migrating from UUID
- [From ULID](from-ulid.md) - Migrating from ULID
- [Basic Usage](../guides/basic-usage.md) - SNID usage patterns

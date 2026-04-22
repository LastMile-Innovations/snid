# Compatibility Modes Guide

Drop-in replacements for UUIDv7, ULID, NanoID, KSUID, and other ID types.

## Overview

SNID provides compatibility modes for generating IDs in the format of other popular ID systems. This allows you to use SNID as a single library for all your ID needs while maintaining compatibility with existing systems.

**Why use compatibility modes:**
- One library for all ID types
- Seamless migration from existing systems
- Team standards compliance
- Interoperability with external systems

## UUIDv7 Mode

### Overview

UUIDv7 (RFC 9562) is the modern standard for time-ordered UUIDs. SNID provides a true drop-in replacement that produces byte-for-byte identical output to reference implementations.

**Binary Layout (RFC 9562):**
```
Bits 0-47:     unix_ts_ms (48-bit Unix timestamp in milliseconds, big-endian)
Bits 48-51:    Version = 0b0111
Bits 52-63:    rand_a (12 bits) - used for monotonicity or sub-ms precision
Bits 64-65:    Variant = 0b10
Bits 66-127:   rand_b (62 bits) - random
```

### Generation

**Go:**
```go
import "github.com/LastMile-Innovations/snid"

uuidv7 := snid.NewUUIDv7()
string := uuidv7.String()  // xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
bytes := uuidv7.Bytes()    // [16]byte
```

**Rust:**
```rust
use snid::SNID;

let uuidv7 = SNID::uuidv7();
let string = uuidv7.to_string();  // xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
let bytes = uuidv7.as_bytes();   // &[u8; 16]
```

**Python:**
```python
import snid

uuidv7 = snid.SNID.new_uuidv7()
string = str(uuidv7)  # xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
bytes = uuidv7.bytes   # bytes
```

### Migration

**From UUIDv7 to SNID:**
```go
// Existing UUIDv7
uuidv7 := uuid.MustParse("0189abcd-1234-5678-9abc-0123456789ab")

// Convert to SNID
snid := snid.FromUUID(uuidv7)
```

**From SNID to UUIDv7:**
```go
// Existing SNID
snid := snid.NewFast()

// Convert to UUIDv7
uuidv7 := snid.ToUUID()
```

### Compatibility

SNID UUIDv7 mode is byte-for-byte compatible with:
- .NET 9 `Guid.CreateVersion7()`
- uuid crate v7
- Python 3.14 `uuid.uuid7()`
- PostgreSQL `uuid_generate_v7()`

### When to Use

Use UUIDv7 mode when:
- Need exact UUIDv7 compatibility
- Migrating from UUIDv7 to SNID
- Interoperating with systems expecting UUIDv7
- Database requires UUID type

## ULID Mode

### Overview

ULID provides 26-character Crockford Base32 time-ordered identifiers. SNID can generate ULID-compatible strings.

**Format:** 26-char Crockford Base32 (no hyphens, no ambiguous characters)

### Generation

**Go:**
```go
import "github.com/LastMile-Innovations/snid"

ulid := snid.NewULID()
string := ulid.String()  // 26-char Crockford Base32
```

**Rust:**
```rust
use snid::SNID;

let ulid = SNID::ulid();
let string = ulid.to_string();  // 26-char Crockford Base32
```

**Python:**
```python
import snid

ulid = snid.SNID.new_ulid()
string = str(ulid)  # 26-char Crockford Base32
```

### When to Use

Use ULID mode when:
- Need human-readable, copy-pasteable IDs
- Migrating from ULID to SNID
- Interoperating with systems expecting ULID
- Prefer Base32 encoding

## NanoID Mode

### Overview

NanoID provides configurable-length URL-safe identifiers. SNID can generate NanoID-compatible strings.

**Format:** URL-safe Base64, configurable length (default 21 chars)

### Generation

**Go:**
```go
import "github.com/LastMile-Innovations/snid"

nanoid := snid.NewNanoID()          // Default 21 chars
nanoid := snid.NewNanoIDLength(10)  // Custom length
string := nanoid.String()
```

**Rust:**
```rust
use snid::SNID;

let nanoid = SNID::nanoid();              // Default 21 chars
let nanoid = SNID::nanoid_with_length(10); // Custom length
let string = nanoid.to_string();
```

**Python:**
```python
import snid

nanoid = snid.SNID.new_nanoid()          # Default 21 chars
nanoid = snid.SNID.new_nanoid(length=10)  # Custom length
string = str(nanoid)
```

### When to Use

Use NanoID mode when:
- Need short, URL-friendly IDs
- Migrating from NanoID to SNID
- Frontend ID generation
- API tokens

## KSUID Mode

### Overview

KSUID provides 20-byte time-ordered identifiers popular in Go ecosystems. SNID can generate KSUID-compatible strings.

**Format:** 20-byte, Base62 encoded

### Generation

**Go:**
```go
import "github.com/LastMile-Innovations/snid"

ksuid := snid.NewKSUID()
string := ksuid.String()
```

**Rust:**
```rust
use snid::SNID;

let ksuid = SNID::ksuid();
let string = ksuid.to_string();
```

**Python:**
```python
import snid

ksuid = snid.SNID.new_ksuid()
string = str(ksuid)
```

### When to Use

Use KSUID mode when:
- Migrating from KSUID to SNID
- Interoperating with Go systems expecting KSUID
- Need 20-byte time-ordered IDs

## CUID2 Mode

### Overview

CUID2 provides collision-resistant identifiers with fingerprinting. SNID can generate CUID2-compatible strings.

**Format:** Variable-length, collision-resistant

### Generation

**Go:**
```go
import "github.com/LastMile-Innovations/snid"

cuid2 := snid.NewCUID2()
string := cuid2.String()
```

**Rust:**
```rust
use snid::SNID;

let cuid2 = SNID::cuid2();
let string = cuid2.to_string();
```

**Python:**
```python
import snid

cuid2 = snid.SNID.new_cuid2()
string = str(cuid2)
```

### When to Use

Use CUID2 mode when:
- Need maximum collision resistance
- Migrating from CUID2 to SNID
- High-concurrency environments

## TSID Mode

### Overview

TSID (Time-Sorted Unique ID) provides compact time-ordered identifiers popular in Java/Spring ecosystems.

**Format:** Compact, time-ordered

### Generation

**Go:**
```go
import "github.com/LastMile-Innovations/snid"

tsid := snid.NewTSID()
string := tsid.String()
```

**Rust:**
```rust
use snid::SNID;

let tsid = SNID::tsid();
let string = tsid.to_string();
```

**Python:**
```python
import snid

tsid = snid.SNID.new_tsid()
string = str(tsid)
```

### When to Use

Use TSID mode when:
- Migrating from TSID to SNID
- Interoperating with Java/Spring systems
- Need compact time-ordered IDs

## Unified Generator

SNID provides a unified generator API for all modes:

**Go:**
```go
id, err := snid.Generate(snid.ModeUUIDv7)
id, err := snid.Generate(snid.ModeULID)
id, err := snid.Generate(snid.ModeNanoID)
```

**Rust:**
```rust
let id = SNID::generate(Mode::UUIDv7)?;
let id = SNID::generate(Mode::ULID)?;
let id = SNID::generate(Mode::NanoID)?;
```

**Python:**
```python
id = snid.SNID.generate("uuidv7")
id = snid.SNID.generate("ulid")
id = snid.SNID.generate("nanoid")
```

## Decision Table

| I want… | Use This Mode | Why |
|---------|---------------|-----|
| Exact UUIDv7 compatibility | `ModeUUIDv7` | Byte-for-byte RFC 9562 compatible |
| Human-readable sortable string | `ModeULID` | 26-char Crockford Base32 |
| Shortest URL-friendly ID | `ModeNanoID` | 21 chars by default, configurable |
| 20-byte time-ordered ID | `ModeKSUID` | Go ecosystem compatible |
| Maximum collision resistance | `ModeCUID2` | Fingerprinting included |
| Compact time-ordered ID | `ModeTSID` | Java/Spring compatible |
| Best performance + features | Native SNID | Full SNID protocol with extended families |

## Migration Guides

See detailed migration guides:
- [From UUID](../migration/from-uuid.md)
- [From ULID](../migration/from-ulid.md)
- [From KSUID](../migration/from-ksuid.md)

## Performance

Compatibility modes have similar performance to native SNID:

| Mode | Generation | Encoding |
|------|------------|----------|
| UUIDv7 | ~3.7ns | ~50ns |
| ULID | ~3.7ns | ~60ns |
| NanoID | ~3.7ns | ~40ns |
| KSUID | ~3.7ns | ~70ns |

## Next Steps

- [Native SNID Guide](native-snid.md) - Native SNID protocol
- [Migration Guides](../migration/) - Detailed migration instructions
- [Performance Comparison](../performance/comparison.md) - Compare with other ID systems
- [Why SNID](../why-snid.md) - Why choose SNID

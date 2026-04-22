# Go API Reference

Go implementation API reference (API V2 - Universal Paradigms).

## Package

```go
import "github.com/LastMile-Innovations/snid"
```

## Core Philosophy

**"One Default, Infinite Extensibility"**

- **Decoupled Presentation**: Atoms are strictly a serialization concern
- **Strict Memory Tiers**: 16-byte (ID) and 32-byte (NID, LID, KID)
- **Zero-Allocation Hot Paths**: No heap allocations in generation
- **Universal Paradigms**: Consistent patterns across all languages

## Universal Paradigms

### Generation

```go
id := snid.New()                    // Fastest path, ~3.7ns
id = snid.NewWith(snid.Options{Tenant: "acme", Shard: 42})
id = snid.NewSpatial(lat, lng)      // Spatial IDs
id = snid.NewSafe()                 // Public-safe mode with time-blurring and CSPRNG entropy (~40-50ns)
```

### Batching

```go
batch := snid.NewBatch(snid.Matter, 1000)  // Pre-allocates and fills
```

### Parsing

```go
id, err := snid.Parse("MAT:2xXFhP...")       // Parse wire string
id, err := snid.ParseUUIDString("018f1c3e-...")  // Parse UUID
```

### Serialization

```go
wire := id.StringDefault()     // Default: "MAT:"
wire := id.WithAtom("IAM")     // Override: "IAM:"
uuid := id.UUIDString()        // UUIDv7 format
base32 := id.StringBase32()    // Crockford Base32 (case-insensitive, human-friendly)
```

## Types

### ID

Core 128-bit identifier type (Tier 1: 16-byte).

```go
type ID [16]byte
```

### Options

Configuration for ID generation with zero-allocation (pass by value).

```go
type Options struct {
    Tenant string
    Shard  uint16
    Time   time.Time
}
```

### Atom

Atom type for wire format prefixes (serialization-time only).

```go
type Atom string

const (
    Identity  Atom = "IAM"
    Tenant    Atom = "TEN"
    Matter    Atom = "MAT"
    Location  Atom = "LOC"
    Character Atom = "CHR"
    Ledger   Atom = "LED"
    Legal    Atom = "LEG"
    Trust    Atom = "TRU"
    Kinetic  Atom = "KIN"
    Cognition Atom = "COG"
    Semantic Atom = "SEM"
    System   Atom = "SYS"
    Event    Atom = "EVT"
    Session  Atom = "SES"
    Key      Atom = "KEY"
)
```

## Functions (Universal Paradigms)

### New

Generate a new SNID with ~3.7ns latency. This is the universal paradigm for fast ID generation.

```go
func New() ID
```

### NewWith

Generate a configured ID using stack-allocated options. Zero-allocation.

```go
func NewWith(opts Options) ID
```

### NewSpatial

Generate a spatial ID from lat/lng. Spatial IDs are 16-byte IDs with H3 encoding.

```go
func NewSpatial(lat, lng float64) ID
```

### NewSpatialPrecise

Generate a spatial ID with specific H3 resolution.

```go
func NewSpatialPrecise(lat, lng float64, resolution int) ID
```

### NewSafe

Generate a public-safe ID with time-blurring and pure CSPRNG entropy. This is the "One ID" solution for database PK + public API use. Time-blurring truncates timestamp to nearest second (instead of millisecond). Pure CSPRNG fills 74 bits with cryptographic randomness (no monotonic counter). Performance: ~40-50ns (vs 3.7ns for New).

```go
func NewSafe() ID
```

### Parse

Parse a wire string and return the ID. Universal paradigm for parsing.

```go
func Parse(s string) (ID, error)
```

### ParseUUIDString

Parse a UUID string and return the ID.

```go
func ParseUUIDString(s string) (ID, error)
```

## Functions (Legacy - Deprecated)

### NewFast

Generate a new SNID with lock-free per-P state. Deprecated: Use New() instead.

```go
func NewFast() ID
```

### NewProjected

Generate a new SNID with tenant and shard. Deprecated: Use NewWith() instead.

```go
func NewProjected(tenantID string, shard uint16) ID
```

### NewBatch

Generate a batch of SNIDs. Deprecated: Use NewBatch() with atom parameter ignored.

```go
func NewBatch(atom Atom, count int) []ID
```

### NewAsset

Generate an asset ID from catalog, tenant, and serial.

```go
func NewAsset(catalogID ID, tenantID string, serial string) ID
```

### NewCatalog

Generate a catalog ID.

```go
func NewCatalog(tenantID string, name string) ID
```

### FromString

Parse a wire string into an ID.

```go
func FromString(s string) (ID, Atom, error)
```

### FromUUID

Convert SNID's dependency-free UUID value to an ID.

```go
func FromUUID(u UUID) ID
```

## Methods

### String

Format ID as wire string with atom.

```go
func (id ID) String(atom Atom) string
```

### StringBase32

Format ID using Crockford Base32 encoding. This is case-insensitive and excludes ambiguous characters (I, L, O). Suitable for human-readable IDs and URLs.

```go
func (id ID) StringBase32() string
```

### StringCompact

Format ID as compact wire string (no atom).

```go
func (id ID) StringCompact() string
```

### UUID

Convert ID to SNID's dependency-free UUID value.

```go
func (id ID) UUID() UUID
```

### Tensor128

Get tensor projection as hi/lo int64 pair.

```go
func (id ID) Tensor128() (hi, lo int64)
```

### MarshalJSON

Marshal ID to JSON (wire format).

```go
func (id ID) MarshalJSON() ([]byte, error)
```

### UnmarshalJSON

Unmarshal ID from JSON.

```go
func (id *ID) UnmarshalJSON(data []byte) error
```

### MarshalBinary

Marshal ID to binary.

```go
func (id ID) MarshalBinary() ([]byte, error)
```

### UnmarshalBinary

Unmarshal ID from binary.

```go
func (id *ID) UnmarshalBinary(data []byte) error
```

### Value

Implement sql.Valuer for database storage.

```go
func (id ID) Value() (driver.Value, error)
```

### Scan

Implement sql.Scanner for database retrieval.

```go
func (id *ID) Scan(value interface{}) error
```

## Extended ID Types

### SGID

Spatial ID type.

```go
type SGID ID

func NewSpatial(lat, lng float64) SGID
func (sgid SGID) String(atom Atom) string
```

### NID

Neural ID type.

```go
type NID [32]byte

func NewNeural(base ID, semantic [16]byte) (NID, error)
func (nid NID) Tensor256() (w0, w1, w2, w3 int64)
```

### LID

Ledger ID type.

```go
type LID [32]byte

func NewLID(head, prev, payload ID, key []byte) (LID, error)
```

### AKID

Access Key ID type.

```go
type AKIDPublic ID
type AKIDSecret [16]byte

func NewAKIDPublic(tenantID string) AKIDPublic
func NewAKIDSecret() (AKIDSecret, error)
```

## Boundary Projections

### Tensor128

```go
func (id ID) Tensor128() (hi, lo int64)
```

### Tensor256

```go
func (nid NID) Tensor256() (w0, w1, w2, w3 int64)
```

### LLMFormatV1

```go
func (id ID) LLMFormatV1(atom Atom) LLMFormatV1
```

## Errors

```go
var ErrInvalidString = errors.New("invalid wire string")
var ErrChecksumMismatch = errors.New("checksum mismatch")
var ErrInvalidLength = errors.New("invalid length")
```

## Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/LastMile-Innovations/snid"
)

func main() {
    id := snid.NewFast()
    wire := id.String(snid.Matter)
    fmt.Printf("ID: %s\n", wire)
}
```

### Batch Generation

```go
batch := snid.NewBatch(snid.Matter, 1000)
for _, id := range batch {
    fmt.Println(id.String(snid.Matter))
}
```

### Database Storage

```go
import "database/sql"

func insertItem(db *sql.DB, id snid.ID, name string) error {
    _, err := db.Exec(
        "INSERT INTO items (id, name) VALUES ($1, $2)",
        id[:], name,
    )
    return err
}
```

## Performance

- `NewFast()`: 4.106ns, 0 allocs
- `NewBurst(1000)`: 2.132μs, 1 alloc for the returned slice
- `String()`: 106.5ns, 1 alloc
- `AppendTo()`: 94.42ns, 0 allocs with caller-provided capacity
- `FromString()`: 173.4ns, 0 allocs

## See Also

- [Protocol Specification](../SPEC.md)
- [Basic Usage Guide](../guides/basic-usage.md)
- [Storage Contracts](../guides/storage-contracts.md)

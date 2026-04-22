# Go API Reference

Go implementation API reference.

## Package

```go
import "github.com/neighbor/snid"
```

## Types

### ID

Core 128-bit identifier type.

```go
type ID [16]byte
```

### Atom

Atom type for wire format prefixes.

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

## Functions

### New

Generate a new SNID using NewFast().

```go
func New(atom Atom) ID
```

### NewFast

Generate a new SNID with lock-free per-P state (~3.7ns).

```go
func NewFast() ID
```

### NewProjected

Generate a new SNID with tenant and shard.

```go
func NewProjected(tenantID string, shard uint16) ID
```

### NewBatch

Generate a batch of SNIDs.

```go
func NewBatch(atom Atom, count int) []ID
```

### NewSpatial

Generate a spatial ID (SGID) from lat/lng.

```go
func NewSpatial(lat, lng float64) ID
```

### NewSpatialPrecise

Generate a spatial ID with specific H3 resolution.

```go
func NewSpatialPrecise(lat, lng float64, resolution int) ID
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

Convert a UUID to an ID.

```go
func FromUUID(u uuid.UUID) ID
```

## Methods

### String

Format ID as wire string with atom.

```go
func (id ID) String(atom Atom) string
```

### StringCompact

Format ID as compact wire string (no atom).

```go
func (id ID) StringCompact() string
```

### UUID

Convert ID to UUID.

```go
func (id ID) UUID() uuid.UUID
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
    "github.com/neighbor/snid"
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

- `NewFast()`: ~3.7ns
- `NewBatch(1000)`: ~2μs
- `String()`: ~50ns
- `FromString()`: ~100ns

## See Also

- [Protocol Specification](../SPEC.md)
- [Basic Usage Guide](../guides/basic-usage.md)
- [Storage Contracts](../guides/storage-contracts.md)

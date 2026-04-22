# Python API Reference

Python implementation API reference (API V2 - Universal Paradigms).

## Installation

```bash
pip install snid
```

With data science dependencies:

```bash
pip install snid[data]
```

## Core Philosophy

**"One Default, Infinite Extensibility"**

- **Decoupled Presentation**: Atoms are strictly a serialization concern
- **Strict Memory Tiers**: 16-byte (SNID) and 32-byte (NID, LID, KID)
- **Zero-Allocation Hot Paths**: No heap allocations in generation
- **Universal Paradigms**: Consistent patterns across all languages

## Universal Paradigms

### Generation

```python
id = snid.new()                    # Fastest path, ~15ns
id = snid.new_with(tenant="acme", shard=42)
id = snid.new_spatial(lat, lng)    # Spatial IDs
id = snid.new_safe()               # Public-safe mode with time-blurring and CSPRNG entropy (~40-50ns)
```

### Batching

```python
batch = snid.batch(1000, backend="snid")  # Default: Python objects
batch = snid.batch(1000, backend="bytes")  # Raw bytes (fastest)
```

### Parsing

```python
id = snid.parse("MAT:2xXFhP...")       # Parse wire string
id = snid.parse_uuid("018f1c3e-...")   # Parse UUID
```

### Serialization

```python
wire = id.string_default()     # Default: "MAT:"
wire = id.with_atom("IAM")     # Override: "IAM:"
uuid = id.to_uuid_string()     # UUIDv7 format
base32 = id.to_base32()       # Crockford Base32 (case-insensitive, human-friendly)
```

## Classes

### SNID

Core 128-bit identifier class (Tier 1: 16-byte).

```python
class SNID:
    @staticmethod
    def new_fast() -> SNID  # Deprecated: Use snid.new()
    @staticmethod
    def new_safe() -> SNID  # Public-safe mode with time-blurring and CSPRNG entropy
    @staticmethod
    def generate_batch(count: int, backend: str = "snid") -> Union[bytes, List[Tuple[int, int]], np.ndarray, pa.Array, pl.Series]
    @staticmethod
    def parse_wire(s: str) -> Tuple[SNID, str]
    @staticmethod
    def from_bytes(data: bytes) -> SNID
    @staticmethod
    def from_uuid_string(s: str) -> SNID
    @staticmethod
    def load_vectors(path: str) -> Dict
    @staticmethod
    def tensor_time_delta(left: Tuple[int, int], right: Tuple[int, int]) -> int
    def to_base32(self) -> str  # Crockford Base32 encoding
```

## Module Functions (Universal Paradigms)

### new

Generate a new SNID with ~15ns latency. This is the universal paradigm for fast ID generation.

```python
def new() -> SNID
```

### new_with

Generate a configured ID using options. This is the universal paradigm for configured ID generation.

```python
def new_with(tenant: str | None = None, shard: int | None = None) -> SNID
```

### new_spatial

Generate a spatial ID from lat/lng coordinates. This is the universal paradigm for spatial ID generation.

```python
def new_spatial(lat: float, lng: float) -> SNID
```

### new_safe

Generate a public-safe ID with time-blurring and pure CSPRNG entropy. This is the "One ID" solution for database PK + public API use. Time-blurring truncates timestamp to nearest second (instead of millisecond). Pure CSPRNG fills 74 bits with cryptographic randomness (no monotonic counter). Performance: ~40-50ns (vs 15ns for new).

```python
def new_safe() -> SNID
```

### batch

Generate a batch of IDs efficiently. This is the universal paradigm for batch generation.

```python
def batch(count: int, *, backend: str = "snid") -> Union[bytes, List[SNID], List[Tuple[int, int]], np.ndarray, pa.Array, pl.Series]
```

### parse

Parse a wire string and return the ID. This is the universal paradigm for parsing wire strings.

```python
def parse(s: str) -> SNID
```

### parse_uuid

Parse a UUID string and return the ID. This is the universal paradigm for parsing UUID strings.

```python
def parse_uuid(s: str) -> SNID
```

## Methods

### to_wire

Format SNID as wire string with atom.

```python
def to_wire(self, atom: str) -> str
```

### string_default

Format SNID using default "MAT:" atom. This is the universal paradigm for serialization (default atom).

```python
def string_default(self) -> str
```

### with_atom

Format SNID with a custom atom. This is the universal paradigm for serialization (override atom).

```python
def with_atom(self, atom: str) -> str
```

### to_base32

Format SNID using Crockford Base32 encoding. This is case-insensitive and excludes ambiguous characters (I, L, O). Suitable for human-readable IDs and URLs.

```python
def to_base32(self) -> str
```

### to_compact

Format SNID as compact wire string (no atom).

```python
def to_compact(self) -> str
```

### to_uuid_string

Format SNID as standard UUID text.

```python
def to_uuid_string(self) -> str
```

### to_bytes

Get SNID as bytes.

```python
def to_bytes(self) -> bytes
```

### tensor128

Get tensor projection as hi/lo int64 pair.

```python
def tensor128(self) -> Tuple[int, int]
```

## Batch Generation Backends

### bytes

Raw bytes backend (fastest).

```python
batch = SNID.generate_batch(1000, backend="bytes")
# Returns: bytes (16 * count)
```

### tensor

Tensor pairs backend.

```python
batch = SNID.generate_batch(1000, backend="tensor")
# Returns: List[Tuple[int, int]]
```

### numpy

NumPy array backend (zero-copy).

```python
batch = SNID.generate_batch(1000, backend="numpy")
# Returns: np.ndarray shape (count, 2)
```

### pyarrow

PyArrow array backend.

```python
batch = SNID.generate_batch(1000, backend="pyarrow")
# Returns: pa.Array
```

### polars

Polars series backend.

```python
batch = SNID.generate_batch(1000, backend="polars")
# Returns: pl.Series
```

### snid

Python object backend (slowest).

```python
batch = SNID.generate_batch(1000, backend="snid")
# Returns: List[SNID]
```

## Extended ID Types

### SGID

Spatial ID type.

```python
class SGID:
    @staticmethod
    def from_spatial_parts(h3_cell: int, entropy: int) -> SGID
    def to_wire(self, atom: str) -> str
```

### NID

Neural ID type.

```python
class NID:
    @staticmethod
    def from_parts(base: SNID, semantic: bytes) -> NID
    def tensor256(self) -> Tuple[int, int, int, int]
```

### LID

Ledger ID type.

```python
class LID:
    @staticmethod
    def from_parts(head: SNID, prev: SNID, payload: SNID, key: bytes) -> LID
```

### KID

Capability ID type.

```python
class KID:
    @staticmethod
    def for_capability(head: SNID, actor: bytes, resource: bytes, capability: bytes, key: bytes) -> KID
```

### AKID

Access Key ID type.

```python
class AKID:
    @staticmethod
    def public(tenant_id: str) -> AKIDPublic
    @staticmethod
    def secret() -> AKIDSecret
```

## Boundary Projections

### Tensor128

```python
def tensor128(self) -> Tuple[int, int]
```

### Tensor256

```python
def tensor256(self) -> Tuple[int, int, int, int]
```

### Time Delta

```python
@staticmethod
def tensor_time_delta(left: Tuple[int, int], right: Tuple[int, int]) -> int
```

## Errors

```python
class ParseError(Exception):
    pass

class ChecksumError(Exception):
    pass
```

## Examples

### Basic Usage (Universal Paradigms)

```python
import snid

# Generation
id = snid.new()
wire = id.string_default()  # "MAT:..."
print(f"ID: {wire}")

# Custom atom
wire = id.with_atom("IAM")  # "IAM:..."
print(f"ID: {wire}")

# Configured generation
id = snid.new_with(tenant="acme", shard=42)
```

### Batch Generation

```python
import snid

# Module-level batch (universal paradigm)
batch = snid.batch(1000, backend="snid")
print(f"Generated {len(batch)} SNIDs")

# Raw bytes (fastest)
batch = snid.batch(1000, backend="bytes")
print(f"Generated {len(batch)} bytes")

# NumPy
import numpy as np
batch = snid.batch(1000, backend="numpy")
print(f"Shape: {batch.shape}")
```

### Parsing (Universal Paradigms)

```python
import snid

# Parse wire string
wire = "MAT:2xXFhP9w7V4sKjBnG8mQpL"
id = snid.parse(wire)
print(f"ID: {id}")

# Parse UUID
uuid_str = "018f1c3e-..."
id = snid.parse_uuid(uuid_str)
print(f"ID: {id}")
```

### Tensor Operations

```python
import snid

batch = snid.batch(1000, backend="numpy")
timestamps = batch[:, 0] >> 16
print(f"Timestamps: {timestamps[:10]}")

# Time delta
delta = snid.SNID.tensor_time_delta(batch[0], batch[1])
print(f"Time delta: {delta} ms")
```

## Performance

- `new_fast()`: ~15ns
- `generate_batch(1000, backend="bytes")`: ~5μs
- `to_wire()`: ~80ns
- `parse_wire()`: ~150ns

## Data Science Integration

### NumPy

```python
import snid
import numpy as np

batch = snid.SNID.generate_batch(10000, backend="numpy")
df = np.column_stack([batch, np.random.rand(10000, 10)])
```

### PyArrow

```python
import snid
import pyarrow as pa

batch = snid.SNID.generate_batch(1000, backend="pyarrow")
table = pa.Table.from_arrays([batch], names=["id"])
```

### Polars

```python
import snid
import polars as pl

batch = snid.SNID.generate_batch(1000, backend="polars")
df = pl.DataFrame({"id": batch, "value": pl.arange(0, 1000)})
```

## Neo4j Integration

```python
from snid import neo4j

neo4j_id = neo4j.to_neo4j_id(id)
session.run("CREATE (n:Item {id: $id})", id=neo4j_id)
```

## See Also

- [Protocol Specification](../SPEC.md)
- [Basic Usage Guide](../guides/basic-usage.md)
- [Storage Contracts](../guides/storage-contracts.md)
- [AI/ML Integration](../guides/ai-ml-integration.md)

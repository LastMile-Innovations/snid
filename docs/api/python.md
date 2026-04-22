# Python API Reference

Python implementation API reference.

## Installation

```bash
pip install snid
```

With data science dependencies:

```bash
pip install snid[data]
```

## Classes

### SNID

Core 128-bit identifier class.

```python
class SNID:
    @staticmethod
    def new_fast() -> SNID
    @staticmethod
    def generate_batch(count: int, backend: str = "snid") -> Union[bytes, List[Tuple[int, int]], np.ndarray, pa.Array, pl.Series]
    @staticmethod
    def parse_wire(s: str) -> Tuple[SNID, str]
    @staticmethod
    def from_bytes(data: bytes) -> SNID
    @staticmethod
    def load_vectors(path: str) -> Dict
    @staticmethod
    def tensor_time_delta(left: Tuple[int, int], right: Tuple[int, int]) -> int
```

## Methods

### to_wire

Format SNID as wire string with atom.

```python
def to_wire(self, atom: str) -> str
```

### to_compact

Format SNID as compact wire string (no atom).

```python
def to_compact(self) -> str
```

### to_uuid

Convert SNID to UUID.

```python
def to_uuid(self) -> uuid.UUID
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

### Basic Usage

```python
import snid

id = snid.SNID.new_fast()
wire = id.to_wire("MAT")
print(f"ID: {wire}")
```

### Batch Generation

```python
import snid

# Raw bytes
batch = snid.SNID.generate_batch(1000, backend="bytes")
print(f"Generated {len(batch)} bytes")

# NumPy
import numpy as np
batch = snid.SNID.generate_batch(1000, backend="numpy")
print(f"Shape: {batch.shape}")
```

### Parsing

```python
import snid

wire = "MAT:2xXFhP9w7V4sKjBnG8mQpL"
parsed, atom = snid.SNID.parse_wire(wire)
print(f"ID: {parsed}, atom: {atom}")
```

### Tensor Operations

```python
import snid

batch = snid.SNID.generate_batch(1000, backend="numpy")
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

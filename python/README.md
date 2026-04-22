# SNID Python Bindings

Python bindings for the SNID polyglot sortable identifier protocol, backed by the Rust `snid-core` implementation.

## Installation

```bash
pip install snid
```

For data backends (numpy, pyarrow, polars):

```bash
pip install snid[data]
```

## Basic Usage

```python
import snid

# Generate a new ID
id = snid.SNID.new_fast()
wire = id.to_wire("MAT")
print(wire)  # MAT:...

# Parse a wire string
parsed, atom = snid.SNID.parse_wire(wire)
print(atom)  # MAT
```

## Batch Generation

The Python bindings provide native batch generation for high-throughput ingestion:

```python
import snid

# Generate raw bytes
batch = snid.SNID.generate_batch(1000, backend="bytes")

# Generate tensor pairs
batch = snid.SNID.generate_batch(1000, backend="tensor")

# Generate zero-copy NumPy arrays
import numpy as np
batch = snid.SNID.generate_batch(1000, backend="numpy")

# Generate PyArrow arrays
import pyarrow as pa
batch = snid.SNID.generate_batch(1000, backend="pyarrow")

# Generate Polars series
import polars as pl
batch = snid.SNID.generate_batch(1000, backend="polars")
```

## Tensor Operations

```python
import snid

# Compute time deltas from tensor words
left = (1234567890000 << 16) | 0x7000
right = (1234567880000 << 16) | 0x7000
delta = snid.SNID.tensor_time_delta(left, right)
```

## Extended Identifier Families

```python
import snid

# Spatial ID (SGID)
sgid = snid.SGID.from_spatial_parts(h3_cell, entropy)

# Neural ID (NID)
nid = snid.NID.from_parts(head_snid, semantic_hash)

# Ledger ID (LID)
lid = snid.LID.from_parts(head_snid, prev, payload, key)

# World ID (WID)
wid = snid.WID.from_parts(head_snid, scenario_hash)

# Edge ID (XID)
xid = snid.XID.from_parts(head_snid, edge_hash)

# Capability ID (KID)
kid = snid.KID.from_parts(head_snid, actor_snid, resource, capability, key)

# Ephemeral ID (EID)
eid = snid.EID.from_parts(unix_millis, counter)

# Content-addressable ID (BID)
bid = snid.BID.from_parts(topology_snid, content_hash)
```

## Neo4j Integration

```python
import snid
from snid import neo4j

# Convert to Neo4j format
neo4j_id = neo4j.to_neo4j_id(id)

# Convert from Neo4j format
id = neo4j.from_neo4j_id(neo4j_id)
```

## Development

Build the native module:

```bash
cd python
maturin develop
```

Run tests:

```bash
cd python
python -m pytest tests/
```

Run benchmarks:

```bash
cd python
python bench_batch.py
```

## License

MIT OR Apache-2.0

## Links

- [Protocol Specification](https://github.com/LastMile-Innovations/snid/blob/main/docs/SPEC.md)
- [Repository](https://github.com/LastMile-Innovations/snid)

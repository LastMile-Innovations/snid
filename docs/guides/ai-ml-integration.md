# AI/ML Integration Guide

Tensor projections and LLM formats for AI/ML pipelines.

## Overview

SNID provides AI/ML-friendly projections for:

- Tensor operations (Tensor128, Tensor256)
- LLM formats (LLMFormatV1, LLMFormatV2)
- Time binning for causal masking
- Zero-copy NumPy/PyArrow/Polars integration

## Tensor Projections

### Tensor128

128-bit IDs as big-endian int64 tensor pairs: `[hi, lo]`

**Go:**
```go
hi, lo := id.Tensor128()
```

**Rust:**
```rust
let (hi, lo) = id.tensor128();
```

**Python:**
```python
hi, lo = id.tensor128()
```

### Tensor256

256-bit IDs as four big-endian int64 words: `[w0, w1, w2, w3]`

**Go:**
```go
w0, w1, w2, w3 := nid.Tensor256()
```

**Rust:**
```rust
let (w0, w1, w2, w3) = nid.tensor256();
```

**Python:**
```python
w0, w1, w2, w3 = nid.tensor256()
```

## Python Batch Generation

### NumPy Backend (Zero-Copy)

```python
import snid
import numpy as np

# Generate as NumPy array (zero-copy view)
batch = snid.SNID.generate_batch(10000, backend="numpy")
print(f"Shape: {batch.shape}")  # (10000, 2)

# Extract timestamps
timestamps = batch[:, 0] >> 16
print(f"Timestamps: {timestamps[:10]}")

# Time deltas
deltas = snid.SNID.tensor_time_delta(batch[0], batch[1])
print(f"Time delta: {deltas} ms")
```

### PyArrow Backend

```python
import snid
import pyarrow as pa

# Generate as PyArrow array
batch = snid.SNID.generate_batch(1000, backend="pyarrow")

# Convert to table
table = pa.Table.from_arrays([batch], names=["id"])

# Export to Parquet
pa.parquet.write_table(table, "ids.parquet")
```

### Polars Backend

```python
import snid
import polars as pl

# Generate as Polars series
batch = snid.SNID.generate_batch(1000, backend="polars")

# Create DataFrame
df = pl.DataFrame({
    "id": batch,
    "timestamp": pl.datetime_range(...)
})

# Efficient operations
df = df.with_columns([
    (pl.col("id").str.extract_bytes(0, 8).cast(pl.Int64) >> 16).alias("timestamp_ms")
])
```

## LLM Formats

### LLMFormatV1

Minimal AI projection: `[ATOM, timestamp_ms, machine_or_shard, sequence]`

**Go:**
```go
llm := id.LLMFormatV1(atom)
```

**Rust:**
```rust
let llm = id.llm_format_v1(atom);
```

**Python:**
```python
llm = id.to_llm_format(atom)
```

### LLMFormatV2

Richer AI projection with additional metadata:

```python
llm_v2 = {
    "kind": "snid",
    "atom": "MAT",
    "timestamp_millis": timestamp,
    "machine_or_shard": machine,
    "sequence": sequence,
    "ghosted": False
}
```

## Time Binning

### TimeBin Projection

Resolution-truncated temporal projection for causal masking:

```python
# Get time bin (e.g., 1-second resolution)
time_bin = id.time_bin(resolution_ms=1000)
```

### Causal Masking

```python
import snid
import numpy as np

# Generate batch
batch = snid.SNID.generate_batch(1000, backend="numpy")

# Extract time bins
time_bins = (batch[:, 0] >> 16) // 1000  # 1-second bins

# Create causal mask
causal_mask = time_bins[:, None] <= time_bins[None, :]
```

## Use Cases

### Training Data Pipelines

```python
import snid
import numpy as np
import polars as pl

# Generate IDs for training data
ids = snid.SNID.generate_batch(100000, backend="numpy")

# Create training DataFrame
df = pl.DataFrame({
    "id": ids,
    "timestamp": ids[:, 0] >> 16,
    "features": np.random.rand(100000, 128)
})

# Save to Parquet
df.write_parquet("training_data.parquet")
```

### Vector Database Keys

```python
import snid
import numpy as np

# Generate IDs for vectors
ids = snid.SNID.generate_batch(10000, backend="bytes")

# Insert into vector database
for i in range(0, len(ids), 16):
    vector_db.insert(ids[i:i+16], embeddings[i//16])
```

### Causal Attention Windows

```python
import snid
import numpy as np

# Generate IDs with timestamps
ids = snid.SNID.generate_batch(1000, backend="numpy")
timestamps = ids[:, 0] >> 16

# Create attention windows
window_size = 100
for i in range(len(ids)):
    window_start = max(0, i - window_size)
    window_ids = ids[window_start:i+1]
    # Process window
```

### Semantic Search with NIDs

```python
import snid
import numpy as np

# Generate NIDs for embeddings
embeddings = np.random.rand(1000, 128)
semantic_hashes = [hash_embedding(e) for e in embeddings]

base_ids = snid.SNID.generate_batch(1000, backend="bytes")
nids = [snid.NID.from_parts(base_ids[i:i+16], semantic_hashes[i]) 
        for i in range(1000)]

# Search by semantic similarity
query_hash = hash_embedding(query_embedding)
similar = search_by_semantic_hash(query_hash)
```

## Performance Tips

1. **Use tensor backend** for ML operations
2. **Zero-copy NumPy** for large batches
3. **Time binning** for efficient causal masking
4. **Batch generation** for high-throughput pipelines

## Next Steps

- [Batch Generation](batch-generation.md) - High-throughput patterns
- [Neural IDs](neural-ids.md) - Semantic IDs for ML
- [Boundary Projections](boundary-projections.md) - All projections
- [Performance](../performance/benchmarks.md) - Performance benchmarks

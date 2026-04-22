# Boundary Projections Guide

Tensor, LLM, and storage projections for SNID identifiers.

## Overview

SNID provides canonical boundary projections for:

- **Tensor projections**: Tensor128, Tensor256 for ML/AI workflows
- **LLM formats**: LLMFormatV1, LLMFormatV2 for LLM integration
- **Time binning**: TimeBin for causal masking
- **Binary storage**: Raw byte storage for databases

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

**Layout:**
- `hi`: Bits 0-63 (big-endian)
- `lo`: Bits 64-127 (big-endian)

**Use cases:**
- ML tensor operations
- Time delta calculations
- Efficient binary storage

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

**Layout:**
- `w0`: Bits 0-63 (big-endian)
- `w1`: Bits 64-127 (big-endian)
- `w2`: Bits 128-191 (big-endian)
- `w3`: Bits 192-255 (big-endian)

**Use cases:**
- 32-byte ID families (NID, LID, WID, XID, KID, BID)
- Extended tensor operations
- Large-scale ML pipelines

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

**Structure:**
```json
{
  "kind": "snid",
  "atom": "MAT",
  "timestamp_millis": 1234567890000,
  "machine_or_shard": 12345,
  "sequence": 6789
}
```

**Use cases:**
- LLM tokenization
- AI pipeline metadata
- Temporal reasoning

### LLMFormatV2

Richer AI projection with additional metadata:

```json
{
  "kind": "snid",
  "atom": "MAT",
  "timestamp_millis": 1234567890000,
  "machine_or_shard": 12345,
  "sequence": 6789,
  "ghosted": false,
  "version": "5.3"
}
```

## Time Binning

### TimeBin Projection

Resolution-truncated temporal projection for causal masking.

**Go:**
```go
timeBin := id.TimeBin(resolution_ms)
```

**Rust:**
```rust
let time_bin = id.time_bin(resolution_ms);
```

**Python:**
```python
time_bin = id.time_bin(resolution_ms=1000)
```

**Common Resolutions:**
- 1ms: Millisecond precision
- 10ms: 10-millisecond precision
- 100ms: 100-millisecond precision
- 1s: Second precision
- 1m: Minute precision

**Use cases:**
- Causal masking in transformers
- Time-based partitioning
- Temporal aggregation

## Binary Storage

### BinaryStorage

Raw 16-byte storage form for databases.

**Go:**
```go
bytes := id[:]
```

**Rust:**
```rust
let bytes = id.as_bytes();
```

**Python:**
```python
bytes = id.to_bytes()
```

**Database Storage:**

**PostgreSQL:**
```sql
CREATE TABLE items (
    id BYTEA PRIMARY KEY
);
```

**ClickHouse:**
```sql
CREATE TABLE items (
    id FixedString(16) PRIMARY KEY
) ENGINE = MergeTree()
ORDER BY id;
```

**MySQL:**
```sql
CREATE TABLE items (
    id BINARY(16) PRIMARY KEY
);
```

## Time Delta Calculation

### Tensor Time Delta

Calculate millisecond delta between two tensor projections.

**Python:**
```python
import snid

id1 = snid.SNID.new_fast()
id2 = snid.SNID.new_fast()

delta = snid.SNID.tensor_time_delta(id1.tensor128(), id2.tensor128())
print(f"Time delta: {delta} ms")
```

**Go:**
```go
hi1, lo1 := id1.Tensor128()
hi2, lo2 := id2.Tensor128()
delta := TimeDelta(hi1, lo1, hi2, lo2)
```

## Use Cases

### ML Pipeline Integration

```python
import snid
import numpy as np

# Generate batch as tensor
batch = snid.SNID.generate_batch(10000, backend="numpy")

# Extract timestamps
timestamps = batch[:, 0] >> 16

# Create causal mask
causal_mask = timestamps[:, None] <= timestamps[None, :]
```

### Vector Database Keys

```python
import snid

# Generate IDs for vectors
ids = snid.SNID.generate_batch(1000, backend="bytes")

# Insert into vector database
for i in range(0, len(ids), 16):
    vector_db.insert(ids[i:i+16], embeddings[i//16])
```

### Time-Based Partitioning

```python
import snid

# Generate IDs with time binning
ids = snid.SNID.generate_batch(10000, backend="bytes")

# Partition by hour
for id_bytes in ids:
    id = snid.SNID.from_bytes(id_bytes)
    time_bin = id.time_bin(resolution_ms=3600000)  # 1 hour
    partition_key = f"hour_{time_bin}"
    store_in_partition(partition_key, id_bytes)
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

## Performance

### Tensor Projection Performance

| Operation | Go | Rust | Python |
|-----------|-----|------|--------|
| Tensor128 | ~5ns | ~8ns | ~20ns |
| Tensor256 | ~8ns | ~12ns | ~30ns |
| Time Delta | ~10ns | ~15ns | ~25ns |

### Storage Performance

| Operation | Go | Rust | Python |
|-----------|-----|------|--------|
| to_bytes | ~2ns | ~3ns | ~5ns |
| from_bytes | ~3ns | ~4ns | ~8ns |

## Best Practices

1. **Use tensor backend** for ML operations
2. **Use binary storage** for databases
3. **Use time binning** for causal masking
4. **Choose appropriate resolution** for time binning
5. **Use zero-copy views** in NumPy

## Next Steps

- [AI/ML Integration](ai-ml-integration.md) - AI/ML pipeline guide
- [Batch Generation](batch-generation.md) - High-throughput patterns
- [Neural IDs](neural-ids.md) - Semantic IDs for ML
- [Performance](../performance/benchmarks.md) - Performance benchmarks

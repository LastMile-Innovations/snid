# Batch Generation Guide

High-throughput ID generation patterns for production systems.

## Overview

SNID provides optimized batch generation for high-throughput scenarios:

- **Go**: `NewBatch()` for efficient batch generation
- **Rust**: `generate_batch()` for zero-allocation batches
- **Python**: `generate_batch()` with multiple backends (bytes, tensor, numpy, pyarrow, polars)

## Python Batch Generation

### Raw Bytes Backend (Fastest)

```python
import snid

# Generate 100,000 IDs as raw bytes
batch = snid.SNID.generate_batch(100000, backend="bytes")
print(f"Generated {len(batch)} bytes ({len(batch)//16} IDs)")

# Process as bytes
for i in range(0, len(batch), 16):
    id_bytes = batch[i:i+16]
    # Process ID
```

### Tensor Backend

```python
import snid

# Generate as tensor pairs (hi, lo)
batch = snid.SNID.generate_batch(1000, backend="tensor")
print(f"Generated {len(batch)} tensor pairs")

for hi, lo in batch:
    # Process tensor words
    timestamp_ms = hi >> 16
    print(f"Timestamp: {timestamp_ms}")
```

### NumPy Backend (Zero-Copy)

```python
import snid
import numpy as np

# Generate as NumPy array (zero-copy view)
batch = snid.SNID.generate_batch(10000, backend="numpy")
print(f"Generated NumPy array with shape: {batch.shape}")

# Efficient operations
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
print(f"Generated PyArrow array with {len(batch)} elements")

# Convert to table
table = pa.Table.from_arrays([batch], names=["id"])
print(table)
```

### Polars Backend

```python
import snid
import polars as pl

# Generate as Polars series
batch = snid.SNID.generate_batch(1000, backend="polars")
print(f"Generated Polars series with {len(batch)} elements")

# Create DataFrame
df = pl.DataFrame({"id": batch})
print(df.head())
```

## Go Batch Generation

```go
package main

import (
    "fmt"
    "github.com/neighbor/snid"
)

func main() {
    // Generate batch of IDs
    batch := snid.NewBatch(snid.Matter, 1000)
    fmt.Printf("Generated %d batch IDs\n", len(batch))

    // Process batch
    for i, id := range batch {
        wire := id.String(snid.Matter)
        fmt.Printf("[%d] %s\n", i, wire)
    }
}
```

## Rust Batch Generation

```rust
use snid::SNID;

fn main() {
    // Generate batch of IDs
    let batch = SNID::generate_batch(1000);
    println!("Generated {} batch IDs", batch.len());

    // Process batch
    for (i, id) in batch.iter().enumerate() {
        let wire = id.to_wire("MAT");
        println!("[{}] {}", i, wire);
    }
}
```

## Performance Considerations

### Backend Selection (Python)

| Backend | Speed | Memory | Use Case |
|---------|-------|--------|----------|
| `bytes` | Fastest | Low | Raw storage, network transmission |
| `tensor` | Fast | Low | Tensor operations, ML pipelines |
| `numpy` | Fast | Medium | NumPy workflows, data science |
| `pyarrow` | Medium | Medium | Arrow-based systems |
| `polars` | Medium | Medium | Polars workflows |
| `snid` | Slow | High | Python object wrappers |

### Batch Size Recommendations

- **Small batches** (< 100): Use single ID generation
- **Medium batches** (100-10,000): Use batch generation
- **Large batches** (> 10,000): Use raw bytes backend

### Memory Management

```python
# Process in chunks to avoid memory pressure
chunk_size = 10000
total = 100000

for i in range(0, total, chunk_size):
    count = min(chunk_size, total - i)
    batch = snid.SNID.generate_batch(count, backend="bytes")
    # Process chunk
    process_chunk(batch)
```

## Use Cases

### Database Bulk Insert

```python
import snid
import psycopg2

# Generate batch for bulk insert
batch = snid.SNID.generate_batch(10000, backend="bytes")

conn = psycopg2.connect("...")
cursor = conn.cursor()

# Bulk insert
cursor.executemany(
    "INSERT INTO items (id, name) VALUES (%s, %s)",
    [(batch[i:i+16], f"item_{i}") for i in range(0, len(batch), 16)]
)
conn.commit()
```

### Message Queue Producers

```python
import snid
import kafka

# Generate batch for Kafka
batch = snid.SNID.generate_batch(1000, backend="bytes")

producer = kafka.KafkaProducer(...)
for i in range(0, len(batch), 16):
    producer.send('topic', key=batch[i:i+16], value=b'message')
```

### Data Pipeline Ingestion

```python
import snid
import polars as pl

# Generate batch for data pipeline
batch = snid.SNID.generate_batch(100000, backend="polars")

df = pl.DataFrame({
    "id": batch,
    "timestamp": pl.datetime_range(...)
})

# Write to storage
df.write_parquet("data.parquet")
```

## Concurrency

### Python (Thread-Safe)

```python
import snid
from concurrent.futures import ThreadPoolExecutor

def generate_batch(count):
    return snid.SNID.generate_batch(count, backend="bytes")

with ThreadPoolExecutor(max_workers=4) as executor:
    futures = [executor.submit(generate_batch, 10000) for _ in range(10)]
    batches = [f.result() for f in futures]
```

### Go (Concurrent)

```go
func generateBatches(numBatches, batchSize int) [][]snid.ID {
    results := make([][]snid.ID, numBatches)
    var wg sync.WaitGroup
    wg.Add(numBatches)

    for i := 0; i < numBatches; i++ {
        go func(idx int) {
            defer wg.Done()
            results[idx] = snid.NewBatch(snid.Matter, batchSize)
        }(i)
    }

    wg.Wait()
    return results
}
```

## Next Steps

- [Basic Usage](basic-usage.md) - Common patterns
- [AI/ML Integration](ai-ml-integration.md) - Tensor projections
- [Storage Contracts](storage-contracts.md) - Database integration
- [Performance](../performance/benchmarks.md) - Performance benchmarks

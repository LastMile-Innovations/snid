# Performance Optimization Tips

Guidance for optimizing SNID performance in your applications.

## Overview

SNID is already highly optimized, but there are patterns to get the best performance:

- Use appropriate generation mode for your use case
- Choose the right backend for batch generation
- Avoid unnecessary encoding/decoding
- Use binary storage instead of wire strings
- Leverage lock-free per-P state in Go

## Go Optimization

### Use NewFast() for Single IDs

```go
// Good
id := snid.NewFast()

// Avoid in hot paths
id := snid.New(snid.Matter) // Calls NewFast() internally
```

### Use TurboStreamer for Hot Loops

```go
streamer := snid.NewTurboStreamer()
for i := 0; i < 1000000; i++ {
    id := streamer.Next() // ~1.7ns
}
```

### Use NewBatch() for Bulk Operations

```go
// Good
batch := snid.NewBatch(snid.Matter, 1000)

// Avoid
for i := 0; i < 1000; i++ {
    id := snid.NewFast()
}
```

### Avoid String Conversion in Hot Paths

```go
// Good
id := snid.NewFast()
storeBinary(id[:])

// Avoid
id := snid.NewFast()
wire := id.String(snid.Matter) // ~50ns overhead
storeString(wire)
```

### Use Binary Storage

```go
// Good
_, err := db.Exec("INSERT INTO items (id) VALUES ($1)", id[:])

// Avoid
_, err := db.Exec("INSERT INTO items (id) VALUES ($1)", id.String(snid.Matter))
```

## Rust Optimization

### Use generate_batch() for Bulk Operations

```rust
// Good
let batch = SNID::generate_batch(1000);

// Avoid
let batch: Vec<SNID> = (0..1000).map(|_| SNID::new()).collect();
```

### Enable Release Mode

```bash
cargo test --release
cargo bench --release
```

### Avoid Unnecessary Features

```toml
# Only enable features you need
[dependencies]
snid = { version = "0.2", features = [] }  # No serde by default
```

## Python Optimization

### Choose the Right Backend

```python
# Fastest - raw bytes
batch = snid.SNID.generate_batch(1000, backend="bytes")

# Fast - tensor pairs
batch = snid.SNID.generate_batch(1000, backend="tensor")

# Medium - NumPy (zero-copy)
batch = snid.SNID.generate_batch(1000, backend="numpy")

# Slow - Python objects
batch = snid.SNID.generate_batch(1000, backend="snid")  # Avoid
```

### Use NumPy for Data Science

```python
import snid
import numpy as np

# Good - zero-copy view
batch = snid.SNID.generate_batch(10000, backend="numpy")
df = np.column_stack([batch, features])

# Avoid - Python objects
batch = snid.SNID.generate_batch(10000, backend="snid")
df = np.array([id.to_bytes() for id in batch])
```

### Process in Chunks

```python
# Good - process in chunks
chunk_size = 10000
for i in range(0, total, chunk_size):
    batch = snid.SNID.generate_batch(chunk_size, backend="bytes")
    process_chunk(batch)

# Avoid - generate all at once
batch = snid.SNID.generate_batch(1000000, backend="bytes")  # High memory
```

### Avoid Wire String Conversion

```python
# Good
id = snid.SNID.new_fast()
store_binary(id.to_bytes())

# Avoid
id = snid.SNID.new_fast()
wire = id.to_wire("MAT")  # ~80ns overhead
store_string(wire)
```

## Database Optimization

### Use Binary Storage

```sql
-- Good
CREATE TABLE items (
    id BYTEA PRIMARY KEY
);

-- Avoid
CREATE TABLE items (
    id TEXT PRIMARY KEY
);
```

### Index Appropriately

```sql
-- Good - B-tree index
CREATE INDEX idx_items_id ON items (id);

-- For equality only
CREATE INDEX idx_items_id_hash ON items USING HASH (id);
```

### Batch Insert

```sql
-- Good
INSERT INTO items (id, name) VALUES 
    ($1, $2),
    ($3, $4),
    ($5, $6);

-- Avoid
INSERT INTO items (id, name) VALUES ($1, $2);
INSERT INTO items (id, name) VALUES ($3, $4);
INSERT INTO items (id, name) VALUES ($5, $6);
```

## Concurrency Optimization

### Go - Lock-Free Per-P State

```go
// Good - uses lock-free per-P state
for i := 0; i < 1000; i++ {
    id := snid.NewFast()  // Lock-free
}

// Avoid - shared state contention
var mu sync.Mutex
for i := 0; i < 1000; i++ {
    mu.Lock()
    id := snid.NewFast()
    mu.Unlock()
}
```

### Python - Thread-Safe Generation

```python
from concurrent.futures import ThreadPoolExecutor

# Good - each thread generates independently
def generate_batch(count):
    return snid.SNID.generate_batch(count, backend="bytes")

with ThreadPoolExecutor(max_workers=4) as executor:
    futures = [executor.submit(generate_batch, 10000) for _ in range(10)]
    batches = [f.result() for f in futures]
```

## Memory Optimization

### Reuse Buffers

```go
// Good - reuse buffer
var buf [16]byte
for i := 0; i < 1000; i++ {
    id := snid.NewFast()
    copy(buf[:], id[:])
    process(buf[:])
}

// Avoid - allocate each time
for i := 0; i < 1000; i++ {
    id := snid.NewFast()
    buf := make([]byte, 16)
    copy(buf, id[:])
    process(buf)
}
```

### Use Zero-Copy Views

```python
import snid
import numpy as np

# Good - zero-copy view
batch = snid.SNID.generate_batch(10000, backend="numpy")
# batch is a view, not a copy

# Avoid - copy
batch = snid.SNID.generate_batch(10000, backend="snid")
arr = np.array([id.to_bytes() for id in batch])  # Copy
```

## Profiling

### Go Profiling

```bash
cd go
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

### Rust Profiling

```bash
cd rust
cargo bench -- --profile-time 10
```

### Python Profiling

```bash
cd python
python -m cProfile -s time bench_batch.py
```

## Common Pitfalls

### 1. Using Wire Strings in Hot Paths

```go
// Bad
for i := 0; i < 1000000; i++ {
    id := snid.NewFast()
    wire := id.String(snid.Matter)  // 50ns overhead
    process(wire)
}

// Good
for i := 0; i < 1000000; i++ {
    id := snid.NewFast()
    process(id[:])  // Binary
}
```

### 2. Generating IDs One-by-One in Python

```python
# Bad
ids = [snid.SNID.new_fast() for _ in range(10000)]

# Good
ids = snid.SNID.generate_batch(10000, backend="bytes")
```

### 3. Using Wrong Backend in Python

```python
# Bad - slowest
batch = snid.SNID.generate_batch(10000, backend="snid")

# Good - fastest
batch = snid.SNID.generate_batch(10000, backend="bytes")
```

### 4. Not Using Release Mode in Rust

```bash
# Bad - debug mode
cargo bench

# Good - release mode
cargo bench --release
```

## Next Steps

- [Benchmarks](benchmarks.md) - Performance benchmarks
- [Comparison](comparison.md) - Comparison with other ID systems
- [Basic Usage](../guides/basic-usage.md) - SNID usage patterns

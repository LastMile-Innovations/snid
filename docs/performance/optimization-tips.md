# Performance Optimization Tips

Guidance for optimizing SNID performance in your applications.

## Overview

SNID is already highly optimized, but there are patterns to get the best performance:

**Use first principles and critical deep thinking.** Question assumptions, derive from fundamentals, and don't accept "good enough" as an answer.

- Use appropriate generation mode for your use case
- Choose the right backend for batch generation
- Avoid unnecessary encoding/decoding
- Use binary storage instead of wire strings
- Leverage lock-free per-P state in Go

## Implementation-Level Optimizations

SNID implementations include several low-level optimizations for maximum performance:

### Modern Crypto Dependencies

SNID uses the latest versions of cryptographic libraries which include significant performance improvements:

- **getrandom 0.4**: Uses Rust Edition 2024 with optimized `fill()` API, bringing 75% faster random number generation
- **hmac 0.13**: Uses efficient block-level state representation via digest 0.11, bringing 75% faster HMAC operations
- **sha2 0.11**: Uses hardware-accelerated backends (aarch64-sha2, x86-sha, x86-avx2) when available, with automatic fallback to software implementation

**Performance Impact:**
- ID generation: 75% faster due to getrandom 0.4 optimizations
- HMAC operations (KID, LID, GrantID): 75% faster due to hmac 0.13 block-level API
- Hash operations: Hardware acceleration when available (SHA-NI on x86, sha2 on ARM64)

### Cache-Line Padding

Both Go and Rust implementations use 64-byte cache-line padding on generator state to prevent false sharing in multi-threaded scenarios. This ensures that concurrent threads don't fight over the same cache line when updating different shards.

**Performance Impact:**
- Prevents false sharing when multiple threads generate IDs concurrently
- Core hot paths improved 6-13% (snid_new_fast, snid_to_wire, snid_to_uuid_string)
- Brings Rust implementation to parity with Go's cache-line strategy

### Aggressive Inlining

Hot path functions are marked with `#[inline(always)]` in Rust and similar inlining hints in Go to eliminate function call overhead in critical paths:

- ID generation: `GeneratorState::next()`, `Snid::new()`
- Byte conversion: `to_bytes()`, `from_bytes()`
- Encoding: `encode_payload()`, `decode_payload()`, `crc8()`
- UUID formatting: `encode_uuid_string()`, `decode_uuid_hex()`

**Performance Impact:**
- Eliminates function call overhead in hot paths
- Enables compiler to melt SNID logic directly into user code
- Particularly beneficial for batch operations

### Batch Optimization

Extended identifier families include optimized batch helpers:

```rust
// Nid batch generation with pre-allocation
let batch = Nid::batch_from_head(head, &semantic_hashes);
```

**Performance Impact:**
- Batch operations improved 21-33% (nid_batch_100, nid_hamming_distance)
- Reduces memory allocations in batch operations
- Optimized hamming_distance uses direct byte comparison

## Advanced Optimization Roadmap

SNID's performance optimization is an ongoing journey. The following advanced optimizations are on our roadmap:

### SIMD Base58 Encoding

**Current State**: Base58 encoding uses 128-bit integer division with `bits.Div64` (Go) or similar approach (Rust). This is already optimized compared to byte-by-byte approaches (44 divisions vs 352).

**Optimization Target**: Implement AVX-512 / NEON SIMD vectorized Base58 encoding/decoding for batch operations.

**Implementation Guidance**:
- Target batch encoding operations (encoding arrays of IDs) rather than single ID encoding
- Use AVX-512 on x86-64, NEON on ARM64
- Consider carry-less multiplication for division-by-58 operations
- Benchmark against current 128-bit division approach
- Maintain conformance with existing test vectors

**Expected Impact**: Encode batch arrays in fewer CPU cycles by parallelizing the division operations.

**Current Status**: Planned - see `go/encoding.go` and `rust/src/encoding.rs`

### Fork-safe RNG Entropy

**Current State**: RNG state is seeded at initialization using process-unique entropy (PID, timestamp, maphash).

**Optimization Target**: Implement state-check counters to instantly detect a forked process in production.

**Implementation Guidance**:
- Add generation counter to each shard state
- On each ID generation, validate counter against expected range
- If counter indicates potential fork, re-seed RNG state immediately
- Use atomic operations for counter checks to minimize overhead
- Add fork detection tests to conformance suite

**Expected Impact**: Prevent duplicate ID generation after fork events without performance penalty in normal operation.

**Current Status**: Planned - see `go/generator.go` shard structures

### Adaptive Cache-Line Padding

**Current State**: Go implementations already use 64-byte padding on `shard`, `fastPShard`, `adaptiveShard`, and `alignedClock` structures.

**Optimization Target**: Ensure Rust implementation has equivalent cache-line padding and validate multi-core scaling.

**Implementation Guidance**:
- Verify Rust generator state has 64-byte padding
- Add cache-line padding benchmarks
- Test with varying core counts (1, 2, 4, 8, 16, 32+)
- Measure false-sharing using perf counters
- Document padding strategy in architecture docs

**Expected Impact**: Eliminate false-sharing cache thrashing in high-concurrency scenarios.

**Current Status**: Partially complete in Go, needs Rust validation

### Constant-time MAC tails

**Current State**: KID and LID verification uses standard comparison operations.

**Optimization Target**: Harden signature verification into constant-time comparisons to prevent timing-sidechannel attacks.

**Implementation Guidance**:
- Use `subtle.ConstantTimeCompare` in Go
- Use `subtle` crate in Rust
- Ensure all MAC verification paths are constant-time
- Add timing attack resistance tests
- Document security guarantees

**Expected Impact**: Eliminate timing side-channel vulnerabilities in verification paths.

**Current Status**: Planned - see `go/akid.go` and `rust/src/akid.rs`

### PGO (Profile-Guided Optimization)

**Current State**: Standard compiler optimizations used in release builds.

**Optimization Target**: Generate PGO profiles during CI/CD benchmark runs and feed back into compiler.

**Implementation Guidance**:
- Add PGO profile collection to CI benchmark workflows
- Run representative workloads (single ID, batch, concurrent)
- Generate PGO profiles for Go (`go test -cpuprofile`) and Rust (`cargo-pgo`)
- Integrate PGO into release build process
- Document PGO workflow for contributors

**Expected Impact**: Compiler optimizes hot loops for exact branch probabilities observed in production workloads.

**Current Status**: Planned - see `.github/workflows/`

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

# Debugging

Tips and tools for debugging SNID issues.

## Overview

This guide covers debugging techniques for SNID across Go, Rust, and Python.

## Go Debugging

### Using Delve

```bash
cd go
dlv test
```

### Adding Debug Logging

```go
import "log"

id := snid.NewFast()
log.Printf("Generated ID: %x", id[:])
```

### Checking Byte Layout

```go
id := snid.NewFast()
fmt.Printf("Bytes: %x\n", id[:])
fmt.Printf("UUID: %s\n", id.UUID())
fmt.Printf("Wire: %s\n", id.String(snid.Matter))
```

### Profiling

```bash
cd go
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

### Memory Profiling

```bash
go test -bench=. -memprofile=mem.prof
go tool pprof mem.prof
```

## Rust Debugging

### Using GDB

```bash
cd rust
cargo build
gdb target/debug/snid
```

### Adding Debug Logging

```rust
use log::debug;

let id = SNID::new();
debug!("Generated ID: {:?}", id.as_bytes());
```

### Using println! for Quick Debugging

```rust
let id = SNID::new();
println!("Bytes: {:?}", id.as_bytes());
println!("Wire: {}", id.to_wire("MAT"));
```

### Profiling

```bash
cd rust
cargo bench -- --profile-time 10
```

### Flamegraph

```bash
cargo install flamegraph
cargo flamegraph
```

## Python Debugging

### Using pdb

```python
import pdb; pdb.set_trace()

id = snid.SNID.new_fast()
```

### Using ipdb

```bash
pip install ipdb
```

```python
import ipdb; ipdb.set_trace()
```

### Adding Debug Logging

```python
import logging

logging.basicConfig(level=logging.DEBUG)
id = snid.SNID.new_fast()
logging.debug(f"Generated ID: {id.to_bytes().hex()}")
```

### Checking Byte Layout

```python
id = snid.SNID.new_fast()
print(f"Bytes: {id.to_bytes().hex()}")
print(f"UUID: {id.to_uuid()}")
print(f"Wire: {id.to_wire('MAT')}")
```

### Profiling

```bash
cd python
python -m cProfile -s time bench_batch.py
```

### Memory Profiling

```bash
pip install memory_profiler
python -m memory_profiler bench_batch.py
```

## Conformance Debugging

### Check Vector Generation

```bash
cd conformance/cmd/generate_vectors
go run . --out ../../vectors.json
```

### Inspect Vectors

```bash
cat conformance/vectors.json | jq '.vectors[0]'
```

### Validate Specific Vector

```python
import snid
import json

with open('conformance/vectors.json') as f:
    vectors = json.load(f)

for v in vectors['vectors']:
    parsed, atom = snid.SNID.parse_wire(v['output']['wire'])
    print(f"Validated: {v['name']}")
```

## Database Debugging

### Check Binary Storage

```sql
-- Check stored bytes
SELECT id, encode(id, 'hex') FROM items LIMIT 10;
```

### Check Wire Format

```sql
-- Check wire format
SELECT id::text FROM items LIMIT 10;
```

### Verify Encoding

```go
// Check roundtrip
id := snid.NewFast()
wire := id.String(snid.Matter)
parsed, atom, err := snid.FromString(wire)
if err != nil {
    log.Fatal(err)
}
if parsed != id {
    log.Fatal("Roundtrip failed")
}
```

## Performance Debugging

### Benchmark Specific Operation

**Go:**
```bash
cd go
go test -bench=NewFast -benchmem
```

**Rust:**
```bash
cd rust
cargo bench new
```

**Python:**
```python
import timeit

timeit.timeit('snid.SNID.new_fast()', setup='import snid', number=100000)
```

### Compare Backends

```python
import snid
import time

start = time.time()
batch = snid.SNID.generate_batch(10000, backend="bytes")
print(f"bytes: {time.time() - start:.4f}s")

start = time.time()
batch = snid.SNID.generate_batch(10000, backend="numpy")
print(f"numpy: {time.time() - start:.4f}s")
```

## Common Debugging Scenarios

### ID Generation Not Working

**Symptom:** IDs not generating or always the same.

**Debug Steps:**
1. Check clock is advancing: `time.Now().UnixMilli()`
2. Check sequence overflow
3. Check machine fingerprint
4. Verify conformance tests pass

### Wire Format Not Parsing

**Symptom:** `FromString` or `parse_wire` failing.

**Debug Steps:**
1. Check wire string format: `ATOM:payload`
2. Verify Base58 characters
3. Check checksum
4. Verify atom is valid

### Batch Generation Slow

**Symptom:** Batch generation slower than expected.

**Debug Steps:**
1. Check backend selection
2. Profile with cProfile
3. Check memory usage
4. Use correct backend for use case

### Conformance Test Failing

**Symptom:** Conformance tests fail for one implementation.

**Debug Steps:**
1. Regenerate vectors: `just conformance`
2. Check encoding/decoding logic
3. Verify byte layout matches SPEC.md
4. Check for platform-specific issues

## Tools

### Go Tools

- `dlv` - Debugger
- `pprof` - Profiler
- `go test` - Testing
- `gofmt` - Formatting
- `golangci-lint` - Linting

### Rust Tools

- `gdb` - Debugger
- `cargo bench` - Benchmarking
- `cargo clippy` - Linting
- `cargo fmt` - Formatting
- `flamegraph` - Flamegraph profiling

### Python Tools

- `pdb` - Debugger
- `ipdb` - Enhanced debugger
- `cProfile` - Profiler
- `memory_profiler` - Memory profiling
- `ruff` - Linting and formatting

## Logging

### Go Logging

```go
import "log"

log.SetFlags(log.LstdFlags | log.Lshortfile)
log.Printf("Debug info: %v", id)
```

### Rust Logging

```toml
# Cargo.toml
[dependencies]
log = "0.4"
env_logger = "0.10"
```

```rust
use log::debug;

env_logger::init();
debug!("Debug info: {:?}", id);
```

### Python Logging

```python
import logging

logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logging.debug(f"Debug info: {id}")
```

## Testing

### Unit Tests

**Go:**
```go
func TestNewFast(t *testing.T) {
    id := snid.NewFast()
    if id == [16]byte{} {
        t.Fatal("ID is zero")
    }
}
```

**Rust:**
```rust
#[test]
fn test_new() {
    let id = SNID::new();
    assert_ne!(id.as_bytes(), &[0u8; 16]);
}
```

**Python:**
```python
def test_new_fast():
    id = snid.SNID.new_fast()
    assert id.to_bytes() != b'\x00' * 16
```

### Integration Tests

Test database integration, API compatibility, etc.

## Getting Help

If you're stuck:

1. Check [Common Errors](common-errors.md) for known issues
2. Check [FAQ](faq.md) for common questions
3. Search [GitHub Issues](https://github.com/LastMile-Innovations/snid/issues)
4. Create a new issue with:
   - Error message
   - Code snippet
   - Environment details
   - Debugging steps taken

## Next Steps

- [Common Errors](common-errors.md) - Error troubleshooting
- [FAQ](faq.md) - Frequently asked questions
- [Contributing](../../CONTRIBUTING.md) - Contribution guidelines

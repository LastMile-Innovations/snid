# Generator Design

High-performance ID generation architecture.

## Overview

SNID v5.3 FINAL generator provides:
- ~3.7ns latency for `NewFast()` (single ID, thread-safe)
- ~1.7ns for `TurboStreamer.Next()` (hot loop, single-thread)
- ~2μs for `NewBurst(1000)` (batch mode)
- Lock-free per-P state when runtime pinning is available
- Adaptive coarse clock (10ms to 500μs based on GOMAXPROCS)

## Byte Layout

```
Bits 0-47:     Unix timestamp (ms)
Bits 48-51:    Version nibble (0b0111)
Bits 52-65:    Monotonic sequence (14 bits)
Bits 66-89:    Machine/process fingerprint (24 bits)
Bits 90-127:   Entropy tail (38 bits)
```

## Generator State

### Per-P Shard

Each P (processor) has its own shard for lock-free generation:

```go
type fastPShard struct {
    _ [64]byte          // Front padding: prevents false sharing
    lastTime uint64
    sequence uint16
    entropy  uint64
    _ [64]byte          // Back padding: prevents false sharing
}
```

### Shared Shard

Fallback for systems without runtime pinning:

```go
type shard struct {
    _ [64]byte          // Front padding: prevents false sharing
    mu       sync.Mutex
    lastTime uint64
    sequence uint16
    entropy  uint64
    _ [64]byte          // Back padding: prevents false sharing
}
```

### Cache-Line Padding

Both Go and Rust implementations use 64-byte cache-line padding to prevent false sharing in multi-threaded scenarios. Each shard is padded with 64 bytes before and after the hot state fields to ensure that concurrent threads don't fight over the same cache line when updating different shards.

**Performance Impact:**
- Prevents false sharing when multiple threads generate IDs concurrently
- Core hot paths improved 6-13% (snid_new_fast, snid_to_wire, snid_to_uuid_string)
- Brings Rust implementation to parity with Go's cache-line strategy

## Clock Strategy

### Coarse Clock

Adaptive tick rate based on GOMAXPROCS:
- 1-2 CPUs: 10ms tick
- 3-4 CPUs: 5ms tick
- 5-8 CPUs: 2ms tick
- 9-16 CPUs: 1ms tick
- 17+ CPUs: 500μs tick

### Virtual Time Advancement

When the coarse clock hasn't advanced, the generator uses virtual time:

```go
if now <= shard.lastTime {
    shard.sequence++
    if shard.sequence > maxSequence {
        // Wait for clock to advance
    }
} else {
    shard.lastTime = now
    shard.sequence = 0
}
```

## Generation Modes

### NewFast()

Default thread-safe ID generation:

```go
func NewFast() ID {
    p := runtime_pin()
    if p >= 0 && p < len(fastShards) {
        return fastShards[p].next()
    }
    return sharedShard.next()
}
```

### NewProjected()

Tenant-aware generation with shard:

```go
func NewProjected(tenantID string, shard uint16) ID {
    // Uses tenant-specific shard
}
```

### NewBatch()

Batch generation for high throughput:

```go
func NewBatch(atom Atom, count int) []ID {
    // Pre-allocates and fills batch
}
```

### TurboStreamer

Hot loop single-threaded generation:

```go
type TurboStreamer struct {
    lastTime uint64
    sequence uint16
    entropy  uint64
}

func (t *TurboStreamer) Next() ID {
    // No locking, no atomics
}
```

## Machine Fingerprint

24-bit machine/process fingerprint for collision resistance:

```go
func fingerprint() uint32 {
    var fp uint32
    // Hash of hostname, process ID, etc.
    return fp & 0xFFFFFF
}
```

## Entropy Generation

Cryptographically secure entropy for the tail:

```go
func entropy() uint64 {
    // Uses crypto/rand or fast entropy source
}
```

## Performance Targets

| Operation | Target | Notes |
|-----------|--------|-------|
| NewFast() | ~3.7ns | Thread-safe, lock-free per-P |
| TurboStreamer.Next() | ~1.7ns | Hot loop, single-thread |
| NewBurst(1000) | ~2μs | Batch mode |
| NewProjected() | ~5ns | Tenant-aware |

## Concurrency Model

### Lock-Free Path

When runtime pinning is available:
- Each P has its own shard
- No locks or atomics
- True lock-free generation

### Fallback Path

When runtime pinning is not available:
- Shared shard with mutex
- Still fast (~5ns)
- Safe for concurrent use

## Sequence Overflow

When sequence overflows (16384 IDs per ms):

```go
if shard.sequence > maxSequence {
    // Wait for clock to advance
    time.Sleep(time.Until(nextTick))
}
```

## Clock Drift Handling

The generator handles clock drift by:
- Using monotonic clock when available
- Advancing virtual time when coarse clock stalls
- Detecting clock backwards and waiting

## Implementation Details

### Go

See `go/generator.go` for full implementation.

### Rust

See `rust/src/lib.rs` for deterministic core.

### Python

Python bindings use Rust core for generation.

## Next Steps

- [Encoding Design](encoding-design.md) - Base58 encoding
- [Conformance Design](conformance-design.md) - Cross-language conformance
- [Diagrams](diagrams.md) - Architecture diagrams

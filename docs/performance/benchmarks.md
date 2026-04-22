# Performance Benchmarks

Performance benchmarks for SNID across implementations with a portable, cloud-agnostic benchmarking platform.

## Overview

SNID is optimized for high-performance ID generation:

- **Go**: ~3.7ns per ID (NewFast, lock-free per-P state), ~1.7ns (TurboStreamer hot loop)
- **Rust**: ~6.4ns per ID (deterministic core, thread-local state)
- **Python**: ~15ns per ID (native bindings), ~6.4ms for 100k batch (bytes backend)

## Benchmarking Platform (Phase 2)

SNID includes a portable benchmarking platform that runs anywhere via Docker with zero harness overhead.

### Key Features

- **Pure Mode**: Benchmarks run in isolated subprocess with no dashboard/logging overhead
- **Web Dashboard**: FastAPI UI for triggering and viewing results
- **Statistical Analysis**: Mean, median, p95, p99, stddev, 95% confidence intervals
- **HTML Reports**: Auto-generated reports with trend charts
- **Regression Detection**: Flags performance regressions >10% (configurable)
- **Property-Based Testing**: Hypothesis (Python) and proptest (Rust) for invariants
- **Cloud Deployment**: Railway-optimized with volume persistence
- **Nightly Automation**: GitHub Actions scheduled runs with alerting

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Railway Container                         │
│                                                               │
│  ┌────────────────────┐         ┌──────────────────────┐    │
│  │   FastAPI App      │         │   Isolated Runner    │    │
│  │   (Dashboard + API)│ ──────▶ │   (Pure Mode)        │    │
│  │                    │  spawn  │                      │    │
│  └────────────────────┘         └──────────┬───────────┘    │
│                                            │                │
│                                    Writes results to disk   │
│                                    (after benchmark ends)   │
└─────────────────────────────────────────────────────────────┘
```

## Running Benchmarks

### Local Development

#### Using just (Recommended)

```bash
# Run all benchmarks
just bench

# Run specific language
just bench-go
just bench-rust
just bench-python

# Run comparison benchmarks
just bench-comparison

# Run LLM token efficiency
just bench-llm
```

#### Direct Commands

```bash
# Go
cd go && go test -bench=. -benchmem -count=5

# Rust
cd rust && cargo bench -- --save-baseline main

# Python
cd python && python -m pytest tests/test_bench.py --benchmark-only
```

### Docker Platform

#### Web Dashboard Mode

```bash
cd benchmarks
docker-compose up -d
```

Access dashboard at http://localhost:8080

#### CLI Mode (One-off Runs)

```bash
cd benchmarks
docker-compose --profile cli up snid-benchmarks-cli
```

#### Pure Mode (Zero Overhead)

The platform uses **pure mode** by default to ensure the benchmarking harness does not affect results:

- Benchmarks run in isolated subprocess with no FastAPI/dashboard code loaded
- Result files written only after benchmark completion
- No logging or metrics collection during measurement

Enable via environment variable:
```bash
BENCH_PURE_MODE=true python benchmarks/runner.py all
```

### Railway Deployment

```bash
# Deploy to Railway
just railway-deploy

# Run one-off benchmark
just railway-run

# View logs
just railway-logs
```

See [deploy/RAILWAY.md](../../deploy/RAILWAY.md) for detailed setup.

## Benchmark Results

Latest benchmark results (April 2026), Apple Silicon, Go package `github.com/LastMile-Innovations/snid`.

### Single ID Generation

| Implementation | Operation | Latency | Notes |
|----------------|-----------|---------|-------|
| Go | NewFast() | ~3.7ns | Lock-free per-P state, 0 allocs |
| Go | NewUUIDv7() | ~3.7ns | UUIDv7-compatible wrapper |
| Go | TurboStreamer.Next() | ~1.7ns | Hot loop, single-thread |
| Rust | new_fast() | ~6.4ns | Thread-local state |
| Rust | uuidv7() | ~6.4ns | UUIDv7-compatible wrapper |
| Python | new_fast() | ~15ns | Native bindings |

### Batch Generation

| Implementation | Operation | Count | Total Time | Per ID |
|----------------|-----------|-------|------------|--------|
| Go | NewBurst(100) | 100 | ~200ns | ~2ns |
| Go | NewBurst(1000) | 1000 | ~2μs | ~2ns |
| Go | NewBurst(10000) | 10000 | ~20μs | ~2ns |
| Rust | generate_binary_batch(10k) | 10000 | ~30μs | ~3ns |
| Python | generate_batch(bytes, 100k) | 100000 | 6.4ms | 64ns |
| Python | generate_batch(tensor, 100k) | 100000 | 19.6ms | 196ns |
| Python | generate_batch(numpy, 100k) | 100000 | 9.8ms | 98ns |

### Encoding/Decoding

| Implementation | Operation | Latency | Allocation |
|----------------|-----------|---------|------------|
| Go | String() | ~100ns | 48 B, 1 alloc |
| Go | StringCompact() | ~98ns | 24 B, 1 alloc |
| Go | AppendTo() | ~87ns | 0 B, 0 allocs |
| Go | FromString() | ~157ns | 0 B, 0 allocs |
| Go | ParseCompact() | ~151ns | 0 B, 0 allocs |
| Go | ParseUUIDString() | ~150ns | 0 B, 0 allocs |
| Go | Base58 encode, 8 bytes | ~51ns | 16 B, 1 alloc |
| Go | Base58 decode, 8 bytes | ~42ns | 8 B, 1 alloc |
| Go | Base58 encode, 24 bytes | ~787ns | 48 B, 1 alloc |
| Go | Base58 decode, 24 bytes | ~321ns | 24 B, 1 alloc |
| Go | EncodeAKIDSecret() | ~710ns | 96 B, 2 allocs |
| Go | VerifyAKIDSecretChecksum() | ~320ns | 24 B, 1 alloc |
| Go | ParseAKID() | ~510ns | 56 B, 2 allocs |
| Rust | to_wire() | ~206ns | n/a |
| Rust | to_uuid_string() | ~28ns | n/a |
| Python | to_wire() | ~80ns | n/a |
| Python | parse_wire() | ~150ns | n/a |

### Extended Types

| Implementation | Operation | Latency | Notes |
|----------------|-----------|---------|-------|
| Go | LID Verify (parallel) | ~320ns | HMAC verification |
| Go | BID WireFormat | ~500ns | Content-addressable |
| Go | EID New | ~1.7ns | Extremely fast, 0 allocs |
| Go | ToTensorWords | ~1.67ns | Tensor projection |
| Go | ToLLMFormat | ~28ns | LLM projection |
| Go | SGID NewSpatialPrecise | ~250ns | H3 cell encoding, 24 B/op |
| Go | NID NewNeural | ~184ns | Neural ID creation, 0 allocs |
| Go | BID NewBIDFromContent | ~84ns | BLAKE3 hashing, 0 allocs |
| Go | EID Bytes | ~0.25ns | Zero-copy conversion |
| Go | EID Time | ~0.24ns | Timestamp extraction |
| Go | EID String | ~5.2ns | String formatting |
| Rust | NID new | ~8.9ns | Neural ID creation |
| Rust | NID hamming_distance | ~1.68ns | Very fast (595 Melem/s) |
| Rust | BID new | ~13.7ns | Content-addressable |
| Rust | BID wire | ~384ns | Wire format (2.6 Melem/s) |
| Rust | BID parse_wire | ~208ns | Parse wire (4.8 Melem/s) |
| Rust | EID from_parts | ~695ps | Extremely fast (1.44 Gelem/s) |
| Rust | EID to_bytes | ~1.57ns | Byte conversion (636 Melem/s) |
| Rust | EID counter | ~483ps | Counter extraction (2.07 Gelem/s) |
| Rust | EID timestamp_millis | ~436ps | Timestamp extraction (2.29 Gelem/s) |
| Rust | LID from_parts | ~602ns | HMAC verification (1.66 Melem/s) |
| Rust | LID head | ~6.4ns | Head extraction (157 Melem/s) |

### Extended Family Batch Operations

| Implementation | Operation | Count | Total Time | Per ID |
|----------------|-----------|-------|------------|--------|
| Go | EID Batch (100) | 100 | ~169ns | ~1.7ns |
| Go | EID Batch (1000) | 1000 | ~1.7μs | ~1.7ns |
| Go | SGID Batch (100) | 100 | ~25μs | ~250ns |
| Go | NID Batch (100) | 100 | ~18μs | ~180ns |
| Rust | NID Batch (100) | 100 | ~414ns | ~4.1ns |
| Rust | NID Batch (1000) | 1000 | ~4.2μs | ~4.2ns |
| Rust | EID Batch (100) | 100 | ~34ns | ~0.34ns |
| Rust | EID Batch (1000) | 1000 | ~316ns | ~0.32ns |

### Concurrency Stress Tests

| Scenario | Workers | Throughput | Latency | Allocations |
|----------|---------|------------|---------|-------------|
| High Contention | 10 | 5.7M ops/sec | 1.7μs | 240 B/op |
| High Contention | 50 | 8.1M ops/sec | 6.1μs | 1.2 KB/op |
| High Contention | 100 | 8.4M ops/sec | 11.9μs | 2.4 KB/op |
| High Contention | 200 | 8.8M ops/sec | 22.8μs | 4.8 KB/op |
| Cache Line (shared) | N/A | 5.6M ops/sec | 1.8μs | 240 B/op |
| Cache Line (per-worker) | N/A | 6.0M ops/sec | 1.7μs | 240 B/op |
| Sustained Load (5s) | N/A | 25.3M ops/sec | N/A | 96 B/op |
| Batch (10) | N/A | 51M ops/sec | 1.9μs | 240 B/op |
| Batch (100) | N/A | 227M ops/sec | 4.4μs | 240 B/op |
| Batch (1000) | N/A | 492M ops/sec | 20.3μs | 240 B/op |
| Batch (10000) | N/A | 756M ops/sec | 129.6μs | 240 B/op |
| TurboStreamer (1 worker) | 1 | 2.1M ops/sec | 486ns | 1.1 KB/op |
| TurboStreamer (2 workers) | 2 | 1.9M ops/sec | 1.0μs | 2.3 KB/op |
| TurboStreamer (4 workers) | 4 | 2.2M ops/sec | 1.8μs | 4.6 KB/op |
| TurboStreamer (8 workers) | 8 | 2.7M ops/sec | 3.0μs | 9.2 KB/op |
| AdaptiveStreamer (1 worker) | 1 | 4.3M ops/sec | 233ns | 32 B/op |
| AdaptiveStreamer (2 workers) | 2 | 4.5M ops/sec | 441ns | 64 B/op |
| AdaptiveStreamer (4 workers) | 4 | 4.7M ops/sec | 849ns | 128 B/op |
| AdaptiveStreamer (8 workers) | 8 | 5.5M ops/sec | 1.5μs | 256 B/op |

### Memory Profiling

| Operation | Latency | Bytes/op | Allocs/op | Notes |
|-----------|---------|----------|-----------|-------|
| Heap Allocation (NewFast) | 4.0ns | 0 B | 0 allocs | Zero-alloc |
| Stack vs Heap (stack) | 3.9ns | 0 B | 0 allocs | No difference |
| Stack vs Heap (heap) | 3.9ns | 0 B | 0 allocs | Escape analysis working |
| GC Pressure (5s sustained) | N/A | 0 B | 0 allocs | 0 GC/sec |
| Batch (10) | 56ns | 16 B | 1 alloc | Slice overhead |
| Batch (100) | 496ns | 18 B | 1 alloc | Slice overhead |
| Batch (1000) | 4.7μs | 16 B | 1 alloc | Slice overhead |
| Batch (10000) | 44.9μs | 16 B | 1 alloc | Slice overhead |
| Streamer Init | 2.7μs | 65.5 KB | 1 alloc | One-time cost |
| Streamer Next | 3.1ns | 0 B | 0 allocs | Zero-alloc after init |
| TurboStreamer Init | 140ns | 1.1 KB | 2 allocs | One-time cost |
| TurboStreamer Next | 1.6ns | 0 B | 0 allocs | Zero-alloc after init |
| String Compact | 101ns | 24 B | 1 alloc | Base58 encoding |
| String Canonical | 104ns | 48 B | 1 alloc | UUID format |
| Escape Analysis (no escape) | 3.8ns | 0 B | 0 allocs | Stack allocation |
| Escape Analysis (escape) | 3.9ns | 0 B | 0 allocs | Still optimized |

## Statistical Analysis

The benchmarking platform provides statistical rigor:

- **Mean**: Average across samples
- **Median**: 50th percentile
- **p95/p99**: 95th/99th percentiles
- **StdDev**: Standard deviation
- **95% CI**: Confidence interval using t-distribution

### Generating Reports

```bash
# Generate HTML report with trend charts
python benchmarks/report_generator.py

# Check for regressions
python benchmarks/regression_detector.py

# Clean up old results (90-day retention)
python benchmarks/cleanup_results.py --days 90 --keep-last 10
```

## Property-Based Testing

### Python (Hypothesis)

```bash
cd python
python -m pytest tests/test_properties.py
```

Tests invariants:
- Wire format roundtrip
- Sorting invariants
- Batch uniqueness
- Time monotonicity
- Base58/Base32 encoding roundtrip
- Version/variant bits

### Rust (proptest)

```bash
cd rust
cargo test --test property_tests
```

Matching invariants for cross-language validation.

## Optimization Tips

### Go

- Use `NewFast()` for single-threaded generation
- Use `TurboStreamer` for hot loops
- Use `NewBatch()` for batch generation
- Avoid `New()` in hot paths (uses `NewFast()` internally)

### Rust

- Use `generate_batch()` for bulk operations
- Enable release mode for production: `cargo test --release`
- Use `--features serde` only when needed

### Python

- Use `backend="bytes"` for fastest batch generation
- Use `backend="numpy"` for NumPy workflows (zero-copy)
- Use `backend="tensor"` for tensor operations
- Avoid `backend="snid"` in hot paths (slowest)

## Performance Comparison

### vs UUID v4

| Metric | UUID v4 | SNID |
|--------|---------|------|
| Generation | ~50ns | 4.106ns (Go) |
| Ordering | Random | Time-ordered |
| Collisions | Possible | Extremely unlikely |
| Size | 16 bytes | 16 bytes |

### vs ULID

| Metric | ULID | SNID |
|--------|------|------|
| Generation | ~50ns | 4.106ns (Go) |
| Encoding | Base32 | Base58 |
| Time precision | Milliseconds | Milliseconds |
| Extended families | No | Yes |

### vs KSUID

| Metric | KSUID | SNID |
|--------|-------|------|
| Generation | ~100ns | 4.106ns (Go) |
| Size | 20 bytes | 16 bytes |
| Time precision | Seconds | Milliseconds |
| Extended families | No | Yes |

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

## Memory Usage

### Go

- Single ID: 16 bytes
- Batch (1000): ~16KB
- Per-P shard: ~32 bytes per P

### Rust

- Single ID: 16 bytes
- Batch (1000): ~16KB
- Zero-copy where possible

### Python

- Single ID: ~32 bytes (object overhead)
- Batch (bytes): 16KB
- Batch (numpy): ~16KB (zero-copy view)

## Scaling

### Throughput

| Implementation | IDs/sec (single) | IDs/sec (batch) |
|----------------|-----------------|----------------|
| Go | ~270M | ~500M |
| Rust | ~156M | ~330M |
| Python | ~67M | ~15.7M (bytes) / 5.1M (tensor) / 10.2M (numpy) |

### Concurrency

- Go: Lock-free per-P state scales with CPU count
- Rust: Deterministic, scales with CPU count
- Python: GIL limits, but batch generation is efficient

## Next Steps

- [Comparison](comparison.md) - Detailed comparison with other ID systems
- [Optimization Tips](optimization-tips.md) - Performance optimization guidance
- [Implementation Tracks](../IMPLEMENTATION_TRACKS.md) - Performance workstreams

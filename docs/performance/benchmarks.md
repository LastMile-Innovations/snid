# Performance Benchmarks

Performance benchmarks for SNID across implementations with a portable, cloud-agnostic benchmarking platform.

## Overview

SNID is optimized for high-performance ID generation:

- **Go**: ~3.7ns per ID (NewFast), ~1.7ns (TurboStreamer hot loop)
- **Rust**: ~5ns per ID (deterministic core)
- **Python**: ~15ns per ID (native bindings), ~5μs for 1000 batch (bytes backend)

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

### Single ID Generation

| Implementation | Operation | Latency | Notes |
|----------------|-----------|---------|-------|
| Go | NewFast() | ~3.7ns | Lock-free per-P state |
| Go | TurboStreamer.Next() | ~1.7ns | Hot loop, single-thread |
| Rust | new() | ~5ns | Deterministic core |
| Python | new_fast() | ~15ns | Native bindings |

### Batch Generation

| Implementation | Operation | Count | Total Time | Per ID |
|----------------|-----------|-------|------------|--------|
| Go | NewBatch() | 1000 | ~2μs | ~2ns |
| Rust | generate_batch() | 1000 | ~3μs | ~3ns |
| Python | generate_batch(bytes) | 1000 | ~5μs | ~5ns |
| Python | generate_batch(tensor) | 1000 | ~8μs | ~8ns |
| Python | generate_batch(numpy) | 1000 | ~10μs | ~10ns |

### Encoding/Decoding

| Implementation | Operation | Latency |
|----------------|-----------|---------|
| Go | String() | ~50ns |
| Go | FromString() | ~100ns |
| Rust | to_wire() | ~60ns |
| Rust | parse_wire() | ~120ns |
| Python | to_wire() | ~80ns |
| Python | parse_wire() | ~150ns |

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
| Generation | ~50ns | ~3.7ns (Go) |
| Ordering | Random | Time-ordered |
| Collisions | Possible | Extremely unlikely |
| Size | 16 bytes | 16 bytes |

### vs ULID

| Metric | ULID | SNID |
|--------|------|------|
| Generation | ~50ns | ~3.7ns (Go) |
| Encoding | Base32 | Base58 |
| Time precision | Milliseconds | Milliseconds |
| Extended families | No | Yes |

### vs KSUID

| Metric | KSUID | SNID |
|--------|-------|------|
| Generation | ~100ns | ~3.7ns (Go) |
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
| Rust | ~200M | ~330M |
| Python | ~67M | ~200M |

### Concurrency

- Go: Lock-free per-P state scales with CPU count
- Rust: Deterministic, scales with CPU count
- Python: GIL limits, but batch generation is efficient

## Next Steps

- [Comparison](comparison.md) - Detailed comparison with other ID systems
- [Optimization Tips](optimization-tips.md) - Performance optimization guidance
- [Implementation Tracks](../IMPLEMENTATION_TRACKS.md) - Performance workstreams

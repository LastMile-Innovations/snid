# SNID Monorepo

Polyglot SNID protocol and reference implementation repository for Go, Rust, and Python.

## Layout

- `docs/`: canonical specification and topology guidance
- `go/`: standalone Go reference implementation
- `rust/`: deterministic Rust core
- `python/`: PyO3 bindings and Python wrapper
- `conformance/`: vector generation and cross-language checks

## Conformance flow

1. Generate vectors with Go.
2. Validate vectors with Rust.
3. Validate vectors with Python.
4. Fail the build on any divergence.

## Boundary APIs

The repo standardizes canonical wire strings plus AI, tensor, and storage projections:

- `Tensor128`: `[hi:int64, lo:int64]` big-endian tensor words
- `Tensor256`: four big-endian `int64` words for 32-byte ID families
- `LLMFormatV1`: `[ATOM, timestamp_ms, machine_or_shard, sequence]`
- `LLMFormatV2`: richer AI-facing projection for temporal or spatial pipelines
- `TimeBin`: resolution-truncated temporal projection
- `BinaryStorage`: raw 16-byte storage form, with hex fallback only when bytes are not supported

Extended ID families currently represented in this repo:

- `SNID`, `SGID`, `NID`, `LID`, `EID`, `BID`
- `WID`, `XID`, `KID`
- `AKID` dual-part public-plus-secret credentials

Notable entry points:

- `go/boundary.go`
- `go/projections.go`
- `go/composite.go`
- `go/neo4j/adapter.go`
- `rust/src/lib.rs`
- `python/snid/__init__.py`
- `python/snid/neo4j.py`

Protocol and contract docs:

- `docs/SPEC.md`
- `docs/TOPOLOGIES.md`
- `docs/IMPLEMENTATION_TRACKS.md`
- `docs/INTEGRATION_CONTRACTS.md`

## Python Native Path

The Python binding exposes native batch generation for high-throughput ingestion:

- `SNID.generate_batch(n, backend="bytes")` returns contiguous native bytes
- `SNID.generate_batch(n, backend="tensor")` returns decoded tensor pairs
- `SNID.generate_batch(n, backend="numpy")` creates a zero-copy NumPy view from native tensor bytes
- `SNID.tensor_time_delta(left, right)` computes millisecond deltas from tensor words

To run the extension-backed Python tests locally, build the native module in `python/` with `maturin develop` or your preferred PyO3 build/install path.

Quick benchmark entry points:

- Go: `cd go && go test -bench 'Boundary|Deterministic'`
- Python: `cd python && python3 bench_batch.py`

## Repo Hygiene

See `WORKTREE_HYGIENE.md` for what should stay out of diffs and how to keep the repo portable.

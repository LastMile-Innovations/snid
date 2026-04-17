# SNID Implementation Tracks

This document maps the target architecture into execution buckets. The protocol spec remains stable; these tracks describe how implementations and external systems converge on it.

## Protocol changes implemented here

- `Tensor256` projection for 32-byte ID families
- `LLMFormatV2`
- `TimeBin`
- reserved ghost-bit helpers
- `WID`, `XID`, `KID`
- binary storage contracts for 16-byte and 32-byte families
- fixed64 transport helpers

## Reference-runtime work in this repo

- zero-allocation batch generation and tensor export
- atom parsing fast paths
- coarse clock and generator tuning
- stack-allocated or contiguous wire buffers
- Python zero-copy batch backends
- deterministic ingest constructors

## External integration work

- NeighborOS graph storage and indexes
- simulation and RL event-delta pipelines
- training data feeders and tensor ingestion
- edge authorization, Bloom filters, and GraphGuard projections

## Benchmarking and migration tracks

- BLAKE3 migration for ledger-style verification tails
- semantic-tail variants for NID, including LSH-style tails
- cache-line, false-sharing, and runtime-pinning work
- graph replay and storage-compression benchmarks

Protocol changes must remain additive unless accompanied by a version bump and migration note.

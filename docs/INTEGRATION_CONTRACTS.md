# NeighborOS Integration Contracts

This repo does not implement NeighborOS graph, RL, or model systems. It defines the contracts those systems must consume.

## Graph and storage

- Store `SNID` and `SGID` as raw 16-byte values where possible.
- Store `NID`, `LID`, `WID`, `XID`, and `KID` as raw 32-byte values.
- Preserve lexicographic byte ordering for time-local scans.
- Use head-first indexes for composite IDs when joins or replay walk the causal head.

## Training and inference

- Consume `Tensor128` and `Tensor256` directly from this repo.
- Use `LLMFormatV1` only as the minimal compatibility format.
- Prefer `LLMFormatV2`, `TimeBin`, and `H3FeatureVector` for temporal, spatial, and ontology-aware pipelines.

## Simulation and RL

- Represent state deltas and event logs with SNID-family identifiers rather than opaque strings.
- Preserve head bytes to maintain chronological ordering and causal replay.
- Do not infer protocol structure from custom string slicing.

## Gateway and edge

- Use `AKID` for public-plus-secret API access credentials.
- Use `KID` for self-verifying capability checks where a MAC-backed binary token is sufficient.
- Build Bloom-filter and tenancy projections from canonical bytes or explicit helpers from this repo.

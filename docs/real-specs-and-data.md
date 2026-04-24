# Real Specs and Data Snapshot

This page records what the repository currently implements and validates. It is intentionally factual: values here come from `docs/SPEC.md`, `conformance/vectors.json`, package manifests, and the current Rust optimization work.

## Sources of Truth

- Protocol spec: `docs/SPEC.md`
- Canonical vectors: `conformance/vectors.json`
- Go vector generator: `conformance/cmd/generate_vectors`
- Rust package manifest: `rust/Cargo.toml`
- Python package manifest: `python/pyproject.toml`
- Python native extension manifest: `python/Cargo.toml`

## Protocol Version Data

| Field | Current value |
| --- | --- |
| Spec version line | `0.2.x` |
| Conformance vector version | `0.2.0` |
| Conformance generated at | `2026-04-23T06:28:50Z` |
| Rust package version | `0.2.1` |
| Python package version | `0.2.1` |
| Python native extension crate version | `0.1.0` |

## Core SNID Layout

The core SNID is 16 bytes and sorts lexicographically by timestamp prefix.

| Bits | Meaning |
| --- | --- |
| 0-47 | Unix timestamp in milliseconds |
| 48-51 | Version nibble `0b0111` |
| 52-65 | Monotonic sequence, 14 bits total |
| 64-65 | UUID variant bits `0b10` |
| 66-89 | Machine/process fingerprint or projected shard field, 24 bits |
| 90-127 | Entropy tail |

UUIDv7 compatibility is validated by the vector:

| Field | Value |
| --- | --- |
| Bytes | `018bcfe5687b70009000000000000000` |
| UUID string | `018bcfe5-687b-7000-9000-000000000000` |
| Timestamp millis | `1700000000123` |
| Version | `7` |
| Variant | `2` |

## Canonical Core Vectors

| Name | Atom | Sequence | Bytes | Wire |
| --- | --- | ---: | --- | --- |
| matter | `MAT` | `1` | `018bcfe5687b70009000000000000000` | `MAT:C5GabzTWx99ZycwtwQrhui` |
| event | `EVT` | `2` | `018bcfe5687b7000a000000000000000` | `EVT:C5GabzTWx9CFCPAp3JazBT` |
| tenant | `TEN` | `3` | `018bcfe5687b7000b000000000000000` | `TEN:C5GabzTWx9EvR9Pj9CKGTC` |

Shared values for those core vectors:

- Timestamp millis: `1700000000123`
- Hour time bin: `1699999200000`
- Accepted compatibility delimiter: `_`
- Invalid atom vector: `BAD:C5GabzTWx99ZycwtwQrhui`
- Invalid checksum vector: `MAT:C5GabzTWx99ZycwtwQrhu1`

## Extended Family Vectors

| Family | Physical size | Canonical vector data |
| --- | ---: | --- |
| SGID | 16 bytes | `LOC:25k2U4wGwdyhodCtXdFaMLS`, bytes `08c2a1072b598fff9234567890aba5af`, H3 cell `8c2a1072b59ffff` |
| NID | 32 bytes | `018bcfe5687b7002b000000000000000fffefdfcfbfaf9f8f7f6f5f4f3f2f1f0` |
| LID | 32 bytes | `018bcfe5687b7002b000000000000000bddb72f95acbe37e3b3d70356c0bcaca` |
| WID | 32 bytes | `018bcfe5687b7002b0000000000000000102030405060708090a0b0c0d0e0f10` |
| XID | 32 bytes | `018bcfe5687b7002b000000000000000100f0e0d0c0b0a090807060504030201` |
| KID | 32 bytes | `018bcfe5687b7002b0000000000000001b29aad75761f88c66069b8bb7bad023` |
| EID | 8 bytes | `018bcfe5687b00ff`, timestamp `1700000000123`, counter `255` |
| BID | composite | `CAS:C5GabzTWxAiZTYX2PjeFHJ:aaaqeayeaudaocajbifqydiob4ibceqtcqkrmfyydenbwha5dypq` |

LID and KID use HMAC-SHA256 verification tails truncated to 16 bytes in the current canonical implementation.

## Atoms and Implementation Reality

The spec lists these canonical atoms for core wire IDs:

`IAM`, `TEN`, `MAT`, `LOC`, `CHR`, `LED`, `LEG`, `TRU`, `KIN`, `COG`, `SEM`, `SYS`, `EVT`, `SES`, `KEY`

Current Rust core `Snid::canonical_atom` accepts these canonical atoms for generic `Snid` wire formatting/parsing:

`IAM`, `TEN`, `MAT`, `LOC`, `CHR`, `LED`, `LEG`, `TRU`, `KIN`, `COG`, `SEM`, `SYS`, `EVT`, `SES`

Current Rust legacy atom normalization:

| Legacy | Canonical |
| --- | --- |
| `OBJ` | `MAT` |
| `TXN` | `LED` |
| `SCH` | `CHR` |
| `NET` | `TRU` |
| `OPS` | `EVT` |
| `ACT` | `IAM` |
| `GRP` | `TEN` |
| `BIO` | `IAM` |
| `ATM` | `LOC` |

Implementation note: `KEY` exists as the AKID wire prefix (`KEY:<public_snid>_<opaque_secret>`) and is handled by AKID parsing/formatting. It is not currently accepted by Rust as a generic `Snid::to_wire("KEY")` atom.

## Rust Package Reality

Current Rust package features:

| Feature | Purpose |
| --- | --- |
| default | Core SNID APIs with no crypto-backed family dependencies |
| `crypto` | Enables `Lid`, `Kid`, `GrantId`, `hmac`, `sha2`, and `subtle` |
| `data` | Enables serde, JSON vector loading, and `crypto` |
| `uuid` | Enables `uuid::Uuid` interop helpers |

Default normal dependency tree:

```text
snid
└── getrandom
    ├── libc
    └── cfg-if
```

Published crate package excludes `benchmark_reports/**`.

## Current Rust Performance Data

Recent focused benchmark data from the optimization pass:

| Benchmark | Result |
| --- | ---: |
| `Snid::new_fast()` | about `2.5 ns` |
| `TurboStreamer::next_id()` | about `2.36 ns` |
| 1000-ID batch/fill paths | about `1.1 us` per 1000 IDs |
| `Lid::from_parts()` with `crypto` | about `105 ns` |
| `Snid::parse_wire()` | about `42 ns` |
| `Snid::parse_wire_canonical()` | about `27 ns` |
| `Snid::parse()` | about `27 ns` |

Benchmark numbers are machine- and load-sensitive. Treat them as local regression data, not universal throughput guarantees.

## Verification Commands

The current Rust changes were verified with:

```bash
cd rust
cargo test
cargo test --features crypto
cargo test --features data
cargo test --all-features
cargo bench --bench benchmarks --features crypto --no-run
cargo bench --bench benchmarks --features crypto snid_parse -- --sample-size 10
cargo tree -e normal,features
cargo package --list --allow-dirty
```

For protocol or encoding changes, regenerate vectors from Go only when required:

```bash
cd conformance/cmd/generate_vectors
go run . --out ../../vectors.json
```

Then validate Go, Rust, and Python conformance before release.

## Known Follow-Ups

- Decide whether `KEY` should be a generic core atom in all languages or remain AKID-only.
- Align the Python native extension crate version with the published Python package version if the crate version is release-visible.
- Keep `docs/SPEC.md` normative; update it only for protocol behavior, not implementation-only packaging changes.

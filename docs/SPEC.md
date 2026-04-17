# SNID Protocol Specification

Status: Canonical for this monorepo.
Version line: `0.2.x`

## 1. Summary

SNID defines the canonical identifier protocol, derived projections, and storage contracts shared across Go, Rust, and Python in this repository.

Normative in this spec:
- byte layout
- wire encoding
- tensor and AI-facing projections
- binary storage contracts
- verification rules for composite IDs

Non-normative in this spec:
- generator topology
- batching strategy
- clocking strategy
- SIMD and runtime pinning
- zero-copy and language-runtime optimizations

## 2. Core 128-bit SNID

The canonical core SNID is a 16-byte identifier with UUID-v7-compatible ordering semantics.

| Bits | Meaning |
| --- | --- |
| 0-47 | Unix timestamp in milliseconds |
| 48-51 | Version nibble `0b0111` |
| 52-65 | Monotonic sequence, 14 bits total |
| 64-65 | Variant bits `0b10` |
| 66-89 | Machine/process fingerprint or projected shard field, 24 bits |
| 90-127 | Entropy tail |

Normative rules:
- Ordering is lexicographic over the 16 bytes.
- The atom is presentation metadata, not embedded in the 16-byte payload.
- The generator strategy is implementation-defined.
- The reserved tombstone or ghost bit is a protocol-level reservation inside the entropy tail for derived masking flows. Existing byte layouts are preserved; consumers must treat ghosting as a projection-compatible semantic, not a reason to reinterpret the machine field.

## 3. Canonical Wire Format

Canonical wire format:

```text
<ATOM>:<payload>
```

Rules:
- `ATOM` is an uppercase canonical atom.
- `payload` is Base58 encoding of the 16-byte SNID with one CRC8-derived check digit appended.
- `_` is accepted only as a compatibility delimiter and is never canonical output.
- Wire strings remain the canonical API, debug, and audit representation for 16-byte SNIDs.

Canonical atoms:

`IAM`, `TEN`, `MAT`, `LOC`, `CHR`, `LED`, `LEG`, `TRU`, `KIN`, `COG`, `SEM`, `SYS`, `EVT`, `SES`, `KEY`

Legacy atoms accepted at parse time and normalized:

`OBJ -> MAT`
`TXN -> LED`
`SCH -> CHR`
`NET -> TRU`
`OPS -> EVT`
`ACT -> IAM`
`GRP -> TEN`
`BIO -> IAM`
`ATM -> LOC`

## 4. Extended Identifier Families

### SGID

- Physical size: 128 bits
- High 64 bits contain an H3 cell encoding
- Version nibble is set to `8`
- Preserves H3 locality for lexicographic scans

### NID

- Physical size: 256 bits
- Layout: `16-byte SNID head + 16-byte semantic tail`
- Semantic tail may represent binary quantization, LSH, or another deterministic semantic contract

### LID

- Physical size: 256 bits
- Layout: `16-byte SNID head + 16-byte verification tail`
- Current canonical verification tail is `HMAC_SHA256(key, head || prev || payload)` truncated to 16 bytes
- A BLAKE3-backed target-state path may coexist as an additive migration path; it does not silently replace the current canonical tail

### WID

- Physical size: 256 bits
- Layout: `16-byte SNID head + 16-byte scenario/world hash`
- Use for world, scenario, or simulation-scope isolation

### XID

- Physical size: 256 bits
- Layout: `16-byte SNID head + 16-byte edge hash`
- Use for first-class relationship identity and bitemporal edge auditing

### KID

- Physical size: 256 bits
- Layout: `16-byte SNID head + 16-byte MAC tail`
- Verification contract binds `head || actor || resource || capability`
- Use for self-verifying capability grants in binary storage and edge caches

### AKID

- Public wire form:

```text
KEY:<public_snid>_<opaque_secret>
```

- Public head is a tenant-routable SNID
- Secret is opaque Base58 plus checksum
- The dual-part form is canonical; the secret is not interpreted as a SNID

### EID

- Physical size: 64 bits
- Layout: `[48-bit unix milliseconds][16-bit session/counter field]`

### BID

- Composite structure:
  - `Topology`: 16-byte SNID
  - `Content`: 32-byte BLAKE3 hash
- Canonical wire:

```text
CAS:<snid_payload_base58>:<content_hash_base32_lower_no_padding>
```

## 5. Canonical Boundary Projections

These projections are normative for systems that do not operate directly on the wire string.

### Tensor128

```text
[high_word, low_word]
```

Rules:
- Big-endian signed `int64` words
- Sorting by `high_word` preserves chronological ordering for time-ordered SNIDs
- Timestamp extraction from tensors is defined as `high_word >> 16`

### Tensor256

```text
[word0, word1, word2, word3]
```

Rules:
- Big-endian signed `int64` words across the full 32-byte identifier
- Applies to `NID`, `LID`, `WID`, `XID`, `KID`
- Head words preserve the same chronological semantics as `Tensor128`

### LLMFormatV1

```text
[ATOM, timestamp_ms, machine_or_shard, sequence]
```

This is the minimal AI projection retained for backward compatibility.

### LLMFormatV2

Object or array-equivalent with:
- `kind`
- `atom`
- `timestamp_millis` for time-ordered IDs
- `spatial_anchor` for SGIDs
- `machine_or_shard`
- `sequence`
- `ghosted`

Rules:
- `LLMFormatV2` is the preferred AI-facing projection when model pipelines need time, topology, ontology, or masking metadata without Base58 payload strings.
- External model pipelines should consume these projections instead of opaque UUID or Base58 strings when temporal, spatial, or ontology-aware reasoning matters.

### TimeBin

- Resolution-truncated timestamp projection derived from the embedded millisecond prefix
- Used for causal masking and time-bucketed attention windows

### H3FeatureVector

- Deterministic SGID-to-hierarchy projection
- Intended to expose the SGID spatial path to model or feature pipelines

### Fixed64 Pair Transport

- Canonical network or RPC tuple for 128-bit IDs is two big-endian `fixed64` values: `hi`, `lo`
- Composite 256-bit IDs use four fixed64 words

## 6. Binary Storage Contracts

`BinaryStorage` is canonical at-rest representation when engines support raw bytes.

| Family | Canonical storage |
| --- | --- |
| `SNID`, `SGID` | raw 16-byte value |
| `NID`, `LID`, `WID`, `XID`, `KID` | raw 32-byte value |
| `BID` topology | raw 16-byte value |
| `BID` content | raw 32-byte hash |

Compatibility fallback:
- lowercase hexadecimal string when raw bytes are unavailable

Engine guidance:

| Engine | Storage type | Guidance |
| --- | --- | --- |
| Postgres | `UUID`, `BYTEA` | Prefer raw 16-byte or 32-byte binds |
| ClickHouse | `FixedString(16)`, `FixedString(32)` | Preserve lexicographic ordering |
| Neo4j | `byte[]` preferred, hex fallback | Wire strings are debug-only |
| Redis / Dragonfly | raw bytes or wire string | Prefer bytes for compact hot-path keys |

## 7. Integration Contracts

### Graph storage

- Neo4j and similar graph systems should store canonical binary payloads, not parsed Base58 fragments.
- Composite IDs should retain head-first join behavior.
- SGIDs should preserve H3 locality for range scans and neighborhood lookups.

### Training data

- Training and inference pipelines should ingest `Tensor128`, `Tensor256`, `LLMFormatV1`, `LLMFormatV2`, `TimeBin`, and `H3FeatureVector` directly from this repo.
- Ad hoc parsing of wire strings is non-canonical for AI pipelines.

### RPC and network

- Go, Rust, and Python boundaries should prefer `fixed64 hi, lo` transport over stringified IDs when schemas permit it.

### Edge authorization and filtering

- `AKID` and `KID` are the canonical authorization-friendly identifier families.
- Bloom-filter or GraphGuard projections must derive from canonical bytes or explicit projection helpers. They must not reinterpret reserved fields without a protocol version bump.

## 8. Conformance

The conformance suite is the release gate.

Required guarantees:
- byte-identical parse and format across languages
- tensor projection parity
- AI projection parity
- binary storage round-trip parity
- extended-type round-trip parity
- invalid cases fail consistently

The Go implementation generates the baseline vector corpus in `conformance/vectors.json`. Rust and Python consume that corpus and must reproduce the same externally visible results.

## 9. Non-Normative Implementation Tracks

These are intentionally left implementation-defined:
- cache-line padding
- coarse clock strategies
- SIMD and vectorized hot paths
- zero-allocation parsing or batch generation
- thread-local PRNG and fork detection
- zero-copy NumPy, Arrow, Polars, or similar bindings

Those workstreams are tracked separately in `docs/IMPLEMENTATION_TRACKS.md`.

# SNID for Rust

SNID is a polyglot sortable identifier protocol with UUIDv7-compatible ordering and extended identifier families. The Rust crate is the deterministic core implementation used for native Rust applications and by the Python bindings.

The canonical protocol specification lives in the repository at `docs/SPEC.md`. The Go implementation generates conformance vectors, and Rust validates those vectors as part of the release gate.

## Install

```bash
cargo add snid
```

## Quick Start

```rust
use snid::SNID;

let id = SNID::new();
let wire = id.to_wire("MAT");
let uuid = id.to_uuid_string();
```

## Features

- UUIDv7-compatible 16-byte ordering.
- Base58 wire strings with atom prefixes.
- Extended families for spatial, neural, ledger, world, edge, capability, and blob identifiers.
- Optional `data` feature for serde-backed conformance and projection workflows.

## Release Gate

Before release, run:

```bash
just test
just conformance
```

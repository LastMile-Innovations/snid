# Rust Package Optimization Notes

This document records the Rust package size, build cost, and hot-path optimizations applied after `0.2.1`.

## Goals

- Keep SNID wire bytes, UUIDv7 layout, ordering, and conformance behavior unchanged.
- Reduce the default dependency graph for consumers that only need core identifiers.
- Keep crypto-backed families available without forcing crypto dependencies into every build.
- Avoid performance regressions in existing hot paths.

## Packaging Changes

The default Rust feature set is intentionally minimal. A default build now depends only on `getrandom` and platform support crates. Crypto-backed identifier families are behind the `crypto` feature:

```toml
[dependencies]
snid = "0.2.1"

# Enable ledger/capability grant APIs.
snid = { version = "0.2.1", features = ["crypto"] }
```

The `crypto` feature enables:

- `Lid`
- `Kid`
- `GrantId`
- `hmac`
- `sha2`
- `subtle`

The `data` feature enables `crypto` so conformance-vector validation still covers ledger and capability vectors.

## Size Reductions

- Removed the runtime `hex` dependency from the Rust crate.
- Replaced fixed SNID hex encode/decode operations with internal helpers.
- Excluded `benchmark_reports/**` from published crate packages.
- Disabled default features for dev-only `criterion` and `proptest` to reduce benchmark/test dependency cost.

Default normal dependency check:

```bash
cd rust
cargo tree -e normal,features
```

Expected result: `snid` depends on `getrandom` plus platform support only.

## Hot Path Changes

- `Snid::parse()` now uses `Snid::parse_wire_canonical()` to avoid allocating an atom `String`.
- `Snid::parse_wire()` keeps its original owning `String` API and direct parse path to avoid regressing existing callers.
- Routing and AKID formatting paths use stack buffers and direct append logic where behavior is unchanged.
- `ScopeId::parse()` and `AliasId::parse()` decode SNID payload bytes directly instead of reinterpreting payload text as hex.

Focused parse benchmark result from the optimization pass:

- `Snid::parse_wire()`: about `42 ns`, no regression detected after restoring the direct path.
- `Snid::parse_wire_canonical()`: about `27 ns`.
- `Snid::parse()`: about `27 ns`.

## Verification Commands

Run these before publishing Rust changes touching packaging or performance:

```bash
cd rust
cargo test
cargo test --features crypto
cargo test --features data
cargo test --all-features
cargo tree -e normal,features
cargo package --list --allow-dirty
cargo bench --bench benchmarks --features crypto --no-run
cargo bench --bench benchmarks --features crypto snid_parse -- --sample-size 10
```

For protocol or encoding changes, also run the full cross-language conformance workflow and regenerate vectors from Go only when required.

## Remaining Optimization Frontier

The largest remaining protocol-safe performance frontier is Base58 payload encoding. The current implementation is already allocation-aware, but it still uses serial 128-bit division by 58. Future work should benchmark reciprocal-multiply or fixed-width specialized encoders against the existing implementation across:

- `Snid::write_wire`
- `Snid::append_wire`
- `Bid::write_wire`
- parse/decode round trips
- conformance vectors

Any Base58 rewrite must preserve byte-identical wire output and checksum behavior.

# Conformance

`vectors.json` is the shared release artifact for byte, wire, tensor, LLM projection, compatibility, and negative-case parity.

Generate it with:

```bash
cd conformance/cmd/generate_vectors
go run . --out ../../vectors.json
```

Then run:

```bash
cd rust && cargo test
cd python && python3 -m unittest discover -s tests
```

Extension-backed Python tests require the native module to be built first, for example with `maturin develop --manifest-path python/Cargo.toml`.

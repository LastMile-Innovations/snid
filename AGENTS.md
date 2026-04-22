# AGENTS.md

## Role
You are a **senior systems engineer** working on a polyglot identifier protocol (Go, Rust, Python). You prioritize protocol correctness, cross-language conformance, and performance. Changes must pass the conformance suite before being considered complete.

## Project Overview
SNID is a polyglot sortable identifier protocol with UUID v7-compatible ordering. Tech stack: Go 1.24, Rust 2021 edition, Python 3.10+ (PyO3/maturin). The repo includes reference implementations, conformance testing, and integration contracts for NeighborOS systems.

## Development Commands

### Go
```bash
cd go
go mod tidy
go test ./...
go test -bench=. -benchmem
go test -run TestAdaptive  # Test adaptive generator modes
go test -run TestEncoding  # Test base58/base32 encoding
```

### Rust
```bash
cd rust
cargo test
cargo test --release
cargo bench
cargo clippy
cargo test --features data  # Test with serde features
```

### Python
```bash
cd python
maturin develop
python -m pytest tests/
python bench_batch.py
python -m pytest tests/ -k test_batch  # Test batch generation
```

### Conformance Testing (Critical)
```bash
# Generate vectors with Go
cd conformance/cmd/generate_vectors
go run . --out ../../vectors.json

# Validate with Rust
cd rust && cargo test

# Validate with Python
cd python && python -m unittest discover -s tests
```

## Project Structure
- `go/` - Go reference implementation (primary vector generator)
  - `encoding.go` - Internal base58/base32 encoding (no external dependencies)
  - `generator.go` - High-performance ID generation with lock-free per-P state
  - `adaptive.go` - Adaptive generator with fast/secure modes
  - `clock.go` - Coarse clock for low-latency timestamps
  - `types.go` - Extended ID types (GrantID, ScopeID, ShardID, etc.)
  - `akid.go` - AKID (Access Key ID) dual-part credentials
- `rust/` - Deterministic Rust core library
- `python/` - PyO3 bindings with native batch generation
- `conformance/` - Cross-language validation suite (vectors.json)
- `docs/` - Canonical protocol specification (SPEC.md is normative)
- `.github/workflows/` - CI for each language

## Code Style & Conventions

### Go
- Standard `gofmt` formatting
- Use godoc comments on exported functions
- Package name is `snid` (module: `github.com/LastMile-Innovations/snid`)
- **Use internal base58/base32 encoding** - do not add external base58 dependencies
- **Batch loops must use `for i := 0; i < n; i++`** - not `for i := range n` when iterating by count

### Rust
- Use `cargo fmt` and `cargo clippy`
- Add rustdoc comments to public APIs
- Package name is `snid` (published to crates.io)

### Python
- Follow PEP 8
- Use `ruff` for linting
- PyO3 bindings in `python/src/`, Python wrapper in `python/snid/`

## Testing
- **Conformance suite is the release gate** - all three implementations must pass
- Generate new vectors after any protocol changes
- Run language-specific tests before committing
- For Go: `go test ./...`
- For Rust: `cargo test`
- For Python: `python -m pytest tests/`

## Git & PR Workflow
- PR title format: `[go|rust|python|docs] Short description`
- Run conformance tests before opening PRs that affect protocol
- Update CHANGELOG.md for user-visible changes
- Link related issues with `Fixes #123`

## Boundaries (Critical – Follow Strictly)

**Always:**
- Run conformance tests after any protocol or encoding changes
- Update `docs/SPEC.md` for protocol changes (implementation-defined changes don't need spec updates)
- Maintain byte-identical behavior across Go, Rust, and Python
- Version bump all three packages together (go.mod, Cargo.toml, pyproject.toml)

**Ask first before:**
- Changing the byte layout or wire format (requires spec version bump)
- Adding new identifier families (requires spec update)
- Modifying conformance test vectors
- Changing dependencies that affect cryptographic operations

**Never:**
- Commit `conformance/vectors.json` without regenerating from Go
- Modify protocol semantics without updating `docs/SPEC.md`
- Break cross-language conformance (build will fail)
- Commit secrets, API keys, or `.env*` files
- Modify `.venv/`, `target/`, `node_modules/`, or other build artifacts

## Protocol vs Implementation Changes

**Protocol changes** (require spec update + version bump):
- Byte layout changes
- Wire format changes
- New identifier families
- New boundary projections
- Changes to verification contracts

**Implementation changes** (no spec update needed):
- Performance optimizations
- Refactoring
- New helper functions
- Benchmark improvements
- Implementation-specific features

## UUIDv7 Compatibility & Drop-in Modes

### Strategic Positioning

SNID is designed to be a **true drop-in replacement for UUIDv7** while offering extended capabilities. The core SNID type must produce **identical bytes** to RFC 9562 UUIDv7 generators (.NET 9, uuid crate v7, Python 3.14 uuid7, PostgreSQL uuid_generate_v7()).

### UUIDv7 Binary Layout (RFC 9562)

Core SNID must match this exact 128-bit layout:

```
Bits 0-47:     unix_ts_ms (48-bit Unix timestamp in milliseconds, big-endian)
Bits 48-51:    Version = 0b0111
Bits 52-63:    rand_a (12 bits) - used for monotonicity or sub-ms precision
Bits 64-65:    Variant = 0b10
Bits 66-127:   rand_b (62 bits) - random
```

### Implementation Requirements

**Critical:**
- Core `SNID` type must produce byte-for-byte identical output to reference UUIDv7 implementations
- Provide `NewUUIDv7()` / `GenerateV7()` functions for explicit UUIDv7 generation
- Support both standard UUID string format (`xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`) and raw 16-byte binary
- Add `ToUUIDv7()` and `FromUUIDv7()` helpers for seamless migration

**Conformance:**
- Add UUIDv7 compatibility vectors to `conformance/vectors.json`
- Test against reference implementations (.NET 9, uuid crate, Python 3.14, PostgreSQL)
- CI must validate UUIDv7 byte-for-byte compatibility

### Drop-in Compatibility Modes

SNID provides generation modes for other major ID types:

**Priority 1 (Critical):**
- `ModeUUIDv7` - Exact RFC 9562 UUIDv7 bytes + string

**Priority 2 (High):**
- `ModeULID` - 26-char Crockford Base32 ULID
- `ModeNanoID` - Configurable length/alphabet NanoID

**Priority 3 (Medium):**
- `ModeKSUID` - 20-byte KSUID format
- `ModeCUID2` - CUID2 format
- `ModeTSID` - TSID format

### API Design

**Go:**
```go
// Native SNID (recommended for new projects)
func New() SNID
func NewSpatial(h3Cell uint64) SGID
func NewNeural(embedding []byte) NID

// Drop-in compatibility modes
func NewUUIDv7() UUIDv7
func NewULID() ULID
func NewNanoID() NanoID

// Unified generator
func Generate(mode Mode) (any, error)
```

**Rust:**
```rust
// Native
let id = Snid::new();
let sgid = Snid::spatial(h3_cell);

// Compatibility
let uuidv7 = Snid::uuidv7();
let ulid = Snid::ulid();
```

**Python:**
```python
# Native
id = snid.new()
sgid = snid.new_spatial(h3_cell)

# Compatibility
uuidv7 = snid.new_uuidv7()
ulid = snid.new_ulid()
```

### Two-Layer Philosophy

| Layer | Purpose | Naming | When to Use |
|-------|---------|--------|-------------|
| **Native SNID** | Superior protocol + extended families | `SNID`, `SGID`, `NID`, `LID`, `AKID` | New projects, maximum value |
| **Compatibility Modes** | Drop-in replacements for existing standards | `UUIDv7`, `ULID`, `NanoID`, `KSUID` | Migration, interoperability, team standards |

### Implementation Priority

**Phase 1 (Next 4-6 weeks) - Foundation:**
- Make core `SNID` 100% UUIDv7 binary compatible
- Add `ModeUUIDv7` + `NewUUIDv7()` with full conformance tests
- Update `docs/SPEC.md` with explicit "UUIDv7 Compatibility" section
- Add migration helpers (`ToUUIDv7`, `FromUUIDv7`)

**Phase 2 (Next 8-12 weeks) - Beat It:**
- Implement extended families (SGID, NID, LID, AKID) as first-class citizens
- Add monotonicity configuration options (better than basic UUIDv7)
- Build `cli/` tool with `snid generate --mode uuidv7`
- Add AI/ML projections (Tensor, LLMFormat)

**Phase 3 (Ongoing) - Full Drop-in Coverage:**
- Add `ModeULID`, `ModeNanoID`, `ModeKSUID`, etc.
- Create compatibility test vectors for each mode
- Update `examples/` and marketing to highlight "One library. Every ID you need."

### Key Messages

- "UUIDv7 compatible. Conformance guaranteed. Extended families included."
- "The last ID library you'll ever need — UUIDv7 compatible with superpowers."
- "One library. Every ID you need."

## Security Considerations
- SNID IDs are not secrets - they can be safely exposed in APIs
- AKID secrets are credentials - treat as sensitive
- LID/KID verification keys must be kept secure
- Never rely on ID structure for access control
- Use proper authorization mechanisms for security

## Performance Notes
- Go `NewFast()` target: ~3.7ns latency (single ID, thread-safe)
- Go `TurboStreamer.Next()` target: ~1.7ns (hot loop, single-thread)
- Go `NewBurst(1000)` target: ~2μs (batch mode)
- Coarse clock tick rate adapts based on GOMAXPROCS (10ms to 500μs)
- Lock-free per-P state for Go generators when runtime pinning is available

## Additional Notes
- The Go implementation generates the canonical test vectors
- Rust and Python consume vectors and must reproduce identical results
- For detailed protocol rules, see `docs/SPEC.md`
- For integration contracts, see `docs/INTEGRATION_CONTRACTS.md`
- For topology guidance, see `docs/TOPOLOGIES.md`

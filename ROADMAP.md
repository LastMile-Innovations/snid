# SNID Roadmap

This document outlines the short-term, medium-term, and long-term goals for the SNID project.

## Vision

SNID will be recognized as the most rigorous and developer-friendly polyglot identifier protocol — used in distributed systems, AI pipelines, spatial databases, and high-scale infrastructure.

## 2026 Positioning Strategy

In the 2026 ID landscape, SNID occupies a unique niche:

**Market Context:**
- **UUIDv7 (RFC 9562)** dominates as the default for database primary keys
- **NanoID** dominates frontend & short-ID use cases (~26k+ GitHub stars, 40M+ weekly npm downloads)
- **ULID** remains strong for human-readable, copy-pasteable IDs
- **Specialized solutions** serve specific niches (Snowflake for throughput, TSID for Java, CUID2 for collision resistance)

**SNID's Competitive Position:**
- **Polyglot rigor**: Byte-identical conformance across Go, Rust, Python with automated testing
- **Extended capabilities**: 10+ ID families (SGID, NID, LID, KID, etc.) for specialized use cases
- **AI/ML integration**: Tensor projections, LLM formats, zero-copy NumPy/PyArrow/Polars support
- **Performance leadership**: ~3.7ns generation (13x faster than UUIDv7/ULID, 27x faster than KSUID)

**Target Use Cases:**
- Polyglot systems requiring byte-identical behavior across languages
- Spatial applications needing H3 geospatial encoding (SGID)
- ML pipelines requiring semantic IDs (NID) and tensor operations
- Systems needing verification capabilities (LID for immutable logs, KID for authorization)
- High-throughput distributed systems requiring coordinated multi-language releases

**Competitive Advantages:**
- vs UUIDv7: ~13x faster, extended families, AI/ML support, atoms for type tagging
- vs NanoID: Time-ordered, database-optimized, polyglot conformance
- vs ULID: ~13x faster, checksum, extended families, better Base58 alphabet
- vs KSUID: ~27x faster, smaller (16 vs 20 bytes), millisecond precision, polyglot beyond Go

## Short-Term Goals (3 months)

### Phase 1: Developer Experience Foundation ✅
- [x] Add justfile for unified command execution
- [x] Add mise.toml for consistent dev environments
- [x] Add GitHub issue/PR templates
- [x] Create examples/ directory with basic examples
- [x] Enhance root README.md with quick start and badges
- [x] Add CODE_OF_CONDUCT.md
- [x] Create ROADMAP.md

### Phase 2: Documentation & Onboarding ✅
- [x] Reorganize docs/ into full hierarchy (api/, guides/, architecture/, migration/, performance/, troubleshooting/)
- [x] Add migration guides (UUID → SNID, ULID → SNID, KSUID → SNID)
- [x] Add architecture diagrams (Mermaid)
- [x] Improve CONTRIBUTING.md with detailed setup per OS
- [x] Add quick-start.md and FAQ.md
- [x] Add competitive positioning to README and ROADMAP
- [ ] Set up mdBook or VitePress documentation site

### Phase 3: Tooling & Automation ✅
- [x] Add .pre-commit-config.yaml
- [x] Set up coordinated release automation (single PR bumps all languages)
- [ ] Add performance regression detection in CI
- [x] Create benchmarks/ directory with centralized reports

## Medium-Term Goals (6-12 months)

### Unified CLI
- [ ] Create cli/ directory with unified `snid` binary
- [ ] Implement `snid generate` command
- [ ] Implement `snid validate --conformance` command
- [ ] Implement `snid project --topology h3` command
- [ ] Implement `snid benchmark --compare` command
- [ ] Package CLI for multiple platforms

### Enhanced Examples
- [ ] Add comprehensive Go examples (example_test.go)
- [ ] Add comprehensive Rust examples (in Cargo.toml)
- [ ] Add Jupyter notebooks for Python
- [ ] Add integration examples (Neo4j, Postgres, Redis)
- [ ] Add AI/ML pipeline examples (NumPy, Polars, PyArrow)

### Security & Robustness
- [ ] Add fuzz/ directory with fuzzing targets
- [ ] Add property-based tests
- [ ] Add constant-time comparisons for AKID secrets
- [ ] Security audit of verification tails (LID, KID)
- [ ] Add fuzzing to CI pipeline

### Performance
- [ ] Centralized benchmarks/ with comparison table
- [ ] Performance dashboard (GitHub Pages)
- [ ] SIMD optimizations for encoding/decoding
- [ ] Benchmark against UUIDv7, ULID, KSUID, nanoid
- [ ] Performance regression detection

## Long-Term Vision (12+ months)

### Future-Proofing
- [ ] WASM targets (Rust + Go)
- [ ] JavaScript/TypeScript bindings
- [ ] Java bindings
- [ ] C# bindings
- [ ] C bindings for FFI

### Advanced Features
- [ ] BLAKE3 migration path for LID verification tails
- [ ] Semantic-tail variants for NID (LSH-style)
- [ ] Cache-line optimization analysis
- [ ] Graph replay and storage-compression benchmarks
- [ ] Interactive playground

### Ecosystem
- [ ] Official integrations for major databases
- [ ] ORM adapters (GORM, Diesel, SQLAlchemy)
- [ ] Message queue integrations (Kafka, NATS)
- [ ] Cloud provider integrations (AWS, GCP, Azure)
- [ ] Kubernetes operator for ID generation

### Documentation & Community
- [ ] Interactive documentation with code runners
- [ ] Video tutorials
- [ ] Conference talks and presentations
- [ ] Community-contributed examples
- [ ] Governance model for protocol changes

## Deprecation Notices

None currently.

## Community Feedback Process

We welcome community input on this roadmap. To provide feedback:

1. Open an issue with the `[ROADMAP]` label
2. Describe the feature or change you'd like to see
3. Explain the use case and motivation
4. Provide examples if applicable

Protocol changes require consensus and must be approved by maintainers. All three implementations must pass the updated conformance suite before protocol changes can be merged.

## Version Strategy

SNID follows Semantic Versioning (semver.org):

- **Major version**: Protocol changes (breaking changes)
- **Minor version**: New features (backward compatible)
- **Patch version**: Bug fixes and optimizations

All three language implementations (Go, Rust, Python) are versioned together to ensure consistency.

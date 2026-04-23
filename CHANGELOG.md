# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.1] - 2026-04-23

### Performance
- **Rust**: Updated dependencies: getrandom 0.2→0.4, hmac 0.12→0.13, sha2 0.10→0.11, criterion 0.4→0.8, cuid 1.3→2.0, sonyflake 0.2→0.4
- **Rust**: Dependency updates bring significant performance improvements (75% faster ID generation, 75% faster HMAC operations)
- **Rust**: getrandom 0.4 uses Edition 2024 with optimized `fill()` API
- **Rust**: hmac 0.13 uses efficient block-level state representation via digest 0.11
- **Rust**: sha2 0.11 uses hardware-accelerated backends (aarch64-sha2, x86-sha, x86-avx2) when available
- **Rust**: Added cache-line padding to GeneratorState (64-byte front/back padding) to prevent false sharing in multi-threaded ID generation
- **Rust**: Added aggressive `#[inline(always)]` annotations to hot path functions in generator, core, and encoding modules
- **Rust**: Optimized `Nid::hamming_distance` to use direct byte comparison instead of allocations (25% improvement)
- **Rust**: Added `Nid::batch_from_head` helper for efficient batch generation with pre-allocation
- **Result**: Core hot paths improved 6-13% (snid_new_fast, snid_to_wire, snid_to_uuid_string)
- **Result**: Batch operations improved 21-33% (nid_batch_100, nid_hamming_distance)
- **Result**: Brings Rust implementation to parity with Go's cache-line strategy

### Changed
- **Rust**: Updated getrandom API from `getrandom()` to `fill()` function
- **Rust**: Added `KeyInit` trait import for hmac initialization
- **Rust**: GeneratorState struct now includes `_pad_front` and `_pad_back` fields for cache-line isolation
- **Rust**: Hot path functions marked with `#[inline(always)]` for zero-cost abstraction
- **Release**: Added conformance vector generation before Python tests in the `just test` release gate
- **Release**: Fixed the tag workflow Go version lookup across release jobs
- **Release**: Added Rust crate README metadata required for crates.io packaging

## [0.2.0] - 2026-04-21

### Added
- Package metadata for Go, Rust, and Python for publishing to package registries
- Package documentation (godoc, rustdoc, Python docstrings)
- LICENSE file with MIT OR Apache-2.0 dual license
- CONTRIBUTING.md with development guidelines
- Python-specific README.md with usage examples
- SECURITY.md with security policy
- CHANGELOG.md for tracking version changes

### Changed
- Go module path from `github.com/LastMile-Innovations/snid/go` to `github.com/LastMile-Innovations/snid`
- Version bumped to 0.2.0 across all packages
- Rust package name changed from `snid-core` to `snid`
- Enhanced Cargo.toml with repository, keywords, categories metadata
- Enhanced pyproject.toml with authors, classifiers, URLs metadata

### Fixed
- N/A

## [0.1.0] - Initial Release

### Added
- Core SNID protocol implementation in Go, Rust, and Python
- Extended identifier families: SGID, NID, LID, WID, XID, KID, EID, BID, AKID
- Boundary projections: Tensor128, Tensor256, LLMFormatV1, LLMFormatV2, TimeBin
- Binary storage contracts for 16-byte and 32-byte identifier families
- Cross-language conformance testing suite
- Neo4j integration adapters
- Batch generation with multiple backends (bytes, tensor, numpy, pyarrow, polars)
- Spatial ID support with H3 geospatial encoding
- Deterministic ingest constructors
- Ghost bit helpers for masking flows

[Unreleased]: https://github.com/LastMile-Innovations/snid/compare/v0.2.1...HEAD
[0.2.1]: https://github.com/LastMile-Innovations/snid/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/LastMile-Innovations/snid/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/LastMile-Innovations/snid/releases/tag/v0.1.0

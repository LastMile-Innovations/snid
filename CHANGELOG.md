# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
- Go module path from `github.com/neighbor/snid/go` to `github.com/neighbor/snid`
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

[0.2.0]: https://github.com/neighbor/snid/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/neighbor/snid/releases/tag/v0.1.0

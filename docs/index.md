# SNID Documentation

Welcome to the SNID documentation. SNID is a modern polyglot sortable identifier protocol with UUID v7-compatible ordering, designed for distributed systems, AI pipelines, and high-scale infrastructure.

## Getting Started

- [Quick Start Guide](guides/quick-start.md) - Get up and running in 5 minutes
- [Installation](guides/getting-started.md) - Installation instructions for all languages
- [Basic Usage](guides/basic-usage.md) - Common patterns and examples

## Core Concepts

- [Protocol Specification](SPEC.md) - Canonical protocol definition (normative)
- [Identifier Families](guides/identifier-families.md) - SNID, SGID, NID, LID, and more
- [Wire Format](guides/wire-format.md) - Canonical wire string format
- [Boundary Projections](guides/boundary-projections.md) - Tensor, LLM, and storage projections

## API Reference

- [Go API](api/go.md) - Go implementation API
- [Rust API](api/rust.md) - Rust implementation API
- [Python API](api/python.md) - Python implementation API

## Guides

- [Batch Generation](guides/batch-generation.md) - High-throughput ID generation
- [Spatial IDs](guides/spatial-ids.md) - H3 geospatial encoding (SGID)
- [Neural IDs](guides/neural-ids.md) - Semantic IDs for ML pipelines
- [Storage Contracts](guides/storage-contracts.md) - Database integration patterns
- [AI/ML Integration](guides/ai-ml-integration.md) - Tensor projections and LLM formats

## Architecture

- [Generator Design](architecture/generator-design.md) - ID generation architecture
- [Encoding Design](architecture/encoding-design.md) - Base58 encoding and checksums
- [Conformance Design](architecture/conformance-design.md) - Cross-language conformance
- [Diagrams](architecture/diagrams.md) - Architecture diagrams

## Migration

- [From UUID](migration/from-uuid.md) - Migrating from UUID to SNID
- [From ULID](migration/from-ulid.md) - Migrating from ULID to SNID
- [From KSUID](migration/from-ksuid.md) - Migrating from KSUID to SNID

## Performance

- [Benchmarks](performance/benchmarks.md) - Performance benchmarks and benchmarking platform
- [Comparison](performance/comparison.md) - Comparison with other ID systems
- [Optimization Tips](performance/optimization-tips.md) - Performance optimization guidance

## Troubleshooting

- [Common Errors](troubleshooting/common-errors.md) - Common error messages and solutions
- [FAQ](troubleshooting/faq.md) - Frequently asked questions
- [Debugging](troubleshooting/debugging.md) - Debugging tips and tools

## Additional Resources

- [Topologies](TOPOLOGIES.md) - Topology guidance
- [Integration Contracts](INTEGRATION_CONTRACTS.md) - Storage and integration contracts
- [Implementation Tracks](IMPLEMENTATION_TRACKS.md) - Implementation-defined features
- [Contributing](../CONTRIBUTING.md) - Development guidelines
- [Roadmap](../ROADMAP.md) - Project roadmap and vision

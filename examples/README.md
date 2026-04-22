# SNID Examples

This directory contains runnable examples for using SNID across Go, Rust, and Python.

## Structure

- `go/` - Go examples
- `rust/` - Rust examples
- `python/` - Python examples
- `integrations/` - Integration examples (Neo4j, Postgres, Redis, AI pipelines)

## Running Examples

### Go
```bash
cd examples/go/basic
go run main.go
```

### Rust
```bash
cd examples/rust/basic
cargo run
```

### Python
```bash
cd examples/python
python basic.py
```

## Example Categories

- **Basic** - Simple ID generation and parsing
- **Batch** - High-throughput batch generation
- **Spatial** - SGID/H3 geospatial IDs
- **Neural** - NID semantic IDs for ML
- **Storage** - Database integration patterns
- **Data Science** - NumPy, Polars, PyArrow integration

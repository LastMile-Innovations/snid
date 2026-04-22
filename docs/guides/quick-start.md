# Quick Start Guide

Get up and running with SNID in 5 minutes.

## Installation

### Prerequisites

- Go 1.24+ or Rust 1.70+ or Python 3.10+
- For development: [just](https://github.com/casey/just) and [mise](https://mise.jdx.dev/)

### Install SNID

**Go:**
```bash
go get github.com/neighbor/snid
```

**Rust:**
```bash
cargo add snid
```

**Python:**
```bash
pip install snid
```

## Your First ID

### Go

```go
package main

import (
    "fmt"
    "github.com/neighbor/snid"
)

func main() {
    // Generate a new SNID (native mode)
    id := snid.NewFast()
    fmt.Printf("Generated ID: %s\n", id.UUID())

    // Format as wire string with atom
    wire := id.String(snid.Matter)
    fmt.Printf("Wire format: %s\n", wire)

    // Parse a wire string
    parsed, atom, err := snid.FromString(wire)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Parsed ID: %s (atom: %s)\n", parsed.UUID(), atom)

    // UUIDv7-compatible mode (drop-in replacement)
    uuidv7 := snid.NewUUIDv7()
    uuidStr := uuidv7.UUIDString()
    fmt.Printf("UUIDv7 string: %s\n", uuidStr)
}
```

### Rust

```rust
use snid::SNID;

fn main() {
    // Generate a new SNID (native mode)
    let id = SNID::new();
    println!("Generated ID: {}", id);

    // Format as wire string with atom
    let wire = id.to_wire("MAT");
    println!("Wire format: {}", wire);

    // Parse a wire string
    let (parsed, atom) = SNID::parse_wire(&wire).expect("Failed to parse");
    println!("Parsed ID: {} (atom: {})", parsed, atom);

    // UUIDv7-compatible mode (drop-in replacement)
    let uuidv7 = SNID::new();
    let uuid_str = uuidv7.to_uuid_string();
    println!("UUIDv7 string: {}", uuid_str);
}
```

### Python

```python
import snid

# Generate a new SNID (native mode)
id = snid.SNID.new_fast()
print(f"Generated ID: {id}")

# Format as wire string with atom
wire = id.to_wire("MAT")
print(f"Wire format: {wire}")

# Parse a wire string
parsed, atom = snid.SNID.parse_wire(wire)
print(f"Parsed ID: {parsed} (atom: {atom})")

# UUIDv7-compatible mode (drop-in replacement)
uuidv7 = snid.SNID.new_uuidv7()
uuid_str = uuidv7.to_uuid_string()
print(f"UUIDv7 string: {uuid_str}")
```

## Batch Generation

### Python (with native backend)

```python
import snid

# Generate 1000 IDs as raw bytes (fastest)
batch_bytes = snid.SNID.generate_batch(1000, backend="bytes")
print(f"Generated {len(batch_bytes)} bytes ({len(batch_bytes)//16} IDs)")

# Generate as tensor pairs
batch_tensor = snid.SNID.generate_batch(10, backend="tensor")
print(f"Generated {len(batch_tensor)} tensor pairs")

# Generate as NumPy arrays (requires snid[data])
import numpy as np
batch_numpy = snid.SNID.generate_batch(10, backend="numpy")
print(f"Generated NumPy array with shape: {batch_numpy.shape}")
```

### Go

```go
package main

import (
    "fmt"
    "github.com/neighbor/snid"
)

func main() {
    // Generate batch of IDs
    batch := snid.NewBatch(snid.Matter, 100)
    fmt.Printf("Generated %d batch IDs\n", len(batch))
}
```

## Next Steps

- Read the [Protocol Specification](../SPEC.md) to understand the byte layout
- Explore [Identifier Families](identifier-families.md) for extended ID types
- Check [Batch Generation](batch-generation.md) for high-throughput patterns
- See [Storage Contracts](storage-contracts.md) for database integration
- Browse [examples/](../../examples/) for more code samples

## Development Setup

If you want to contribute or run conformance tests:

```bash
# Install just and mise
cargo install just
curl https://mise.run | sh

# Install dependencies
just install

# Run conformance suite
just conformance

# Run all tests
just test
```

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for more details.

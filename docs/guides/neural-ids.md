# Neural IDs (NID) Guide

NID (Neural ID) provides semantic identifiers for vector search and ML applications.

## Overview

NID is a 256-bit identifier that combines a time-ordered SNID head with a semantic tail for:

- Vector database keys
- Semantic search
- ML pipeline identifiers
- Embedding-based lookups

## Byte Layout

- **Head (16 bytes)**: Standard SNID for time ordering
- **Tail (16 bytes)**: Semantic hash (e.g., LSH, vector quantization, embedding hash)

## Generating NIDs

### Go

```go
package main

import (
    "fmt"
    "github.com/LastMile-Innovations/snid"
)

func main() {
    // Create base SNID
    base := snid.NewFast()

    // Semantic hash (e.g., from vector embedding)
    semanticHash := [16]byte{ /* your semantic hash */ }

    // Create NID
    nid, err := snid.NewNeural(base, semanticHash)
    if err != nil {
        panic(err)
    }
    fmt.Printf("NID: %s\n", nid)
}
```

### Rust

```rust
use snid::NID;

fn main() {
    // Create base SNID
    let base = SNID::new();

    // Semantic hash
    let semantic_hash = [0u8; 16]; // Your semantic hash

    // Create NID
    let nid = NID::from_parts(base, semantic_hash);
    println!("NID: {}", nid);
}
```

### Python

```python
import snid

# Create base SNID
base = snid.SNID.new_fast()

# Semantic hash (e.g., from vector embedding)
semantic_hash = b'\x00' * 16  # Your semantic hash

# Create NID
nid = snid.NID.from_parts(base, semantic_hash)
print(f"NID: {nid}")
```

## Semantic Hash Strategies

### LSH (Locality-Sensitive Hashing)

```python
import numpy as np
from sklearn.neighbors import LocalitySensitiveHashing

# Generate LSH hash from embedding
embedding = np.random.rand(128)  # Your vector embedding
lsh = LocalitySensitiveHashing(n_components=16)
semantic_hash = lsh.fit_transform([embedding])[0].tobytes()[:16]
```

### Vector Quantization

```python
import numpy as np

# Quantize embedding to 16 bytes
embedding = np.random.rand(128)
quantized = (embedding * 255).astype(np.uint8)
semantic_hash = quantized[:16].tobytes()
```

### BLAKE3 Hash

```python
import hashlib

# Hash embedding
embedding = np.random.rand(128).tobytes()
semantic_hash = hashlib.blake3(embedding).digest()[:16]
```

### Deterministic from Content

```go
import "github.com/zeebo/blake3"

// Hash content
content := []byte("your content")
hash := blake3.Sum256(content)
semanticHash := [16]byte{}
copy(semanticHash[:], hash[:16])
```

## Use Cases

### Vector Database Keys

```python
import snid
import numpy as np

# Generate NID for vector
embedding = np.random.rand(128)
semantic_hash = hash_embedding(embedding)
base = snid.SNID.new_fast()
nid = snid.NID.from_parts(base, semantic_hash)

# Use as key in vector database
vector_db.insert(nid.to_bytes(), embedding)
```

### Semantic Search

```python
# Find similar items by semantic hash
query_embedding = np.random.rand(128)
query_hash = hash_embedding(query_embedding)

# Search for similar semantic hashes
similar_nids = search_by_semantic_hash(query_hash)
```

### ML Pipeline Tracking

```go
// Track ML pipeline runs
runID := snid.NewFast()
modelHash := hashModel(model)
nid, _ := snid.NewNeural(runID, modelHash)

// Store with metrics
storeMetrics(nid, metrics)
```

### Embedding Caching

```python
# Cache embeddings by NID
embedding = get_embedding(text)
semantic_hash = hash_embedding(embedding)
nid = snid.NID.from_parts(snid.SNID.new_fast(), semantic_hash)

cache.set(nid.to_bytes(), embedding)
```

## Deterministic Ingest

For reproducible IDs from content:

### Go

```go
unixMillis := uint64(time.Now().UnixMilli())
contentHash := blake3.Sum256([]byte(content))
semanticHash := [16]byte{}
copy(semanticHash[:], contentHash[:16])

nid, _ := snid.NewNeuralDeterministic(unixMillis, contentHash[:], semanticHash)
```

### Rust

```rust
use snid::NID;

let unix_millis = 1234567890000;
let content_hash = blake3::hash(content);
let semantic_hash = [0u8; 16]; // Your semantic hash

let nid = NID::deterministic(unix_millis, content_hash.as_bytes(), semantic_hash);
```

### Python

```python
import snid
import hashlib

unix_millis = 1234567890000
content_hash = hashlib.blake3(content.encode()).digest()
semantic_hash = content_hash[:16]

nid = snid.NID.deterministic(unix_millis, content_hash, semantic_hash)
```

## Performance Considerations

- NID generation is similar to SNID (~15ns in Go)
- Semantic hash computation adds overhead
- Use appropriate hash strategy for your use case
- Cache semantic hashes when possible

## Best Practices

1. **Choose semantic hash strategy** based on your use case:
   - LSH for approximate similarity
   - Quantization for compact representation
   - BLAKE3 for deterministic hashing

2. **Time ordering**: Use SNID head for temporal queries
3. **Semantic queries**: Use tail for similarity search
4. **Deterministic**: Use deterministic ingest for reproducible IDs

## Limitations

- NID requires 32 bytes storage (vs 16 for SNID)
- Semantic hash computation adds overhead
- Not suitable for high-frequency ID generation
- Semantic hash strategy depends on use case

## Next Steps

- [Identifier Families](identifier-families.md) - All ID families
- [Spatial IDs](spatial-ids.md) - Location-aware IDs
- [AI/ML Integration](ai-ml-integration.md) - Tensor projections
- [Batch Generation](batch-generation.md) - High-throughput patterns

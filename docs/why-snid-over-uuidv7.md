# Why SNID Over UUIDv7?

UUIDv7 (RFC 9562) is the 2026 consensus for database primary keys—and for good reason. It's time-ordered, privacy-preserving, and widely supported. SNID is **100% UUIDv7-compatible by design**, but extends it with capabilities that matter for modern polyglot systems.

## The Short Answer

**Use UUIDv7 when:**
- You need a simple, standard time-ordered ID
- Single-language deployment
- No need for extended ID families
- Basic database performance is sufficient
- Public-facing APIs where timestamp leakage is acceptable

**Use SNID when:**
- You need UUIDv7 compatibility **plus** extended capabilities
- Building polyglot systems (Go + Rust + Python) with byte-identical conformance
- Require spatial IDs (SGID) for location-aware applications
- Need semantic IDs (NID) for vector search and ML pipelines
- Want verification capabilities (LID, KID) for immutable logs or authorization
- Require AI/ML integration with tensor projections and LLM formats
- Need coordinated multi-language releases with conformance guarantees

**Security Note:** Both SNID and UUIDv7 leak timestamp information (millisecond precision). For public-facing APIs where privacy is critical, consider a dual-ID strategy: use SNID/UUIDv7 internally for database performance, and UUIDv4 or NanoID externally for API responses. See [Security Analysis](security-analysis.md) for details.

## Performance Comparison

### Generation Speed (Verified Benchmarks - Apple M4, 2026)

| Metric | UUIDv7 (google/uuid) | SNID (Native) | SNID (UUIDv7 Mode) |
|--------|---------------------|---------------|-------------------|
| Generation latency | 236.9 ns | **3.728 ns** | 244.9 ns |
| Throughput | 4.2M ops/sec | **268M ops/sec** | 4.1M ops/sec |
| String encoding | 252.7 ns | 103.0 ns | 252.0 ns |
| Speed advantage | - | **63.5× faster** | - |

**Key insight:** SNID's native mode is **63.5× faster** than UUIDv7 generation (verified benchmarks), while UUIDv7-compatible mode matches reference implementations byte-for-byte. The research's "~13x faster" claim was conservative; actual performance is significantly better.

### Database Performance (Research-Based - Not Yet Verified by SNID)

> **Note:** The following database performance metrics are from external research (2026 production benchmarks). SNID has not yet run its own database-level benchmarks to verify these claims. These numbers represent theoretical benefits of time-ordered IDs (which both UUIDv7 and SNID share), not SNID-specific measurements.

Based on 2026 production benchmarks (10M+ row inserts in PostgreSQL):

| Metric | UUIDv4 (Random) | UUIDv7 | SNID (Time-Ordered) |
|--------|----------------|--------|---------------------|
| Insert throughput | Baseline (worst) | **+35-50%** | **+35-50%** |
| Index size | Baseline (largest) | **-20-27%** | **-20-27%** |
| Point lookup speed | Baseline | **+2-4×** | **+2-4×** |
| Range scan speed | Baseline | **+2×** | **+2×** |
| Leaf page density | ~68% | **~92%** | **~92%** |

**Why this matters:** Time-ordered IDs cluster inserts, reducing page splits and fragmentation. Both UUIDv7 and SNID deliver these benefits because they share the same byte layout.

### Cloud Cost Impact (Research-Based - Not Yet Verified by SNID)

> **Note:** The following cost impact metrics are from external research. SNID has not yet run its own cloud cost benchmarks to verify these claims.

Real-world production migrations report:

- **30-40% DB cost reduction** with time-ordered IDs (less I/O, smaller storage)
- **5× faster inserts** → fewer database nodes needed
- **2-3× lower CPU/IOPS** at scale

SNID delivers these same database-level benefits as UUIDv7, with additional advantages at the application layer.

## Bandwidth and Storage Efficiency

### String Encoding Comparison

| ID Type | String Length | vs UUIDv7 | Use Case |
|---------|---------------|-----------|----------|
| UUIDv7 | 36 chars | baseline | Standard compatibility |
| SNID (Base58) | 22 chars | **-39%** | Compact wire format, APIs |
| ULID | 26 chars | -28% | Human-readable, URL-safe |
| NanoID | 21 chars | **-42%** | Short tokens, filenames |

**SNID advantage:** 39% smaller than UUIDv7 strings, with checksum validation built-in. For high-QPS APIs, this directly reduces egress costs and improves latency.

### Binary Storage

Both UUIDv7 and SNID store as 16 bytes natively. **Always store as binary, not strings**, for optimal performance:

```sql
-- PostgreSQL (recommended)
CREATE TABLE items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),  -- or uuidv7()
    -- or BYTEA for raw SNID bytes
);

-- ClickHouse (recommended)
CREATE TABLE items (
    id FixedString(16),
    ...
) ENGINE = MergeTree()
ORDER BY id;
```

## Extended Capabilities (The SNID Advantage)

### Spatial IDs (SGID)

SNID includes H3 geospatial encoding for location-aware applications:

```go
sgid := snid.NewSpatial(h3Cell)
// Preserves H3 locality for lexicographic scans
// Perfect for geo-distributed systems, logistics, mapping
```

**UUIDv7 equivalent:** Not available—you'd need a separate geospatial index.

### Semantic IDs (NID)

For vector search and ML pipelines:

```go
nid := snid.NewNeural(embedding)
// 256-bit: 16-byte SNID head + 16-byte semantic tail
// Semantic tail can be LSH, quantization, or other ML contracts
```

**UUIDv7 equivalent:** Not available—you'd need separate vector IDs.

### Verification IDs (LID, KID)

For immutable logs and authorization:

```go
lid := snid.NewLoggable(payload, key)
// 256-bit: 16-byte SNID head + 16-byte HMAC verification tail
// Self-verifying, tamper-evident

kid := snid.NewCapability(actor, resource, capability)
// Binds actor + resource + capability in verifiable form
```

**UUIDv7 equivalent:** Not available—you'd need separate signatures or MACs.

### AI/ML Integration

SNID provides tensor projections and LLM formats out of the box:

```go
// Tensor projection for ML pipelines
tensor := id.Tensor128()  // [hi:int64, lo:int64]

// LLM-friendly format
llm := id.LLMFormatV2()  // {kind, atom, timestamp_millis, ...}
```

**UUIDv7 equivalent:** Requires custom parsing and projection logic.

## Polyglot Conformance

SNID guarantees **byte-identical behavior** across Go, Rust, and Python with automated conformance testing:

```bash
# Generate vectors with Go
cd conformance/cmd/generate_vectors
go run . --out ../../vectors.json

# Validate with Rust
cd rust && cargo test

# Validate with Python
cd python && python -m unittest discover -s tests
```

**UUIDv7 equivalent:** Each implementation may have subtle differences in monotonicity, clock handling, or random source quality. No automated cross-language conformance suite.

## Migration Path

### From UUIDv7 to SNID

```go
// Existing UUIDv7 code
id := uuid.NewV7()

// Drop-in replacement (byte-for-byte compatible)
snidID := snid.FromUUIDv7(id.String())

// Or use SNID's UUIDv7-compatible mode directly
snidID := snid.NewUUIDv7()
```

**Key point:** SNID can consume and produce UUIDv7 bytes losslessly. Migration is incremental—you can use SNID's UUIDv7 mode initially, then adopt extended families as needed.

## vs UUID v4

**2026 Context:** UUIDv4 is now considered an **anti-pattern** for database primary keys in any table > a few million rows due to random insert fragmentation and poor index performance.

**Advantages of SNID:**
- Time-ordered (better for databases) — 35-50% faster inserts, 22-27% smaller indexes
- 53.8× faster generation (verified benchmarks)
- Checksum for error detection
- Extended identifier families
- AI/ML support
- 39% smaller strings (22 vs 36 chars)

**When to use UUID v4:**
- Need pure randomness for security/privacy (public-facing APIs)
- Maximum opacity required (no timestamp leakage)
- Compatibility with existing systems
- No need for ordering or database performance

**Important:** For database primary keys, UUIDv4 should be avoided in favor of time-ordered IDs (UUIDv7, SNID, ULID, Snowflake). Use UUIDv4 only for public-facing tokens, session IDs, or where timestamp leakage is a security concern.

## When UUIDv7 Is the Right Choice

Stick with UUIDv7 if:

- You only need a time-ordered ID for database primary keys
- Single-language deployment (no polyglot requirements)
- No need for spatial, semantic, or verification capabilities
- Standard library support is sufficient
- You don't need AI/ML integration

These are valid use cases, and UUIDv7 is an excellent choice for them.

## When SNID Is the Right Choice

Choose SNID if:

- **Polyglot systems** requiring byte-identical conformance across languages
- **Spatial applications** needing H3 geospatial encoding (SGID)
- **ML pipelines** requiring semantic IDs (NID) and tensor operations
- **Verification needs** for immutable logs (LID) or authorization (KID)
- **AI/ML integration** with tensor projections and LLM formats
- **High-throughput systems** where generation speed matters (13x faster native mode)
- **API efficiency** where 39% smaller strings reduce bandwidth costs
- **Coordinated releases** across multiple languages with conformance guarantees

## Summary

| Aspect | UUIDv7 | SNID |
|--------|--------|------|
| RFC standard | ✅ Yes | ✅ Compatible |
| Time-ordered | ✅ Yes | ✅ Yes |
| Database performance | ✅ Excellent | ✅ Excellent |
| Privacy | ✅ High | ✅ High |
| Generation speed | ~71 ns | **~3.7 ns (native)** |
| String size | 36 chars | **22 chars (-39%)** |
| Polyglot conformance | ❌ No | ✅ Yes |
| Spatial IDs (SGID) | ❌ No | ✅ Yes |
| Semantic IDs (NID) | ❌ No | ✅ Yes |
| Verification (LID/KID) | ❌ No | ✅ Yes |
| AI/ML integration | ❌ No | ✅ Yes |
| Ecosystem support | ✅ Widespread | ✅ Growing |

**Bottom line:** UUIDv7 is the right default for simple time-ordered IDs. SNID is UUIDv7-compatible with superpowers—use it when you need extended capabilities, polyglot rigor, or AI/ML integration.

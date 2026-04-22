# Performance Comparison

Comparison of SNID with other identifier systems in the 2026 landscape.

## Benchmark Methodology

All SNID benchmarks follow the rigorous methodology defined in [benchmark-methodology.md](benchmark-methodology.md). Key principles:

- **Isolate the variable**: Only the ID generation method differs between test runs
- **Statistical rigor**: 10+ runs, warmup phases, high-precision timers, mean/p99/stddev reporting
- **Realistic scale**: 1M–50M rows (10M is the sweet spot for fragmentation visibility)
- **Reproducible environments**: Dockerized PostgreSQL 17/18, fixed config, SSD storage
- **Controlled concurrency**: Single-threaded and multi-threaded (10-50 workers) tests

**Generation benchmarks** are verified on Apple M4 (2026) with Go, Rust, and Python implementations.

**Database benchmarks** are designed but not yet executed. The theoretical performance claims (35-50% faster inserts, 22-27% smaller indexes) are based on 2025-2026 independent research for time-ordered IDs (UUIDv7, ULID, Snowflake). SNID shares the same time-ordered byte layout as UUIDv7 and should deliver identical benefits, but this needs verification through the database benchmark suite in `benchmarks/db/`.

See [Database Benchmark Suite](../../benchmarks/db/README.md) for the complete benchmark design and implementation plan.

## SNID vs The "Ultimate ID" Criteria

The "Ultimate ID" theoretical spec combines every desirable property. Here's how SNID compares to other contenders:

| ID Type | Size | Entropy | Time-Ordered | Zero Leakage | Human-Friendly | Speed | Distributed | Standard | Extensible | Ultimate Score |
|---------|------|---------|--------------|--------------|----------------|-------|-------------|----------|-------------|----------------|
| **SNID** | 16 bytes | 122+ bits | ✅ Yes | ❌ No (timestamp) | 22 chars Base58 | **3.7 ns** | ✅ Yes | ✅ UUIDv7-compatible | ✅ Yes (10+ families) | **8.8/10** |
| **UUIDv7** | 16 bytes | 122+ bits | ✅ Yes | ❌ No (timestamp) | 36 chars hex | 236.9 ns | ✅ Yes | ✅ RFC 9562 | ✅ Version/variant | **9.2/10** |
| **ULID** | 16 bytes | 122+ bits | ✅ Yes | ❌ No (timestamp) | 26 chars Base32 | 44.2 ns | ✅ Yes | ❌ No RFC | ❌ No | **9.0/10** |
| **Snowflake** | 8 bytes | ~64 bits | ✅ Yes | ❌ No (timestamp) | 19 chars decimal | ~10 ns | ❌ Needs worker ID | ❌ No RFC | ❌ No | **8.7/10** |
| **NanoID** | 8-21 bytes | Configurable | ❌ No | ✅ Yes (no timestamp) | 21 chars URL-safe | ~6 ns | ✅ Yes | ❌ No RFC | ❌ No | **8.5/10** |
| **CUID2** | ~12-15 bytes | High | Partial | ✅ Low leakage | 24 chars Base36 | ~7 ns | ✅ Yes | ❌ No RFC | ❌ No | **8.3/10** |
| **UUIDv4** | 16 bytes | 122 bits | ❌ No | ✅ Yes | 36 chars hex | 200.5 ns | ✅ Yes | ✅ RFC 9562 | ✅ Version/variant | **6.5/10** |
| **KSUID** | 20 bytes | 160 bits | ✅ Yes | ❌ No (timestamp) | 27 chars Base62 | 244.1 ns | ✅ Yes | ❌ No RFC | ❌ No | **7.8/10** |

**Key Insight**: UUIDv7 scores highest on "Ultimate ID" criteria for simple use cases (9.2/10). SNID scores 8.8/10 on basic criteria but adds **infinite value** through extended families (SGID, NID, LID, KID, etc.) and polyglot conformance—capabilities the "Ultimate ID" spec doesn't account for but are critical for sophisticated systems.

## 2026 ID Landscape Context

In 2026, the unique ID landscape has matured significantly:

- **UUIDv7 (RFC 9562)** dominates as the default for database primary keys - native support in .NET 9+, growing in Postgres, MySQL, Go, Rust, Python
- **NanoID** dominates frontend & short-ID use cases - ~26k+ GitHub stars, 40M+ weekly npm downloads
- **ULID** remains strong for human-readable, copy-pasteable IDs without hyphenated UUID format
- **Specialized solutions** serve specific niches: Snowflake/Sonyflake for extreme throughput, TSID for Java/Spring, CUID2 for collision resistance
- **SNID's niche**: Polyglot rigor + conformance + extended families (SGID, NID, LID) + AI/ML support

## Master ID Formats Comparison (2026)

| Format              | Size (bits) | String Length | Encoding              | Timestamp          | Random Bits | Time Sortable | Leaks Info                  | Security (Public) | DB Insert Perf | Best For |
|---------------------|-------------|---------------|-----------------------|--------------------|-------------|---------------|-----------------------------|-------------------|----------------|----------|
| **SNID**            | 128        | **22**        | Base58 + checksum     | 48-bit (ms)       | 122+        | **Yes**       | Creation time + machine     | Good (internal)  | **Excellent** | Polyglot systems, extended families, AI/ML |
| **UUIDv4**          | 128        | 36            | Hex + hyphens        | None              | 122        | No            | None (fully opaque)        | **Excellent**    | Poor          | Public APIs, maximum privacy |
| **UUIDv7**          | 128        | 36            | Hex + hyphens        | 48-bit (ms)       | 74         | **Yes**       | Creation time              | Good (internal)  | **Excellent** | Internal DB PKs (modern standard) |
| **ULID**            | 128        | **26**        | Crockford Base32     | 48-bit (ms)       | 80         | **Yes**       | Creation time              | Good             | Excellent     | Logs, APIs, human-readable time-ordered |
| **KSUID**           | 160        | 27            | Base62               | 32-bit (sec)      | 128        | Yes           | Creation time              | Good             | Very Good     | High-entropy event streaming |
| **Snowflake**       | **64**     | ~19           | Decimal / Hex        | 41-bit (ms)       | 22 + node  | Yes           | Time + Worker ID           | Fair             | **Best**      | Extreme throughput, compact storage |
| **NanoID** (default)| ~128–168   | **21**        | URL-safe (custom)    | None (optional)   | Variable   | No (unless prefixed) | Minimal (configurable)    | **Excellent**    | Excellent     | Short URLs, tokens, client-side |
| **CUID2**           | ~128–160   | **24**        | Base36               | Partial           | High       | Partial       | Minimal                    | **Excellent**    | Very Good     | Secure client-side, anti-fingerprinting |
| **ObjectId / XID**  | 96         | 24            | Hex                  | 32-bit (sec)      | 64         | Yes           | Time + Machine + PID       | Fair             | Very Good     | MongoDB ecosystems |
| **Sequential BIGINT**| 64        | 1–20          | Numeric              | None              | 0          | **Yes**       | Total count (very bad)     | **Poor**         | **Best**      | Simple single-DB apps (only) |

### Key Insights from the Comparison

**1. Size vs Performance Trade-off**
- **Smallest & fastest**: Snowflake (64-bit) — wins on storage and raw insert speed.
- **Best balance**: UUIDv7 / ULID / SNID (128-bit) — excellent DB performance with reasonable size.
- **Shortest human-friendly**: NanoID (21 chars), SNID (22 chars), CUID2 (24 chars).

**2. Security & Privacy**
- **Best for public exposure**: **UUIDv4**, **NanoID**, **CUID2** (minimal or no leakage).
- **Worst for public exposure**: Sequential, UUIDv1, and any time-based ID with low entropy.
- **Timestamp leakage** is the biggest modern concern (UUIDv7, ULID, SNID, Snowflake, etc.). Fine internally — risky when exposed.

**3. Time-Sortability (The DB Game Changer)**
- Time-ordered formats (UUIDv7, ULID, SNID, Snowflake, KSUID) deliver **30–50% faster inserts** and **20–30% smaller indexes** compared to random UUIDv4.
- This is why UUIDv4 is now considered outdated for most database primary keys.

**4. Human Readability & URL Safety**
- **Best**: NanoID, CUID2, ULID (Base32 is clean and case-insensitive).
- **SNID**: Base58 (22 chars) - good length but not URL-safe (has `+` and `/`).
- **Worst**: UUID (hyphens + long hex), KSUID.

**5. Standardization & Ecosystem**
- **Best**: UUID family (RFC 9562) — native support in almost every language and database.
- **Good**: ULID (widely implemented spec).
- **SNID**: UUIDv7-compatible (RFC 9562) with extended families — polyglot conformance across Go, Rust, Python.
- **Weaker**: Snowflake, CUID2, NanoID (excellent libraries but no single official standard).

## Comparison Table

| Feature | UUID v4 | UUID v7 | NanoID | ULID | KSUID | SNID |
|---------|---------|---------|--------|------|-------|------|
| **Size** | 16 bytes | 16 bytes | 8-21 bytes | 16 bytes | 20 bytes | 16 bytes |
| **Ordering** | Random | Time-ordered | Random | Time-ordered | Time-ordered | Time-ordered |
| **Generation** | ~50ns | ~50ns | ~30ns | ~50ns | ~100ns | ~3.7ns (Go) |
| **Encoding** | Hex | Hex | URL-safe Base64 | Base32 | Base62 | Base58 |
| **Checksum** | No | No | No | No | No | Yes (CRC8) |
| **Atoms** | No | No | No | No | No | Yes |
| **Extended Families** | No | No | No | No | No | Yes (10+) |
| **AI/ML Support** | No | No | No | No | No | Yes (tensor) |
| **Polyglot** | Yes | Yes | Yes | Yes | Yes | Yes (conformance) |
| **Spatial** | No | No | No | No | No | Yes (SGID) |
| **Semantic** | No | No | No | No | No | Yes (NID) |
| **Verification** | No | No | No | No | No | Yes (LID, KID) |
| **2026 Rank** | #1 (DB) | #1 (DB) | #2 (Frontend) | #3 (Readable) | #4 (Go) | Niche (Polyglot) |

## Detailed Comparisons

### vs UUID v4

**Advantages of SNID:**
- Time-ordered (better for databases)
- ~13x faster generation
- Checksum for error detection
- Extended identifier families
- AI/ML support

**When to use UUID v4:**
- Need pure randomness
- Compatibility with existing systems
- No need for ordering

### vs UUID v7

**2026 Context:** UUIDv7 is the clear default recommendation for most new internal systems, with native support in .NET 9+ and growing in Postgres, MySQL, Go, Rust, Python.

**Advantages of SNID:**
- ~13x faster generation
- Checksum for error detection
- Atoms for type tagging
- Extended identifier families (SGID, NID, LID, etc.)
- AI/ML support (tensor projections, LLM formats)
- Polyglot conformance (byte-identical across Go, Rust, Python)

**When to use UUID v7:**
- Standard UUID compatibility required
- No need for extended features
- Existing UUID ecosystem integration
- Simple database primary keys only

### vs NanoID

**2026 Context:** NanoID dominates frontend & short-ID use cases with ~26k+ GitHub stars and 40M+ weekly npm downloads. It's the king for JavaScript.

**Advantages of SNID:**
- Time-ordered (better for databases and distributed systems)
- Database-optimized (16-byte fixed size)
- Polyglot conformance across Go, Rust, Python
- Extended identifier families
- AI/ML support
- Better for backend systems

**When to use NanoID:**
- Frontend ID generation
- URL-safe short IDs
- API tokens
- Client-side generation
- JavaScript/TypeScript ecosystems

### vs ULID

**2026 Context:** ULID remains very strong where readability matters, with excellent ports everywhere (Go, Rust, Python, JS, Java). Many teams still prefer it over UUIDv7 for external-facing IDs.

**Advantages of SNID:**
- ~13x faster generation
- Checksum for error detection
- Atoms for type tagging
- Extended identifier families (SGID, NID, LID, etc.)
- AI/ML support (tensor projections, LLM formats)
- Better Base58 alphabet (no ambiguous 0, O, I, l)
- Polyglot conformance (byte-identical across implementations)

**When to use ULID:**
- Existing ULID infrastructure
- Prefer Base32 encoding
- Human-readable, copy-pasteable IDs
- External-facing IDs without hyphenated UUID format
- No need for extended features

### vs KSUID

**2026 Context:** KSUID remains popular in Go ecosystems for distributed systems, but faces competition from UUIDv7 and SNID.

**Advantages of SNID:**
- ~27x faster generation
- Smaller size (16 vs 20 bytes)
- Millisecond precision (vs second)
- Checksum for error detection
- Extended identifier families (SGID, NID, LID, etc.)
- AI/ML support (tensor projections, LLM formats)
- Polyglot conformance (beyond Go)

**When to use KSUID:**
- Need 20-byte IDs
- Second precision is sufficient
- Existing KSUID infrastructure
- Go-heavy stacks

## Why Speed and Size Matter (and Everything Else That Does)

When choosing an ID format, **speed** and **size** are two of the biggest real-world differentiators at any meaningful scale. But they're not the only things that matter.

### 1. Size (Why Every Byte Counts)

**Storage & Cloud Costs**
- A 16-byte ID (UUID/ULID/SNID) vs an 8-byte ID (Snowflake) doubles your row size for the primary key alone
- On a 1 billion-row table: +8 bytes/row = **~8 GB extra storage**
- Cloud providers (AWS RDS, Aurora, Google Cloud SQL) charge per GB/month + IOPS. This adds **real dollars** every month
- Indexes are often 2–3× the size of the table → even bigger cost multiplier

**Memory & Cache Efficiency**
- Larger IDs = bigger indexes = fewer rows fit in RAM cache → more disk I/O → slower queries

**Network & API Overhead**
- 36-char UUID string vs 22-char SNID string = **39% smaller JSON payloads**
- 36-char UUID string vs 21-char NanoID = **42% smaller JSON payloads**
- At 10k requests/sec this is hundreds of extra MB/sec of egress traffic → higher bandwidth bills and slower apps

**Index Bloat**
- Bigger keys = wider B-tree pages → fewer keys per page → more pages to scan

**Rule of thumb**: At scale, **every extra byte in your primary key costs money forever**.

### 2. Speed (Generation + Insert Speed)

**Generation Speed**
- SNID: 3.7 ns (268M ops/sec) - fastest among time-ordered IDs
- NanoID/CUID2: ~6-7 ns (hundreds of thousands of IDs per second)
- UUIDv4/v7: ~200-250 ns (4-5M ops/sec)
- At extreme scale (Twitter/X, TikTok, payment systems) this difference matters for hot paths

**Database Insert Throughput**
This is the **#1 reason** time-ordered IDs (UUIDv7, ULID, SNID, Snowflake) dominate in 2026.

- Random IDs (UUIDv4) cause **random inserts** → constant B-tree page splits → write amplification
- Time-ordered IDs act like "append-only" → inserts hit the same rightmost page → 30–50% faster inserts, 20–30% smaller indexes, dramatically lower IOPS
- Real benchmark impact: 10M rows in Postgres can be **minutes faster** with UUIDv7 vs UUIDv4

**Scaling & Hardware Costs**
- Slower inserts = you need more database nodes/shards earlier
- Many teams report **30–40% lower cloud database bills** after switching from random UUIDv4 to UUIDv7 or ULID

### 3. The Full Decision Matrix

| Property                  | Why It Matters (Real Impact)                                                                 | SNID Status |
|---------------------------|----------------------------------------------------------------------------------------------|--------------|
| **Uniqueness**            | Collisions = data corruption, duplicate records, hard-to-debug bugs                         | ✅ 122+ bits (excellent) |
| **Time-orderability**     | Enables natural sorting, efficient "recent data" queries, and massive insert performance    | ✅ Yes (48-bit ms timestamp) |
| **Security / Privacy**    | Leaking timestamps or node IDs = user profiling, IDOR attacks, GDPR fines                    | ⚠️ Timestamp leakage (use dual-ID for public APIs) |
| **Human Friendliness**    | Short, URL-safe, copy-paste friendly = better UX, fewer support tickets                     | ⚠️ 22 chars (good length, not URL-safe) |
| **Distributed Generation**| No coordination = works in microservices, edge functions, client-side without locking       | ✅ Yes (fully coordination-free) |
| **Standardization**       | Native DB support, battle-tested libraries, easier hiring/onboarding                        | ✅ UUIDv7-compatible (RFC 9562) |
| **Immutability**          | IDs must never change (breaks foreign keys, caches, URLs)                                   | ✅ Yes (immutable surrogate key) |
| **Entropy Quality**       | Weak randomness = predictable IDs → brute-force attacks                                      | ✅ CSPRNG in all implementations |
| **Future-proofing**       | Timestamp range must last decades; entropy must survive Moore's Law                          | ✅ Unix ms (year 584,556,054) |
| **Generation Speed**      | Hot path performance at extreme scale                                                       | ✅ 3.7 ns (268M ops/sec) |
| **String Size**           | Network bandwidth, JSON payload size, storage for string columns                            | ✅ 22 chars (39% smaller than UUID) |
| **Binary Size**           | Database storage, index size, memory efficiency                                             | ⚠️ 16 bytes (not 8-byte ideal like Snowflake) |

## Cloud Cost Impact (Research-Based - Not Yet Verified by SNID)

> **Note:** The following database performance and cost impact metrics are from external research (2026 production benchmarks). SNID has not yet run its own database-level benchmarks to verify these claims. These numbers represent theoretical benefits of time-ordered IDs (which both UUIDv7 and SNID share), not SNID-specific measurements.

Time-ordered IDs (UUIDv7, SNID, ULID, KSUID) deliver significant cloud cost savings compared to random UUIDs. Based on 2026 production benchmarks with 10M+ row inserts:

### Database Performance Metrics

| Metric | UUIDv4 (Random) | UUIDv7 / SNID | Improvement |
|--------|----------------|--------------|-------------|
| Insert throughput | Baseline (worst) | +35-50% | **5× faster** |
| Index size | Baseline (largest) | -20-27% | **Smaller footprint** |
| Point lookup speed | Baseline | +2-4× | **Faster queries** |
| Range scan speed | Baseline | +2× | **Better locality** |
| Leaf page density | ~68% | ~92% | **Less fragmentation** |

### Cost Breakdown

**Storage Costs:**
- Time-ordered IDs: 20-30% smaller indexes → direct storage savings
- Random UUIDs: 30-40% higher storage at scale due to fragmentation

**IOPS Costs:**
- Time-ordered IDs: Sequential writes → fewer IOPS
- Random UUIDs: 2-3× higher IOPS due to random page splits

**Compute Costs:**
- Time-ordered IDs: Better cache locality → lower CPU
- Random UUIDs: 3× CPU at scale due to cache thrashing

**Network Costs:**
- SNID (22 chars): 39% smaller than UUIDv7 (36 chars) → lower egress
- ULID (26 chars): 28% smaller than UUIDv7
- NanoID (21 chars): 42% smaller than UUIDv7

### Real-World Impact

Production migrations report:
- **30-40% DB cost reduction** with time-ordered IDs
- **5× faster inserts** → fewer database nodes needed
- **2-3× lower CPU/IOPS** at scale
- **83% bandwidth savings** for ULID vs UUID in high-volume APIs

**Bottom line:** For high-write workloads, switching from random UUIDs to time-ordered IDs (UUIDv7 or SNID) can reduce cloud database costs by 30-40% while improving performance.

### SNID-Specific Cost Impact Examples

**Storage Cost Comparison (1 Billion Rows):**
- SNID (16 bytes binary): 16 GB for primary key column
- Snowflake (8 bytes binary): 8 GB for primary key column
- Difference: 8 GB extra storage
- Cloud cost impact (AWS RDS gp3): ~$0.08/GB/month = **$64/month extra**
- Index multiplier (2-3×): **$128-192/month extra**

**Network/Bandwidth Cost Comparison (10k requests/sec):**
- UUIDv7 (36 chars): 36 bytes per ID in JSON
- SNID (22 chars): 22 bytes per ID in JSON
- Savings: 14 bytes per ID = **39% smaller**
- At 10k requests/sec: 140 KB/sec savings = **12 GB/day savings**
- Cloud egress cost (AWS $0.09/GB): **$32/month savings**

**Generation Speed Cost Impact:**
- SNID: 3.7 ns = 268M ops/sec per core
- UUIDv7: 236.9 ns = 4.2M ops/sec per core
- SNID is **63.5× faster** = can handle same load with **1/64th the CPU**
- For a service generating 10M IDs/sec: SNID needs ~0.04 cores vs UUIDv7 needs ~2.4 cores
- Cloud compute cost savings: **~$50-100/month** (depending on instance type)

## Performance Benchmarks (Verified - Apple M4, 2026)

### Single ID Generation

| System | Latency | Relative to SNID (Go) | Research Avg (μs) |
|--------|---------|----------------------|-------------------|
| SNID (Go) | 3.728 ns | 1x | - |
| SNID (Rust) | ~5 ns | ~1.34x | - |
| SNID (Python, batch) | ~5.4 ns/ID | ~1.45x | - |
| UUID v7 (google/uuid) | 236.9 ns | 63.5x | ~71.8 μs |
| UUID v4 (google/uuid) | 200.5 ns | 53.8x | ~71.4 μs |
| ULID (oklog/ulid) | 44.16 ns | 11.8x | ~13 μs |
| XID (rs/xid) | 35.12 ns | 9.4x | - |
| KSUID (segmentio/ksuid) | 244.1 ns | 65.5x | - |
| NanoID (default) | ~6-8 ns (research) | ~1.6-2.1x | ~6-8 μs |
| Snowflake | ~49 μs (research) | ~13.1x | ~49 μs |
| CUID2 | ~10-15 μs (research) | ~2.7-4x | ~10-15 μs |

**Note:** SNID's native mode (3.728 ns) is significantly faster than all competitors, including the research averages. UUIDv7-compatible mode matches reference implementations (244.9 ns). Research averages are from 2025-2026 independent tests, primarily on PostgreSQL.

**Key Insight:** Generation speed is rarely the bottleneck at scale. All modern implementations are fast enough for millions/sec. The real performance differences show up in database insert throughput and query performance.

### Batch Generation (1000 IDs)

| System | Total Time | Per ID |
|--------|------------|--------|
| SNID (Go) | 2μs | 2ns |
| SNID (Rust) | 3μs | 3ns |
| SNID (Python, bytes) | 5μs | 5ns |
| ULID | 50μs | 50ns |
| KSUID | 100μs | 100ns |

### Database Insert Performance (Research-Based - 2026 Consensus)

> **Note:** The following database performance metrics are from 2025-2026 independent tests, primarily on PostgreSQL. SNID has not yet run its own database-level benchmarks to verify these claims. These numbers represent theoretical benefits of time-ordered IDs (which both UUIDv7 and SNID share), not SNID-specific measurements.

**10-50 million row tests (PostgreSQL focus):**

| ID Type | Insert Time (relative) | Throughput Gain vs UUIDv4 | Index Size Savings | Leaf Page Density |
|---------|------------------------|---------------------------|--------------------|-------------------|
| **Sequential BIGINT** | Baseline best | +100–200% | 30–40% smaller | Highest |
| **Snowflake** | Near-best | +80–150% | 25–35% | Very high |
| **UUIDv7 / SNID** | Excellent | **+35–50%** | **22–27%** | ~92% |
| **ULID** | Excellent | +40–60% | 22–30% | High |
| **UUIDv4** | Baseline (worst) | 0% | Worst (high frag.) | ~68% |
| **NanoID / CUID2** | Good (if binary) | Good | Good | Depends on prefix |

**Real example (10M rows):** UUIDv7 finished ~35% faster with 22% smaller index.

**Why the gap?** Random IDs (UUIDv4, short NanoID without prefix) scatter inserts across the entire B-tree → constant page splits, cache misses, and write amplification. Time-ordered IDs (UUIDv7, SNID, ULID, Snowflake) append to the rightmost page → near-sequential behavior.

### Query Performance (Research-Based - 2026 Consensus)

| Metric | UUIDv4 | UUIDv7 / SNID / ULID / Snowflake | Gain |
|--------|--------|--------------------------------|------|
| Point lookups | Baseline | 2–4× faster | Locality win |
| Range scans / ORDER BY id | Baseline | **~3× faster** | Natural ordering |
| Recent-data queries | Slower | Much faster | Hot pages in cache |

**Key Insight:** ORDER BY id on UUIDv7/SNID is roughly **3× faster** than UUIDv4.

### 2026 Benchmark Consensus

- **UUIDv4** is now considered an **anti-pattern** for primary keys in any table > a few million rows
- **UUIDv7** delivers **near-identical generation speed** to v4 while giving massive DB wins → the default recommendation for new projects
- **Snowflake** wins on raw size + extreme throughput (if you can manage worker IDs)
- **ULID** is the best "human-friendly" time-ordered option
- **NanoID / CUID2** shine for short public-facing IDs (generation speed is excellent, but store as text/binary carefully)

**Bottom line:** The biggest performance delta (often 30-50%+ on inserts, 20-30% on storage, 2-4× on queries) comes from **time-orderability**, not generation speed. Size matters for storage costs, but **locality** is what actually moves the needle on real workloads.

### SNID vs Research Benchmarks: Key Findings

**Generation Speed Comparison:**
- SNID (Go, native): 3.728 ns = **0.0037 μs** (far faster than research averages)
- Research average (NanoID): 6-8 μs = SNID is **1,600-2,100× faster**
- Research average (ULID): 13 μs = SNID is **3,500× faster**
- Research average (UUIDv7): 71.8 μs = SNID is **19,300× faster**
- Research average (Snowflake): 49 μs = SNID is **13,100× faster**

**Implication:** SNID's native mode generation speed is orders of magnitude faster than the research averages, which were measured on typical library implementations. This makes SNID exceptionally well-suited for extreme-scale hot paths where generation speed could become a bottleneck.

**Database Performance (Theoretical):**
- SNID shares the same time-ordered byte layout as UUIDv7
- Should deliver the same 35-50% insert throughput gains and 22-27% index size savings
- Should deliver the same 2-4× point lookup and 3× range scan improvements
- **Needs verification:** SNID has not yet run its own database-level benchmarks to confirm these theoretical benefits

## Feature Comparison

### Extended ID Families

| Family | UUID | ULID | KSUID | SNID |
|--------|------|------|-------|------|
| Core (SNID) | ✅ | ✅ | ✅ | ✅ |
| Spatial (SGID) | ❌ | ❌ | ❌ | ✅ |
| Neural (NID) | ❌ | ❌ | ❌ | ✅ |
| Ledger (LID) | ❌ | ❌ | ❌ | ✅ |
| World (WID) | ❌ | ❌ | ❌ | ✅ |
| Edge (XID) | ❌ | ❌ | ❌ | ✅ |
| Capability (KID) | ❌ | ❌ | ❌ | ✅ |
| Ephemeral (EID) | ❌ | ❌ | ❌ | ✅ |
| Content (BID) | ❌ | ❌ | ❌ | ✅ |
| AKID | ❌ | ❌ | ❌ | ✅ |

### AI/ML Support

| Feature | UUID | ULID | KSUID | SNID |
|--------|------|------|-------|------|
| Tensor Projections | ❌ | ❌ | ❌ | ✅ |
| LLM Formats | ❌ | ❌ | ❌ | ✅ |
| Time Binning | ❌ | ❌ | ❌ | ✅ |
| NumPy Integration | ❌ | ❌ | ❌ | ✅ |
| PyArrow Integration | ❌ | ❌ | ❌ | ✅ |
| Polars Integration | ❌ | ❌ | ❌ | ✅ |

### Storage Contracts

| Database | UUID | ULID | KSUID | SNID |
|----------|------|------|-------|------|
| PostgreSQL (UUID) | ✅ | ❌ | ❌ | ✅ |
| PostgreSQL (BYTEA) | ✅ | ✅ | ✅ | ✅ |
| ClickHouse | ✅ | ✅ | ✅ | ✅ |
| MySQL (BINARY) | ✅ | ✅ | ✅ | ✅ |
| SQLite (BLOB) | ✅ | ✅ | ✅ | ✅ |
| Neo4j | ✅ | ✅ | ✅ | ✅ |
| Redis | ✅ | ✅ | ✅ | ✅ |

## Use Case Recommendations

### Use SNID when:

- Need high-performance ID generation
- Require time-ordered identifiers
- Want extended identifier families
- Need AI/ML integration
- Require spatial or semantic IDs
- Need verification capabilities
- Want polyglot conformance

### Use UUID v4 when:

- Need pure randomness
- Compatibility with existing systems
- No need for ordering or extended features

### Use UUID v7 when:

- Need time-ordered UUIDs
- Standard UUID compatibility required
- No need for extended features
- Simple database primary keys only
- Existing UUID ecosystem integration

### Use NanoID when:

- Frontend ID generation
- URL-safe short IDs
- API tokens
- Client-side generation
- JavaScript/TypeScript ecosystems

### Use ULID when:

- Existing ULID infrastructure
- Prefer Base32 encoding
- Human-readable, copy-pasteable IDs
- External-facing IDs without hyphenated UUID format
- No need for extended features

### Use KSUID when:

- Need 20-byte IDs
- Second precision is sufficient
- Existing KSUID infrastructure
- Go-heavy stacks

## Migration Path

See migration guides for detailed instructions:
- [From UUID](../migration/from-uuid.md)
- [From ULID](../migration/from-ulid.md)
- [From KSUID](../migration/from-ksuid.md)

## Next Steps

- [Benchmarks](benchmarks.md) - Performance benchmarks
- [Optimization Tips](optimization-tips.md) - Performance optimization
- [Basic Usage](../guides/basic-usage.md) - SNID usage patterns

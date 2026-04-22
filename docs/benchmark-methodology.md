# SNID Benchmark Methodology

This document defines the rigorous benchmark methodology for SNID, based on 2025-2026 industry standards for ID type benchmarking. All SNID benchmarks should follow this methodology to ensure reproducible, statistically significant results.

## Core Principles

1. **Isolate the variable**: The only difference between test runs is the ID generation method. Same table schema, same hardware, same insert code, same data volume.
2. **Statistical rigor**: Multiple runs (10+), warmup phases, high-precision timers, report mean/p99/stddev.
3. **Realistic scale**: 1M–50M rows (10M is the sweet spot where fragmentation becomes visible).
4. **Reproducible environments**: Dockerized PostgreSQL (specific version), fixed config, SSD storage, scripted setups.
5. **Controlled concurrency**: Tests run both single-threaded and with realistic client counts (10–50 workers).

## 1. Generation Speed Benchmarks

### What They Measure
Time to create one ID (in nanoseconds/microseconds) under load.

### Methodology

**Setup:**
- Run 100,000+ iterations per test (10 full runs minimum)
- Warmup phase: 25,000+ iterations to JIT-compile and warm CPU caches
- High-precision timing: Go `testing.B`, Rust `criterion`, Python `time.perf_counter_ns()`
- Concurrency test: Multiple workers (10 threads) generating IDs simultaneously

**Libraries to Test:**
- SNID (Go, Rust, Python) - native mode
- SNID (Go, Rust, Python) - UUIDv7-compatible mode
- UUIDv7 (google/uuid, uuid crate, Python uuid)
- UUIDv4 (google/uuid, uuid crate, Python uuid)
- ULID (oklog/ulid, ulid crate, ulid-py)
- NanoID (nanoid, nanoid crate, python-nanoid)
- Snowflake (sonyflake, sonyflake crate)
- KSUID (segmentio/ksuid)

**Reporting:**
- Mean latency (ns/op)
- Throughput (ops/sec)
- Memory allocation (B/op, allocs/op)
- String length (chars)
- p95/p99 latency percentiles

**Example Command (Go):**
```bash
cd go
go test -bench=. -benchmem -run=^$ -benchtime=100000x
```

**Example Command (Rust):**
```bash
cd rust
cargo bench --bench benchmarks
```

**Example Command (Python):**
```bash
cd python
python -m pytest tests/test_bench.py --benchmark-only
```

## 2. Database Insert Throughput Benchmarks

### What They Measure
Rows inserted per second (or total time for X million rows).

### Methodology

**Database Setup:**
- PostgreSQL 17 or 18 (native uuidv7 support)
- Dockerized environment with fixed config:
  ```yaml
  shared_buffers: 256MB
  effective_cache_size: 1GB
  max_connections: 100
  wal_buffers: 16MB
  ```
- SSD storage (to eliminate disk I/O as bottleneck)
- Disable autovacuum for pure isolation (or use consistent autovacuum settings)

**Table Schema:**
```sql
-- SNID (binary)
CREATE TABLE snid_binary (
    id BYTEA PRIMARY KEY,
    data TEXT
);

-- SNID (UUID type)
CREATE TABLE snid_uuid (
    id UUID PRIMARY KEY,
    data TEXT
);

-- UUIDv7 (baseline)
CREATE TABLE uuidv7 (
    id UUID PRIMARY KEY,
    data TEXT
);

-- UUIDv4 (baseline)
CREATE TABLE uuidv4 (
    id UUID PRIMARY KEY,
    data TEXT
);

-- ULID (text)
CREATE TABLE ulid (
    id TEXT PRIMARY KEY,
    data TEXT
);
```

**Insert Method:**
- Pre-generate IDs in application (Go, Python, Rust scripts)
- Batched inserts: 1,000–10,000 rows per transaction
- Use prepared statements
- Measure wall-clock time for 10M rows
- Run with and without concurrent clients (10 workers)
- Vacuum/analyze between runs

**Example Insert Script (Go):**
```go
func benchmarkInsert(db *sql.DB, idType string, rows int) time.Duration {
    start := time.Now()
    
    tx, _ := db.Begin()
    stmt, _ := tx.Prepare("INSERT INTO " + idType + " (id, data) VALUES ($1, $2)")
    
    for i := 0; i < rows; i++ {
        var id string
        switch idType {
        case "snid_binary":
            id = snid.NewFast().ToHex()
        case "uuidv7":
            id = uuid.NewV7().String()
        // ... other types
        }
        stmt.Exec(id, "test data")
    }
    
    tx.Commit()
    return time.Since(start)
}
```

**Reporting:**
- Total time for 10M rows
- Rows per second
- Throughput gain vs UUIDv4 (baseline)
- Index size after inserts
- WAL size generated

## 3. Index & Table Size / Bloat Benchmarks

### What They Measure
Disk usage and fragmentation.

### Methodology

**After Insert Benchmarks:**
```sql
-- Table size
SELECT pg_relation_size('snid_binary') as table_size;

-- Index size
SELECT pg_indexes_size('snid_binary') as index_size;

-- Total size
SELECT pg_total_relation_size('snid_binary') as total_size;

-- Leaf page density (requires pgstattuple extension)
SELECT * FROM pgstattuple('snid_binary');
```

**Metrics to Report:**
- Table size (bytes)
- Index size (bytes)
- Total size (bytes)
- Leaf page density (%)
- Fragmentation level
- Size savings vs UUIDv4 (%)

**Expected Results (based on research):**
- UUIDv7/SNID: 22-27% smaller indexes than UUIDv4
- Leaf page density: ~92% vs ~68% for UUIDv4

## 4. Query Performance Benchmarks

### What They Measure
Point lookups and range scans.

### Methodology

**After Table Population (10M rows):**

**Point Lookup:**
```sql
-- Warm cache
SELECT * FROM snid_binary WHERE id = $1 LIMIT 1;

-- Benchmark (10,000 iterations)
EXPLAIN ANALYZE SELECT * FROM snid_binary WHERE id = $1;
```

**Range Scan / Recent Data:**
```sql
-- Recent 1000 rows
EXPLAIN ANALYZE 
SELECT * FROM snid_binary 
WHERE id > $1 
ORDER BY id 
LIMIT 1000;
```

**Metrics to Report:**
- Mean latency (ms)
- p99 latency (ms)
- Cache hit rate
- Index usage (via EXPLAIN ANALYZE)
- I/O operations

**Expected Results (based on research):**
- Point lookups: 2-4× faster than UUIDv4
- Range scans: ~3× faster than UUIDv4

## 5. Network / API Overhead Benchmarks

### What They Measure
JSON payload size and bandwidth impact.

### Methodology

**Test Setup:**
- Create sample API responses with different ID formats
- Measure JSON payload size
- Calculate bandwidth at various request rates (1k, 10k, 100k req/sec)

**ID Formats to Test:**
- SNID (22 chars Base58)
- UUIDv7 (36 chars hex)
- UUIDv4 (36 chars hex)
- ULID (26 chars Base32)
- NanoID (21 chars URL-safe)

**Reporting:**
- String length (chars)
- JSON payload size (bytes)
- Bandwidth at 10k req/sec (GB/day)
- Cost savings vs UUIDv7 (%)

## Reproducibility Checklist

Before publishing benchmark results, ensure:

- [ ] PostgreSQL version specified (17 or 18)
- [ ] Hardware specs documented (CPU, RAM, storage)
- [ ] Docker image/commit hash recorded
- [ ] Configuration settings (shared_buffers, etc.)
- [ ] Number of runs (minimum 10)
- [ ] Warmup iterations documented
- [ ] Statistical measures reported (mean, p99, stddev)
- [ ] Scripts open-sourced or provided
- [ ] Results reproducible by third party

## Common Tools

**Languages:**
- Go (most common for clean scripts)
- Python
- Rust

**Databases:**
- PostgreSQL 17/18 (primary)
- MySQL 8.0+ (secondary validation)

**Benchmarking Tools:**
- Go: `testing.B`, `benchstat`
- Rust: `criterion`
- Python: `pytest-benchmark`
- PostgreSQL: `pgbench`, custom scripts

**Analysis Tools:**
- PostgreSQL: `EXPLAIN ANALYZE`, `pgstattuple`
- System: `perf`, `iostat`, `vmstat`

## Limitations

- Results are PostgreSQL-centric (MySQL may show different patterns)
- Small tables (<1M rows) often show no difference — wins appear at scale
- Real production adds factors like WAL pressure, replication, vacuuming
- String storage adds minor overhead vs binary

## SNID-Specific Considerations

**Dual-Mode Testing:**
- Test SNID in native mode (fastest generation)
- Test SNID in UUIDv7-compatible mode (for DB performance comparison)
- Compare both against UUIDv7 baseline

**Extended Families:**
- Benchmark SGID (spatial) separately if geospatial queries are relevant
- Benchmark NID (neural) separately if ML pipeline performance matters
- Benchmark LID/KID separately if verification overhead is a concern

**Polyglot Conformance:**
- Run identical benchmarks in Go, Rust, and Python
- Verify byte-identical behavior across languages
- Report per-language performance differences

## Next Steps

1. Implement the database benchmark suite in `benchmarks/db/`
2. Create Docker Compose setup for reproducible PostgreSQL environment
3. Add benchmark results to `benchmarks/results/`
4. Update documentation with verified database performance metrics
5. Add continuous benchmark runs to CI pipeline

## References

- 2025-2026 independent ID benchmark studies (PostgreSQL-focused)
- RFC 9562 (UUID) specification
- PostgreSQL 17/18 documentation (native uuidv7 support)
- Industry benchmark best practices (Google Cloud, AWS, Azure)

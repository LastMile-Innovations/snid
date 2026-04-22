# SNID Database Benchmark Suite

This directory contains database-level benchmarks for SNID to verify theoretical performance claims (insert throughput, index size, query performance) following the methodology defined in `docs/benchmark-methodology.md`.

## Setup

### Prerequisites

- Docker and Docker Compose
- Go 1.24+
- Python 3.10+
- PostgreSQL client tools (psql)

### Quick Start

```bash
# Start PostgreSQL 18 with benchmark configuration
docker-compose up -d

# Wait for PostgreSQL to be ready
docker-compose exec postgres pg_isready

# Run all benchmarks
go run main.go --all

# Or run specific benchmark
go run main.go --benchmark insert-throughput
```

## Benchmark Suite

### 1. Insert Throughput Benchmark

**Purpose:** Measure rows inserted per second for different ID types.

**Scale:** 10 million rows (sweet spot for fragmentation visibility).

**ID Types Tested:**
- SNID (binary, 16 bytes)
- SNID (UUID type, 16 bytes)
- SNID (UUIDv7-compatible mode)
- UUIDv7 (baseline)
- UUIDv4 (baseline)
- ULID (text, 26 chars)
- Sequential BIGINT (baseline best)

**Metrics:**
- Total time for 10M rows
- Rows per second
- Throughput gain vs UUIDv4
- WAL size generated

**Run:**
```bash
go run main.go --benchmark insert-throughput --rows 10000000
```

### 2. Index Size Benchmark

**Purpose:** Measure disk usage and fragmentation after inserts.

**Scale:** 10 million rows (post-insert).

**ID Types Tested:** Same as insert throughput.

**Metrics:**
- Table size (bytes)
- Index size (bytes)
- Total size (bytes)
- Leaf page density (%)
- Fragmentation level
- Size savings vs UUIDv4 (%)

**Run:**
```bash
go run main.go --benchmark index-size
```

### 3. Query Performance Benchmark

**Purpose:** Measure point lookups and range scans.

**Scale:** 10 million rows (post-insert, post-vacuum).

**ID Types Tested:** Same as insert throughput.

**Metrics:**
- Mean latency (ms)
- p99 latency (ms)
- Cache hit rate
- Index usage (EXPLAIN ANALYZE)
- I/O operations

**Run:**
```bash
go run main.go --benchmark query-performance
```

### 4. Network Overhead Benchmark

**Purpose:** Measure JSON payload size and bandwidth impact.

**ID Types Tested:**
- SNID (22 chars Base58)
- UUIDv7 (36 chars hex)
- UUIDv4 (36 chars hex)
- ULID (26 chars Base32)
- NanoID (21 chars URL-safe)

**Metrics:**
- String length (chars)
- JSON payload size (bytes)
- Bandwidth at 10k req/sec (GB/day)
- Cost savings vs UUIDv7 (%)

**Run:**
```bash
go run main.go --benchmark network-overhead
```

## Docker Configuration

**PostgreSQL 18 Configuration:**
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:18
    environment:
      POSTGRES_PASSWORD: benchmark
      POSTGRES_DB: snid_bench
    command:
      - "shared_buffers=256MB"
      - "effective_cache_size=1GB"
      - "max_connections=100"
      - "wal_buffers=16MB"
      - "random_page_cost=1.1"  # SSD optimization
    volumes:
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"
```

## Schema

```sql
-- SNID (binary storage)
CREATE TABLE snid_binary (
    id BYTEA PRIMARY KEY,
    data TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- SNID (UUID type storage)
CREATE TABLE snid_uuid (
    id UUID PRIMARY KEY,
    data TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- UUIDv7 (baseline)
CREATE TABLE uuidv7 (
    id UUID PRIMARY KEY,
    data TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- UUIDv4 (baseline)
CREATE TABLE uuidv4 (
    id UUID PRIMARY KEY,
    data TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- ULID (text storage)
CREATE TABLE ulid (
    id TEXT PRIMARY KEY,
    data TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Sequential BIGINT (baseline best)
CREATE TABLE sequential (
    id BIGSERIAL PRIMARY KEY,
    data TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## Expected Results (Based on Research)

**Insert Throughput:**
- SNID/UUIDv7: +35-50% faster than UUIDv4
- Sequential BIGINT: +100-200% faster than UUIDv4 (baseline best)

**Index Size:**
- SNID/UUIDv7: 22-27% smaller than UUIDv4
- Leaf page density: ~92% vs ~68% for UUIDv4

**Query Performance:**
- Point lookups: 2-4× faster than UUIDv4
- Range scans: ~3× faster than UUIDv4

## Results

Results are saved to `results/` with timestamped filenames:

```
results/
├── insert-throughput-20260422-120000.json
├── index-size-20260422-120500.json
├── query-performance-20260422-121000.json
└── network-overhead-20260422-121500.json
```

## Continuous Integration

Add to CI pipeline:

```yaml
- name: Run Database Benchmarks
  run: |
    cd benchmarks/db
    docker-compose up -d
    docker-compose exec postgres pg_isready
    go run main.go --all
    docker-compose down
```

## Troubleshooting

**PostgreSQL not ready:**
```bash
docker-compose logs postgres
docker-compose exec postgres pg_isready
```

**Insufficient disk space:**
```bash
docker system prune -a
```

**Benchmark too slow:**
- Reduce row count for testing: `--rows 1000000`
- Check PostgreSQL configuration
- Verify SSD storage is being used

## Contributing

When adding new benchmarks:

1. Follow the methodology in `docs/benchmark-methodology.md`
2. Add the benchmark to this README
3. Update the schema if needed
4. Include expected results based on research
5. Run benchmarks 10+ times for statistical significance
6. Report mean, p99, and stddev

## References

- [Benchmark Methodology](../../docs/benchmark-methodology.md)
- [Performance Comparison](../../docs/performance/comparison.md)
- [RFC 9562 (UUID)](https://www.rfc-editor.org/rfc/rfc9562)

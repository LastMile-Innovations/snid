# Storage Contracts Guide

Database integration patterns for SNID identifiers.

## Overview

SNID provides canonical binary storage contracts for efficient database storage and retrieval.

## Binary Storage

### Recommended Storage Types

| Engine | Storage Type | Guidance |
|--------|--------------|----------|
| PostgreSQL | `UUID` or `BYTEA` | Prefer raw 16-byte or 32-byte binds |
| ClickHouse | `FixedString(16)` or `FixedString(32)` | Preserve lexicographic ordering |
| MySQL | `BINARY(16)` or `BINARY(32)` | Raw binary storage |
| SQLite | `BLOB` | Raw binary storage |
| Neo4j | `byte[]` | Wire strings are debug-only |
| Redis/Dragonfly | raw bytes or wire string | Prefer bytes for hot-path keys |

## PostgreSQL

### Using UUID Type

```sql
CREATE TABLE items (
    id UUID PRIMARY KEY,
    name TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Insert
INSERT INTO items (id, name) VALUES ($1, $2);

-- Query
SELECT * FROM items WHERE id = $1;
```

**Go:**
```go
import (
    "database/sql"
    "github.com/neighbor/snid"
)

func insertItem(db *sql.DB, id snid.ID, name string) error {
    _, err := db.Exec(
        "INSERT INTO items (id, name) VALUES ($1, $2)",
        id.UUID(), name,
    )
    return err
}
```

**Python:**
```python
import snid
import psycopg2

id = snid.SNID.new_fast()
cursor.execute(
    "INSERT INTO items (id, name) VALUES (%s, %s)",
    (id.to_uuid(), "item name")
)
```

### Using BYTEA Type (Recommended)

```sql
CREATE TABLE items (
    id BYTEA PRIMARY KEY,
    name TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Insert
INSERT INTO items (id, name) VALUES ($1, $2);

-- Query
SELECT * FROM items WHERE id = $1;
```

**Go:**
```go
func insertItem(db *sql.DB, id snid.ID, name string) error {
    _, err := db.Exec(
        "INSERT INTO items (id, name) VALUES ($1, $2)",
        id[:], name,
    )
    return err
}
```

**Python:**
```python
cursor.execute(
    "INSERT INTO items (id, name) VALUES (%s, %s)",
    (id.to_bytes(), "item name")
)
```

## ClickHouse

### Using FixedString

```sql
CREATE TABLE items (
    id FixedString(16) PRIMARY KEY,
    name String,
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY id;

-- Insert
INSERT INTO items (id, name) VALUES ($1, $2);

-- Query
SELECT * FROM items WHERE id = $1;
```

**Go:**
```go
func insertItem(db *sql.DB, id snid.ID, name string) error {
    _, err := db.Exec(
        "INSERT INTO items (id, name) VALUES (?, ?)",
        string(id[:]), name,
    )
    return err
}
```

## MySQL

### Using BINARY

```sql
CREATE TABLE items (
    id BINARY(16) PRIMARY KEY,
    name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert
INSERT INTO items (id, name) VALUES (?, ?);

-- Query
SELECT * FROM items WHERE id = ?;
```

**Go:**
```go
func insertItem(db *sql.DB, id snid.ID, name string) error {
    _, err := db.Exec(
        "INSERT INTO items (id, name) VALUES (?, ?)",
        id[:], name,
    )
    return err
}
```

## Neo4j

### Using Binary Storage

```cypher
// Create node with binary ID
CREATE (n:Item {id: $binary_id, name: $name})

// Query by ID
MATCH (n:Item {id: $binary_id})
RETURN n
```

**Go:**
```go
import "github.com/neighbor/snid/neo4j"

func createItem(driver neo4j.Driver, id snid.ID, name string) error {
    session := driver.NewSession(neo4j.SessionConfig{})
    defer session.Close()

    neo4jID := neo4j.ToNeo4jID(id)
    _, err := session.Run(
        "CREATE (n:Item {id: $id, name: $name})",
        map[string]interface{}{
            "id":   neo4jID,
            "name": name,
        },
    )
    return err
}
```

**Python:**
```python
from snid import neo4j

neo4j_id = neo4j.to_neo4j_id(id)
session.run(
    "CREATE (n:Item {id: $id, name: $name})",
    id=neo4j_id, name="item name"
)
```

## Redis

### Using Raw Bytes

```python
import snid
import redis

r = redis.Redis()
id = snid.SNID.new_fast()

# Set
r.set(id.to_bytes(), b'value')

# Get
value = r.get(id.to_bytes())
```

### Using Wire String

```python
# Set
r.set(id.to_wire("MAT"), b'value')

# Get
value = r.get(id.to_wire("MAT"))
```

## SQLite

### Using BLOB

```sql
CREATE TABLE items (
    id BLOB PRIMARY KEY,
    name TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert
INSERT INTO items (id, name) VALUES (?, ?);

-- Query
SELECT * FROM items WHERE id = ?;
```

**Go:**
```go
func insertItem(db *sql.DB, id snid.ID, name string) error {
    _, err := db.Exec(
        "INSERT INTO items (id, name) VALUES (?, ?)",
        id[:], name,
    )
    return err
}
```

## Extended ID Families

### 32-byte IDs (NID, LID, WID, XID, KID)

For 32-byte identifier families, use 32-byte storage types:

**PostgreSQL:**
```sql
CREATE TABLE neural_items (
    id BYTEA PRIMARY KEY,  -- 32 bytes
    name TEXT
);
```

**ClickHouse:**
```sql
CREATE TABLE neural_items (
    id FixedString(32) PRIMARY KEY,
    name String
) ENGINE = MergeTree()
ORDER BY id;
```

## Indexing

### B-Tree Index (Recommended)

```sql
CREATE INDEX idx_items_id ON items (id);
```

### Hash Index (For equality only)

```sql
CREATE INDEX idx_items_id_hash ON items USING HASH (id);
```

## Migration from UUID

If migrating from UUID to SNID:

```sql
-- Add new column
ALTER TABLE items ADD COLUMN snid_id BYTEA;

-- Migrate data
UPDATE items SET snid_id = id::bytea;

-- Drop old column
ALTER TABLE items DROP COLUMN id;

-- Rename
ALTER TABLE items RENAME COLUMN snid_id TO id;
```

## Best Practices

1. **Prefer binary storage** over wire strings for production
2. **Use appropriate storage type** for your database engine
3. **Index ID columns** for query performance
4. **Use wire strings** only for debugging and logging
5. **Batch insert** for high-throughput scenarios

## Next Steps

- [Basic Usage](basic-usage.md) - Common patterns
- [Batch Generation](batch-generation.md) - High-throughput patterns
- [Integration Contracts](../INTEGRATION_CONTRACTS.md) - Detailed contracts

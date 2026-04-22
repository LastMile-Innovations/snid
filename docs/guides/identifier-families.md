# Identifier Families

SNID provides multiple identifier families for different use cases.

## Core SNID

The core 128-bit time-ordered identifier with UUID v7-compatible ordering.

**Use cases:**
- General-purpose unique identifiers
- Primary keys in databases
- Event sourcing
- Distributed systems

**Byte layout:**
- Bits 0-47: Unix timestamp (ms)
- Bits 48-51: Version nibble (0b0111)
- Bits 52-65: Monotonic sequence (14 bits)
- Bits 66-89: Machine/process fingerprint (24 bits)
- Bits 90-127: Entropy tail

## SGID (Spatial ID)

Spatial identifiers with H3 geospatial encoding for location-aware applications.

**Use cases:**
- Building and location tracking
- Static sensor networks
- Geospatial indexing
- Location-based services

**Byte layout:**
- High 64 bits: H3 cell encoding
- Version nibble: 0b1000 (v8)
- Preserves H3 locality for lexicographic scans

**Go:**
```go
sgid := snid.NewSpatial(37.7749, -122.4194) // San Francisco
sgid := snid.NewSpatialPrecise(37.7749, -122.4194, 12) // Resolution 12
```

**Rust:**
```rust
let sgid = SGID::from_spatial_parts(h3_cell, entropy);
```

**Python:**
```python
sgid = snid.SGID.from_spatial_parts(h3_cell, entropy)
```

## NID (Neural ID)

Neural identifiers with semantic tail for vector search and ML applications.

**Use cases:**
- Vector database keys
- Semantic search
- ML pipeline identifiers
- Embedding-based lookups

**Byte layout:**
- 16-byte SNID head
- 16-byte semantic tail (e.g., LSH, vector quantization)

**Go:**
```go
nid, err := snid.NewNeural(base_snid, semantic_hash)
```

**Rust:**
```rust
let nid = NID::from_parts(base_snid, semantic_hash);
```

**Python:**
```python
nid = snid.NID.from_parts(base_snid, semantic_hash)
```

## LID (Ledger ID)

Ledger identifiers with HMAC verification tail for immutable logs.

**Use cases:**
- Blockchain and distributed ledgers
- Immutable audit trails
- Cryptographic verification
- Tamper-evident logging

**Byte layout:**
- 16-byte SNID head
- 16-byte HMAC-SHA256 verification tail

**Go:**
```go
lid, err := snid.NewLID(prev, payload, key)
```

**Rust:**
```rust
let lid = LID::from_parts(head, prev, payload, key)?;
```

**Python:**
```python
lid = snid.LID.from_parts(head, prev, payload, key)
```

## WID (World ID)

World/scenario identifiers for simulation isolation.

**Use cases:**
- Simulation environments
- Scenario isolation
- World-state tracking
- Multi-tenant simulation

**Byte layout:**
- 16-byte SNID head
- 16-byte scenario/world hash

**Go:**
```go
wid := snid.NewWID(head_snid, scenario_hash)
```

**Rust:**
```rust
let wid = WID::from_parts(head, scenario_hash);
```

**Python:**
```python
wid = snid.WID.from_parts(head, scenario_hash)
```

## XID (Edge ID)

Edge identifiers for relationship identity and bitemporal auditing.

**Use cases:**
- Graph edge identifiers
- Relationship tracking
- Bitemporal edge auditing
- Graph database keys

**Byte layout:**
- 16-byte SNID head
- 16-byte edge hash

**Go:**
```go
xid := snid.NewXID(head_snid, edge_hash)
```

**Rust:**
```rust
let xid = XID::from_parts(head, edge_hash);
```

**Python:**
```python
xid = snid.XID.from_parts(head, edge_hash)
```

## KID (Capability ID)

Capability identifiers with MAC-based verification for authorization.

**Use cases:**
- Self-verifying capability grants
- Authorization tokens
- Edge cache validation
- Binary storage authorization

**Byte layout:**
- 16-byte SNID head
- 16-byte MAC tail (binds actor, resource, capability)

**Go:**
```go
kid, err := snid.NewKIDForCapability(head, actor, resource, capability, key)
```

**Rust:**
```rust
let kid = KID::for_capability(head, actor, resource, capability, key)?;
```

**Python:**
```python
kid = snid.KID.for_capability(head, actor, resource, capability, key)
```

## EID (Ephemeral ID)

64-bit ephemeral identifiers for short-lived sessions.

**Use cases:**
- Session identifiers
- Temporary contexts
- Short-lived operations
- Rate limiting keys

**Byte layout:**
- 48-bit unix milliseconds
- 16-bit session/counter field

**Go:**
```go
eid := snid.NewEphemeral(session)
```

**Rust:**
```rust
let eid = EID::from_parts(unix_millis, counter);
```

**Python:**
```python
eid = snid.EID.from_parts(unix_millis, counter)
```

## BID (Content-Addressable ID)

Content-addressable identifiers for CAS systems.

**Use cases:**
- Content-addressable storage
- Deduplication
- Immutable content tracking
- CAS systems

**Byte layout:**
- Topology: 16-byte SNID
- Content: 32-byte BLAKE3 hash

**Go:**
```go
bid := snid.NewBID(content_hash)
```

**Rust:**
```rust
let bid = BID::from_hash(content_hash)?;
```

**Python:**
```python
bid = snid.BID.from_hash(content_hash)
```

## AKID (Access Key ID)

Dual-part public-plus-secret credentials.

**Use cases:**
- API keys
- Access tokens
- Dual-part credentials
- Secret management

**Wire format:**
```
KEY:<public_snid>_<opaque_secret>
```

**Go:**
```go
public := snid.NewAKIDPublic(tenantID)
secret, err := snid.NewAKIDSecret()
```

**Rust:**
```rust
let public = AKID::public(tenant_id);
let secret = AKID::secret()?;
```

**Python:**
```python
public = snid.AKID.public(tenant_id)
secret = snid.AKID.secret()
```

## Choosing the Right Family

| Use Case | Recommended Family |
|----------|-------------------|
| General purpose | SNID |
| Location tracking | SGID |
| Vector search | NID |
| Immutable logs | LID |
| Simulation | WID |
| Graph edges | XID |
| Authorization | KID |
| Sessions | EID |
| Content storage | BID |
| API keys | AKID |

## Next Steps

- [Wire Format](wire-format.md) - Canonical wire string format
- [Boundary Projections](boundary-projections.md) - Tensor and storage projections
- [Spatial IDs](spatial-ids.md) - Detailed SGID guide
- [Neural IDs](neural-ids.md) - Detailed NID guide

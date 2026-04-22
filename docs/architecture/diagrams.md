# Architecture Diagrams

Visual representations of SNID architecture and components.

## System Architecture

```mermaid
graph TB
    subgraph "SNID Protocol"
        SPEC[docs/SPEC.md<br/>Canonical Specification]
    end

    subgraph "Implementations"
        GO[go/<br/>Reference Implementation]
        RUST[rust/<br/>Deterministic Core]
        PYTHON[python/<br/>PyO3 Bindings]
    end

    subgraph "Conformance"
        VECTORS[conformance/vectors.json<br/>Test Vectors]
        GEN[conformance/cmd/generate_vectors<br/>Vector Generator]
    end

    SPEC --> GO
    SPEC --> RUST
    SPEC --> PYTHON
    GO --> GEN
    GEN --> VECTORS
    VECTORS --> RUST
    VECTORS --> PYTHON
```

## ID Generation Flow

```mermaid
graph LR
    A[Request ID] --> B{Generator}
    B -->|NewFast| C[Lock-free per-P state]
    B -->|NewProjected| D[Tenant + Shard]
    B -->|NewBatch| E[Batch allocator]
    C --> F[Coarse Clock]
    D --> F
    E --> F
    F --> G[Monotonic Sequence]
    G --> H[Machine Fingerprint]
    H --> I[Entropy Tail]
    I --> J[128-bit SNID]
```

## Conformance Testing Flow

```mermaid
graph TB
    A[Go Implementation] --> B[Generate Vectors]
    B --> C[vectors.json]
    C --> D[Rust Validation]
    C --> E[Python Validation]
    D --> F{Pass?}
    E --> G{Pass?}
    F -->|Yes| H[✅ Conformance Pass]
    F -->|No| I[❌ Conformance Fail]
    G -->|Yes| H
    G -->|No| I
```

## Byte Layout

```mermaid
graph LR
    A[SNID 128 bits] --> B[Bits 0-47<br/>Unix ms]
    A --> C[Bits 48-51<br/>Version]
    A --> D[Bits 52-65<br/>Sequence]
    A --> E[Bits 66-89<br/>Machine]
    A --> F[Bits 90-127<br/>Entropy]
```

## Wire Format Encoding

```mermaid
graph LR
    A[128-bit SNID] --> B[Base58 Encode]
    B --> C[CRC8 Checksum]
    C --> D[Payload]
    D --> E[Add Atom Prefix]
    E --> F[Wire String<br/>ATOM:payload]
```

## Extended ID Families

```mermaid
graph TB
    subgraph "16-byte IDs"
        SNID[SNID<br/>Core]
        SGID[SGID<br/>Spatial]
        EID[EID<br/>Ephemeral]
    end

    subgraph "32-byte IDs"
        NID[NID<br/>Neural]
        LID[LID<br/>Ledger]
        WID[WID<br/>World]
        XID[XID<br/>Edge]
        KID[KID<br/>Capability]
        BID[BID<br/>Content]
    end

    subgraph "Dual-Part"
        AKID[AKID<br/>Public + Secret]
    end
```

## Tensor Projections

```mermaid
graph LR
    A[SNID] --> B[Tensor128]
    B --> C[hi: int64]
    B --> D[lo: int64]
    
    E[NID] --> F[Tensor256]
    F --> G[w0: int64]
    F --> H[w1: int64]
    F --> I[w2: int64]
    F --> J[w3: int64]
```

## Database Storage Contracts

```mermaid
graph TB
    A[SNID] --> B{Storage Engine}
    B -->|PostgreSQL| C[UUID or BYTEA]
    B -->|ClickHouse| D[FixedString 16]
    B -->|MySQL| E[BINARY 16]
    B -->|SQLite| F[BLOB]
    B -->|Neo4j| G[byte[]]
    B -->|Redis| H[Raw bytes]
```

## Batch Generation (Python)

```mermaid
graph LR
    A[generate_batch] --> B{Backend}
    B -->|bytes| C[Raw bytes<br/>Fastest]
    B -->|tensor| D[Tensor pairs<br/>Fast]
    B -->|numpy| E[NumPy array<br/>Zero-copy]
    B -->|pyarrow| F[PyArrow array<br/>Medium]
    B -->|polars| G[Polars series<br/>Medium]
    B -->|snid| H[Python objects<br/>Slow]
```

## AI/ML Pipeline Integration

```mermaid
graph TB
    A[Embedding] --> B[Semantic Hash]
    B --> C[NID Generation]
    C --> D[Vector Database]
    D --> E[Semantic Search]
    
    F[Batch Generation] --> G[NumPy Backend]
    G --> H[Training Data]
    H --> I[ML Model]
```

## Spatial ID (SGID) Flow

```mermaid
graph LR
    A[Lat/Lng] --> B[H3 Encoding]
    B --> C[H3 Cell]
    C --> D[SGID Generation]
    D --> E[Wire Format]
    E --> F[LOC:payload]
```

## Development Workflow

```mermaid
graph TB
    A[Code Change] --> B[just fmt]
    B --> C[just lint]
    C --> D[just test]
    D --> E[just conformance]
    E --> F{Pass?}
    F -->|Yes| G[Commit]
    F -->|No| H[Fix]
    H --> A
```

## Release Process

```mermaid
graph TB
    A[Version Bump] --> B[go.mod]
    A --> C[Cargo.toml]
    A --> D[pyproject.toml]
    A --> E[CHANGELOG.md]
    B --> F[Generate Vectors]
    C --> F
    D --> F
    F --> G[Validate All]
    G --> H{Pass?}
    H -->|Yes| I[Publish All]
    H -->|No| J[Fix]
    J --> A
```

## CLI Architecture (Planned)

```mermaid
graph TB
    A[snid CLI] --> B[generate]
    A --> C[validate]
    A --> D[project]
    A --> E[benchmark]
    
    B --> F[--type]
    B --> G[--count]
    
    C --> H[--conformance]
    
    D --> I[--topology]
    D --> J[--cell]
    
    E --> K[--compare]
```

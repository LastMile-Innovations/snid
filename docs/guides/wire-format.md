# Wire Format Guide

Canonical wire string format for SNID identifiers.

## Overview

The canonical wire format is:

```
<ATOM>:<payload>
```

Where:
- `ATOM` is an uppercase canonical atom (type tag)
- `payload` is Base58 encoding of the 16-byte SNID with CRC8 checksum
- `_` is accepted as a compatibility delimiter but never canonical output

## Canonical Atoms

| Atom | Meaning | Use Case |
|------|---------|----------|
| IAM | Identity | Users, personas |
| TEN | Tenant | Multi-tenancy |
| MAT | Matter | Objects, items |
| LOC | Location | Spatial entities |
| CHR | Character | Schema, taxonomy |
| LED | Ledger | Transactions |
| LEG | Legal | Legal entities |
| TRU | Trust | Network, trust |
| KIN | Kinetic | Actions, events |
| COG | Cognition | AI, reasoning |
| SEM | Semantic | Knowledge |
| SYS | System | System entities |
| EVT | Events | Events |
| SES | Session | Sessions |
| KEY | Keys | Credentials |

## Legacy Atoms

These atoms are accepted at parse time and normalized:

| Legacy | Canonical |
|--------|-----------|
| OBJ | MAT |
| TXN | LED |
| SCH | CHR |
| NET | TRU |
| OPS | EVT |
| ACT | IAM |
| GRP | TEN |
| BIO | IAM |
| ATM | LOC |

## Encoding

### Base58 Alphabet

```
123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz
```

### Checksum

CRC8-derived check digit appended to Base58 payload for error detection.

## Examples

### Encoding

**Go:**
```go
id := snid.NewFast()
wire := id.String(snid.Matter)
// MAT:2xXFhP9w7V4sKjBnG8mQpL
```

**Rust:**
```rust
let id = SNID::new();
let wire = id.to_wire("MAT");
// MAT:2xXFhP9w7V4sKjBnG8mQpL
```

**Python:**
```python
id = snid.SNID.new_fast()
wire = id.to_wire("MAT")
# MAT:2xXFhP9w7V4sKjBnG8mQpL
```

### Parsing

**Go:**
```go
parsed, atom, err := snid.FromString("MAT:2xXFhP9w7V4sKjBnG8mQpL")
```

**Rust:**
```rust
let (parsed, atom) = SNID::parse_wire("MAT:2xXFhP9w7V4sKjBnG8mQpL")?;
```

**Python:**
```python
parsed, atom = snid.SNID.parse_wire("MAT:2xXFhP9w7V4sKjBnG8mQpL")
```

## Compact Format

For cases where the atom is not needed:

**Go:**
```go
compact := id.StringCompact()
// 2xXFhP9w7V4sKjBnG8mQpL
```

**Rust:**
```rust
let compact = id.to_compact();
// 2xXFhP9w7V4sKjBnG8mQpL
```

**Python:**
```python
compact = id.to_compact()
# 2xXFhP9w7V4sKjBnG8mQpL
```

## Compatibility Delimiter

The underscore `_` is accepted as a compatibility delimiter:

```
MAT_2xXFhP9w7V4sKjBnG8mQpL  # Accepted (legacy)
MAT:2xXFhP9w7V4sKjBnG8mQpL   # Canonical
```

## Extended ID Families

### AKID (Dual-Part Credentials)

```
KEY:<public_snid>_<opaque_secret>
```

**Go:**
```go
public := snid.NewAKIDPublic(tenantID)
secret, _ := snid.NewAKIDSecret()
wire := fmt.Sprintf("KEY:%s_%s", public.String(snid.Key), secret)
```

**Python:**
```python
public = snid.AKID.public(tenant_id)
secret = snid.AKID.secret()
wire = f"KEY:{public.to_wire('KEY')}_{secret}"
```

### BID (Content-Addressable)

```
CAS:<snid_payload_base58>:<content_hash_base32_lower_no_padding>
```

**Go:**
```go
bid := snid.NewBID(contentHash)
wire := bid.String()
```

## Error Handling

### Invalid Characters

```go
_, _, err := snid.FromString("MAT:invalid!")
// Error: invalid character
```

### Checksum Mismatch

```go
_, _, err := snid.FromString("MAT:2xXFhP9w7V4sKjBnG8mQpX")
// Error: checksum mismatch
```

### Invalid Atom

```go
_, _, err := snid.FromString("XXX:2xXFhP9w7V4sKjBnG8mQpL")
// Error: invalid atom
```

## Best Practices

1. **Use canonical format** for storage and transmission
2. **Use compact format** only when atom is implied
3. **Validate wire strings** before parsing
4. **Use atoms consistently** for your domain
5. **Prefer binary storage** for databases

## Next Steps

- [Basic Usage](basic-usage.md) - Common patterns
- [Identifier Families](identifier-families.md) - All ID families
- [Storage Contracts](storage-contracts.md) - Database integration

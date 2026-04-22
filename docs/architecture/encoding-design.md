# Encoding Design

Base58 encoding and checksum implementation.

## Overview

SNID uses Base58 encoding with CRC8 checksum for wire format:
- Base58 alphabet (no ambiguous characters)
- CRC8-derived check digit for error detection
- Optimized 128-bit integer division for 16-byte SNID wire payloads
- Canonical Bitcoin-style Base58 for auxiliary byte payloads such as AKID secrets and short IDs
- Zero-allocation parse paths for SNID wire payloads

## Base58 Alphabet

```
123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz
```

Excludes: `0`, `O`, `I`, `l` (ambiguous characters)

This is the Bitcoin Base58 alphabet. It does not use `+` or `/`; the wire format is URL-safe aside from normal delimiter escaping rules.

## Encoding Process

### 16-byte SNID payload to Base58

```go
func (id ID) appendPayload(dst []byte) []byte {
    var buf [24]byte
    idx := 23

    buf[idx] = base58Alphabet[crc8(id[:])%58]
    idx--

    hi := binary.BigEndian.Uint64(id[:8])
    lo := binary.BigEndian.Uint64(id[8:])
    for hi > 0 || lo > 0 {
        qhi := hi / 58
        rhi := hi - qhi*58
        qlo, rem := bits.Div64(rhi, lo, 58)
        hi, lo = qhi, qlo
        buf[idx] = base58Alphabet[rem]
        idx--
    }

    for i := 0; i < 16 && id[i] == 0; i++ {
        buf[idx] = '1'
        idx--
    }

    return append(dst, buf[idx+1:]...)
}
```

The 16-byte SNID hot path avoids `big.Int` and external Base58 dependencies. The checksum character is a Base58 symbol selected from `crc8(id[:]) % 58`.

### Auxiliary bytes to Base58

AKID secrets and `ShortID` payloads use canonical big-endian Base58 conversion:

- leading zero bytes are preserved as leading `1` characters
- division uses `(carry << 8) | byte` across the significant input window
- stack scratch space covers current 8-byte and 24-byte call sites without external dependencies
- decode mirrors the same representation and preserves leading zero bytes exactly

### Checksum Calculation

```go
func checksum(data []byte) byte {
    var crc byte
    for _, b := range data {
        crc = crc8Table[crc^b]
    }
    return crc
}
```

## Decoding Process

### SNID payload to binary

```go
func decode16Base58(id *ID, src []byte) error {
    *id = Zero
    for _, c := range src {
        val := b58Map[c]
        if val == -1 {
            return errors.New("snid: invalid char")
        }
        carry := uint32(val)
        for i := 15; i >= 0; i-- {
            res := uint32(id[i])*58 + carry
            id[i] = byte(res)
            carry = res >> 8
        }
        if carry > 0 {
            return errors.New("snid: overflow")
        }
    }
    return nil
}
```

## Performance Optimizations

### 128-bit Integer Division

SNID wire encoding treats the 16-byte ID as a 128-bit big-endian integer split into `hi` and `lo` words. Each Base58 digit is emitted by dividing `(hi:lo)` by 58 with `bits.Div64`, which avoids heap allocation and avoids external Base58 libraries.

### Allocation behavior

- `AppendTo` can encode an atom-prefixed SNID wire string with zero allocations when the caller provides capacity.
- `String` and `StringCompact` allocate only the returned string.
- `FromString` and `ParseCompact` are zero-allocation in current Go benchmarks.

## Wire Format Construction

```
<ATOM>:<BASE58_PAYLOAD><CHECKSUM>
```

### Encoding

```go
func (id ID) String(atom Atom) string {
    var buf [48]byte
    n := copy(buf[:], string(atom))
    buf[n] = ':'
    return bytesToString(id.appendPayload(buf[:n+1]))
}
```

### Decoding

```go
func FromString(s string) (ID, Atom, error) {
    var id ID
    atom, err := id.Parse(s)
    return id, atom, err
}
```

## Error Handling

### Invalid Characters

```go
if val == -1 {
    return nil, errors.New("invalid character in Base58")
}
```

### Checksum Mismatch

```go
if payload[dataLen] != base58Alphabet[crc8(id[:])%58] {
    return nil, errors.New("checksum mismatch")
}
```

### Invalid Length

```go
if len(bytes) != 16 {
    return nil, errors.New("invalid length")
}
```

## Base32 Alternative

For systems that prefer Base32 (e.g., BID content hash):

```go
func encodeBase32(data []byte) string {
    // RFC 4648 Base32 encoding
    // Lowercase, no padding
}
```

## Implementation Details

### Go

See `go/encoding.go` for full implementation.

### Rust

See `rust/src/lib.rs` for encoding implementation.

### Python

Python bindings use Rust core for encoding.

## Security Considerations

- Base58 encoding is not encryption
- Checksum is for error detection, not security
- Wire strings are safe to expose in APIs
- AKID secrets require separate protection

## Next Steps

- [Generator Design](generator-design.md) - ID generation
- [Conformance Design](conformance-design.md) - Cross-language conformance
- [Wire Format](../guides/wire-format.md) - Wire format guide

# Encoding Design

Base58 encoding and checksum implementation.

## Overview

SNID uses Base58 encoding with CRC8 checksum for wire format:
- Base58 alphabet (no ambiguous characters)
- CRC8-derived check digit for error detection
- Optimized 128-bit integer division for encoding
- Unsafe string-byte conversions for performance

## Base58 Alphabet

```
123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz
```

Excludes: `0`, `O`, `I`, `l` (ambiguous characters)

## Encoding Process

### Binary to Base58

```go
func encodeBase58(data []byte) string {
    // 128-bit integer division
    x := new(big.Int).SetBytes(data)
    
    // Base58 conversion
    var result []byte
    base := big.NewInt(58)
    zero := big.NewInt(0)
    mod := new(big.Int)
    
    for x.Cmp(zero) > 0 {
        x.DivMod(x, base, mod)
        result = append(result, alphabet[mod.Int64()])
    }
    
    // Reverse result
    for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
        result[i], result[j] = result[j], result[i]
    }
    
    return string(result)
}
```

### Checksum Calculation

```go
func checksum(data []byte) byte {
    // CRC8-derived check digit
    var crc byte
    for _, b := range data {
        crc ^= b
        for i := 0; i < 8; i++ {
            if crc&0x80 != 0 {
                crc = (crc << 1) ^ 0x07
            } else {
                crc <<= 1
            }
        }
    }
    return crc
}
```

## Decoding Process

### Base58 to Binary

```go
func decodeBase58(s string) ([]byte, error) {
    x := big.NewInt(0)
    base := big.NewInt(58)
    
    for _, c := range s {
        idx := bytes.IndexByte(alphabet, byte(c))
        if idx == -1 {
            return nil, errors.New("invalid character")
        }
        x.Mul(x, base)
        x.Add(x, big.NewInt(int64(idx)))
    }
    
    return x.Bytes(), nil
}
```

### Checksum Validation

```go
func validateChecksum(data []byte, expected byte) bool {
    return checksum(data) == expected
}
```

## Performance Optimizations

### 128-bit Integer Division

Optimized division for 128-bit values:

```go
func encodeUint128(hi, lo uint64) string {
    // Optimized division for 128-bit values
    // Avoids big.Int overhead
}
```

### Unsafe String-Byte Conversion

Zero-copy string-byte conversion:

```go
func bytesToString(b []byte) string {
    return *(*string)(unsafe.Pointer(&b))
}

func stringToBytes(s string) []byte {
    return *(*[]byte)(unsafe.Pointer(&s))
}
```

## Wire Format Construction

```
<ATOM>:<BASE58_PAYLOAD><CHECKSUM>
```

### Encoding

```go
func (id ID) String(atom Atom) string {
    // Encode to Base58
    payload := encodeBase58(id[:])
    
    // Calculate checksum
    ck := checksum(id[:])
    
    // Append checksum
    payload += string(alphabet[ck])
    
    // Add atom prefix
    return atom.String() + ":" + payload
}
```

### Decoding

```go
func FromString(s string) (ID, Atom, error) {
    // Split atom and payload
    parts := strings.Split(s, ":")
    if len(parts) != 2 {
        return ID{}, "", errors.New("invalid format")
    }
    
    atom := ParseAtom(parts[0])
    payload := parts[1]
    
    // Separate checksum
    if len(payload) < 1 {
        return ID{}, "", errors.New("invalid payload")
    }
    
    data := payload[:len(payload)-1]
    ck := payload[len(payload)-1]
    
    // Decode Base58
    bytes, err := decodeBase58(data)
    if err != nil {
        return ID{}, "", err
    }
    
    // Validate checksum
    expected := checksum(bytes)
    if alphabet[expected] != ck {
        return ID{}, "", errors.New("checksum mismatch")
    }
    
    // Convert to ID
    var id ID
    copy(id[:], bytes)
    
    return id, atom, nil
}
```

## Error Handling

### Invalid Characters

```go
if idx == -1 {
    return nil, errors.New("invalid character in Base58")
}
```

### Checksum Mismatch

```go
if checksum(bytes) != expected {
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

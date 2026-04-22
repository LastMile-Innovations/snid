# Common Errors

Common error messages and their solutions.

## Parse Errors

### Invalid wire string format

**Error:**
```
invalid wire string format
```

**Cause:** Wire string doesn't match `<ATOM>:<payload>` format.

**Solution:**
```go
// Correct
wire := "MAT:2xXFhP9w7V4sKjBnG8mQpL"

// Incorrect
wire := "MAT2xXFhP9w7V4sKjBnG8mQpL"  // Missing colon
```

### Invalid character in Base58

**Error:**
```
invalid character in Base58
```

**Cause:** Wire string contains characters not in Base58 alphabet.

**Solution:**
```go
// Valid characters: 123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz
// Invalid: 0, O, I, l

// Correct
wire := "MAT:2xXFhP9w7V4sKjBnG8mQpL"

// Incorrect
wire := "MAT:2xXFhP9w7V4sKjBnG8mQpI"  // Contains 'I'
```

### Checksum mismatch

**Error:**
```
checksum mismatch
```

**Cause:** Wire string was corrupted or modified.

**Solution:**
- Ensure you're using the exact wire string returned by SNID
- Check for transmission errors
- Regenerate the ID if needed

### Invalid length

**Error:**
```
invalid length
```

**Cause:** Decoded bytes are not 16 bytes (or 32 for extended families).

**Solution:**
```go
// SNID: 16 bytes
// NID, LID, etc.: 32 bytes

// Ensure correct ID type for your use case
```

## Encoding Errors

### Invalid atom

**Error:**
```
invalid atom
```

**Cause:** Atom is not a valid canonical atom.

**Solution:**
```go
// Valid atoms: IAM, TEN, MAT, LOC, CHR, LED, LEG, TRU, KIN, COG, SEM, SYS, EVT, SES, KEY

// Correct
id.String(snid.Matter)

// Incorrect
id.String("INVALID")
```

## Python-Specific Errors

### Module not found

**Error:**
```
ModuleNotFoundError: No module named 'snid'
```

**Cause:** SNID not installed.

**Solution:**
```bash
pip install snid
```

### Native module not found

**Error:**
```
ImportError: snid_native module not found
```

**Cause:** Native Rust extension not built.

**Solution:**
```bash
cd python
maturin develop
```

### NumPy not available

**Error:**
```
ModuleNotFoundError: No module named 'numpy'
```

**Cause:** NumPy not installed but numpy backend requested.

**Solution:**
```bash
pip install numpy
# Or use different backend
batch = snid.SNID.generate_batch(1000, backend="bytes")
```

## Go-Specific Errors

### Package not found

**Error:**
```
cannot find package "github.com/LastMile-Innovations/snid"
```

**Cause:** Go module not downloaded.

**Solution:**
```bash
cd go
go mod download
go mod tidy
```

### Build errors

**Error:**
```
build error: undefined: snid.NewFast
```

**Cause:** Using incorrect function name or outdated version.

**Solution:**
```bash
cd go
go mod tidy
go test ./...
```

## Rust-Specific Errors

### Cargo not found

**Error:**
```
cargo: command not found
```

**Cause:** Rust/Cargo not installed.

**Solution:**
```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env
```

### Build errors

**Error:**
```
error[E0433]: failed to resolve: use of undeclared crate
```

**Cause:** Missing dependency or incorrect version.

**Solution:**
```bash
cd rust
cargo update
cargo build
```

## Conformance Errors

### Conformance test failure

**Error:**
```
conformance test failed: expected X, got Y
```

**Cause:** Implementation doesn't match canonical vectors.

**Solution:**
1. Regenerate vectors: `just conformance`
2. Check encoding/decoding logic
3. Verify byte layout matches SPEC.md
4. Check all three implementations

### Vector generation failure

**Error:**
```
failed to generate vectors
```

**Cause:** Go vector generator error.

**Solution:**
```bash
cd conformance/cmd/generate_vectors
go run . --out ../../vectors.json
```

## Database Errors

### Invalid UUID format

**Error:**
```
invalid UUID format
```

**Cause:** Invalid UUID conversion.

**Solution:**
```go
// Ensure correct UUID format
id, err := snid.ParseUUIDString("018f1c3e-5a7b-7c8d-9e0f-1a2b3c4d5e6f")
if err != nil {
    // Handle error
}
```

### Binary storage error

**Error:**
```
invalid byte array length
```

**Cause:** Incorrect binary storage length.

**Solution:**
```go
// SNID: 16 bytes
// Ensure correct length
if len(bytes) != 16 {
    return errors.New("invalid length")
}
```

## Performance Issues

### Slow ID generation

**Symptom:** ID generation is slower than expected.

**Solutions:**
- Use `NewFast()` instead of `New()` in Go
- Use `backend="bytes"` in Python
- Use `generate_batch()` for bulk operations
- Enable release mode in Rust: `cargo test --release`

### High memory usage

**Symptom:** High memory consumption when generating IDs.

**Solutions:**
- Process in chunks instead of all at once
- Use `backend="bytes"` instead of `backend="snid"` in Python
- Reuse buffers in Go
- Use zero-copy views in NumPy

## Installation Issues

### mise not found

**Error:**
```
mise: command not found
```

**Solution:**
```bash
curl https://mise.run | sh
```

### just not found

**Error:**
```
just: command not found
```

**Solution:**
```bash
cargo install just
```

### pre-commit not found

**Error:**
```
pre-commit: command not found
```

**Solution:**
```bash
pip install pre-commit
```

## Getting Help

If you encounter an error not listed here:

1. Check the [FAQ](faq.md) for common questions
2. Search [GitHub Issues](https://github.com/LastMile-Innovations/snid/issues)
3. Create a new issue with:
   - Error message
   - Code snippet
   - Environment details (OS, language version)
   - Steps to reproduce

## Next Steps

- [FAQ](faq.md) - Frequently asked questions
- [Debugging](debugging.md) - Debugging tips
- [Contributing](../../CONTRIBUTING.md) - Contribution guidelines

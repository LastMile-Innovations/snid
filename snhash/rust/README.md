# SNHASH Rust

Rust implementation of the SNHASH companion package. It mirrors the Go v1
surface:

- `Digest` with `h1:<algorithm>:<base32-lower-no-padding>` wire format
- `HashId` with `hid1:<profile>:<size>:<algorithm>:<digest>` wire format
- Profiles: `FastCAS`, `Evidence`, `API`, `FIPS`, and `Cache`
- deterministic manifest JSON, proof roots, chunk roots, directory manifests
- HMAC-SHA256 and keyed BLAKE3 helpers
- optional SNID `Bid` bridge with the `snid-bridge` feature

```rust
use snhash::{hash_bytes_profile, Profile};

let manifest = hash_bytes_profile(b"evidence bytes", Profile::Evidence)?;
let hash_id = manifest.hash_id()?;
let proof = manifest.proof_root()?;

println!("{}", hash_id.wire());
println!("{}", proof.wire());
# Ok::<(), Box<dyn std::error::Error>>(())
```

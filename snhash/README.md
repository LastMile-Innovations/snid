# SNHASH

SNHASH is a Go-first companion package for SNID. It handles content integrity,
evidence manifests, content-addressable storage, and API verification without
changing the SNID identifier protocol.

SNID remains the topology and sortable identity layer. SNHASH produces
canonical digest envelopes that can be attached to SNID `BID` values through
`snhash/snidbridge`.

## Defaults

- `FastCAS`: BLAKE3-256 only, optimized for hot content keys.
- `Evidence`: BLAKE3-256 plus SHA-256 for legal and institutional interchange.
- `API`: SHA-256 plus BLAKE3-256 for request/body proof workflows.
- `FIPS`: SHA-256, SHA3-256, and SHAKE256-256.
- `Cache`: XXH3-64 for fast cache bucketing plus BLAKE3-256 confirmation.

XXH3, MD5, and SHA-1 are accepted only as cache/import metadata. They are never
marked trusted for current verification policy.

## Example

```go
manifest, err := snhash.HashFile(
	"evidence.pdf",
	snhash.WithProfile(snhash.ProfileEvidence),
	snhash.WithChunkSize(4*1024*1024),
)
if err != nil {
	panic(err)
}

root, err := snhash.ProofRoot(manifest)
if err != nil {
	panic(err)
}

hashID, err := manifest.HashID()
if err != nil {
	panic(err)
}

bid, err := snidbridge.AttachSNID(manifest)
if err != nil {
	panic(err)
}

_ = hashID.Wire()
_ = root.Wire()
_ = bid.WireFormat()
```

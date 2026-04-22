import snid

# Generate a new SNID
id = snid.SNID.new_fast()
print(f"Generated ID: {id}")

# Format as wire string with atom
wire = id.to_wire("MAT")
print(f"Wire format: {wire}")

# Parse a wire string
parsed, atom = snid.SNID.parse_wire(wire)
print(f"Parsed ID: {parsed} (atom: {atom})")

# Generate batch of IDs
batch = snid.SNID.generate_batch(5, backend="snid")
print(f"Generated {len(batch)} batch IDs")
for i, id in enumerate(batch):
    print(f"  [{i}] {id.to_wire('MAT')}")

# Generate batch as raw bytes
batch_bytes = snid.SNID.generate_batch(5, backend="bytes")
print(f"Generated {len(batch_bytes)} bytes ({len(batch_bytes)//16} IDs)")

# Generate batch as tensor pairs
batch_tensor = snid.SNID.generate_batch(5, backend="tensor")
print(f"Generated {len(batch_tensor)} tensor pairs")

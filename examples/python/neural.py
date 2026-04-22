import snid

print("=== Neural ID Example ===")

# Create a neural ID with semantic tail
# NID = SNID head + 16-byte semantic tail (for vector search, LSH, etc.)

try:
    # Create a base SNID
    base_snid = snid.SNID.new_fast()
    
    # Semantic hash (e.g., from vector embedding)
    semantic_hash = b'\x00' * 16  # Placeholder for actual semantic hash
    
    nid = snid.NID.from_parts(base_snid, semantic_hash)
    print(f"Created NID: {nid}")
    print(f"  Base SNID: {base_snid}")
    print(f"  Semantic hash: {semantic_hash.hex()}")
except AttributeError:
    print("NID API not yet available in Python bindings")
    print("This is a placeholder for future implementation")

# For now, use Go for neural IDs:
# import subprocess
# result = subprocess.run(["go", "run", "examples/go/neural/main.go"], capture_output=True)
# print(result.stdout.decode())

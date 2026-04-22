import snid

# Note: SGID (Spatial ID) requires H3 geospatial encoding
# This example shows the API usage

print("=== Spatial ID Example ===")

# Create a spatial ID from H3 cell and entropy
# H3 cell for San Francisco (resolution 12)
h3_cell = 0x8a2a1072b59ffff  # Example H3 cell
entropy = 12345

try:
    sgid = snid.SGID.from_spatial_parts(h3_cell, entropy)
    print(f"Created SGID: {sgid}")
    print(f"  H3 cell: {hex(h3_cell)}")
    print(f"  Entropy: {entropy}")
except AttributeError:
    print("SGID API not yet available in Python bindings")
    print("This is a placeholder for future implementation")

# For now, use Go for spatial IDs:
# import subprocess
# result = subprocess.run(["go", "run", "examples/go/spatial/main.go"], capture_output=True)
# print(result.stdout.decode())

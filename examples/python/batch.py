import snid

# Generate batch as raw bytes (fastest)
print("=== Raw Bytes Backend ===")
batch_bytes = snid.SNID.generate_batch(1000, backend="bytes")
print(f"Generated {len(batch_bytes)} bytes ({len(batch_bytes)//16} IDs)")

# Generate batch as tensor pairs
print("\n=== Tensor Backend ===")
batch_tensor = snid.SNID.generate_batch(10, backend="tensor")
print(f"Generated {len(batch_tensor)} tensor pairs")
for i, (hi, lo) in enumerate(batch_tensor[:3]):
    print(f"  [{i}] hi={hi}, lo={lo}")

# Generate batch as NumPy arrays (requires snid[data])
try:
    import numpy as np
    print("\n=== NumPy Backend ===")
    batch_numpy = snid.SNID.generate_batch(10, backend="numpy")
    print(f"Generated NumPy array with shape: {batch_numpy.shape}")
    print(f"First 3 IDs:")
    print(batch_numpy[:3])
except ImportError:
    print("\n=== NumPy Backend (skipped - numpy not installed) ===")

# Generate batch as PyArrow arrays (requires snid[data])
try:
    import pyarrow as pa
    print("\n=== PyArrow Backend ===")
    batch_arrow = snid.SNID.generate_batch(10, backend="pyarrow")
    print(f"Generated PyArrow array with {len(batch_arrow)} elements")
    print(f"Type: {batch_arrow.type}")
except ImportError:
    print("\n=== PyArrow Backend (skipped - pyarrow not installed) ===")

# Generate batch as Polars series (requires snid[data])
try:
    import polars as pl
    print("\n=== Polars Backend ===")
    batch_polars = snid.SNID.generate_batch(10, backend="polars")
    print(f"Generated Polars series with {len(batch_polars)} elements")
    print(batch_polars.head(3))
except ImportError:
    print("\n=== Polars Backend (skipped - polars not installed) ===")

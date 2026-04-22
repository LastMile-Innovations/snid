#!/usr/bin/env python3
"""
SNID Property-Based Tests using Hypothesis
Tests invariants across generated inputs to ensure protocol correctness.
"""

import sys
from pathlib import Path

# Add parent directory to path for snid import
sys.path.insert(0, str(Path(__file__).parent.parent))

try:
    from snid import SNID
except ImportError:
    print("⚠️  SNID module not available. Run: cd python && maturin develop")
    sys.exit(1)

from hypothesis import given, strategies as st, settings
import pytest


# =============================================================================
# Property Tests
# =============================================================================

@given(st.integers(min_value=0, max_value=2**48 - 1))
@settings(max_examples=100)
def test_timestamp_roundtrip(timestamp_ms):
    """Timestamp should be preserved through encode/decode cycle."""
    id1 = SNID.from_timestamp(timestamp_ms)
    extracted_ts = id1.timestamp()
    assert extracted_ts == timestamp_ms, f"Timestamp mismatch: {extracted_ts} != {timestamp_ms}"


@given(st.integers(min_value=0, max_value=2**48 - 1))
@settings(max_examples=100)
def test_sorting_invariant(timestamp_ms):
    """IDs with larger timestamps should sort after IDs with smaller timestamps."""
    id1 = SNID.from_timestamp(timestamp_ms)
    id2 = SNID.from_timestamp(timestamp_ms + 1)
    assert id1 < id2, "Sorting invariant violated: timestamp ordering"


@given(st.integers(min_value=0, max_value=2**48 - 1))
@settings(max_examples=100)
def test_monotonicity(timestamp_ms):
    """Sequential IDs should have non-decreasing timestamps."""
    id1 = SNID.from_timestamp(timestamp_ms)
    id2 = SNID.from_timestamp(timestamp_ms)
    assert id1.timestamp() <= id2.timestamp(), "Monotonicity violated"


@given(st.integers(min_value=0, max_value=2**48 - 1))
@settings(max_examples=100)
def test_string_roundtrip(timestamp_ms):
    """String representation should roundtrip through parse."""
    id1 = SNID.from_timestamp(timestamp_ms)
    str_repr = str(id1)
    id2 = SNID.parse(str_repr)
    assert id1 == id2, f"String roundtrip failed: {id1} != {id2}"


@given(st.integers(min_value=0, max_value=2**48 - 1))
@settings(max_examples=100)
def test_bytes_roundtrip(timestamp_ms):
    """Bytes representation should roundtrip through from_bytes."""
    id1 = SNID.from_timestamp(timestamp_ms)
    bytes_repr = id1.to_bytes()
    id2 = SNID.from_bytes(bytes_repr)
    assert id1 == id2, f"Bytes roundtrip failed: {id1} != {id2}"


@given(st.integers(min_value=0, max_value=2**48 - 1))
@settings(max_examples=100)
def test_base58_roundtrip(timestamp_ms):
    """Base58 encoding should roundtrip."""
    id1 = SNID.from_timestamp(timestamp_ms)
    base58 = id1.to_base58()
    id2 = SNID.from_base58(base58)
    assert id1 == id2, f"Base58 roundtrip failed: {id1} != {id2}"


@given(st.integers(min_value=0, max_value=2**48 - 1))
@settings(max_examples=100)
def test_base32_roundtrip(timestamp_ms):
    """Base32 encoding should roundtrip."""
    id1 = SNID.from_timestamp(timestamp_ms)
    base32 = id1.to_base32()
    id2 = SNID.from_base32(base32)
    assert id1 == id2, f"Base32 roundtrip failed: {id1} != {id2}"


@given(st.integers(min_value=0, max_value=1000))
@settings(max_examples=50)
def test_batch_uniqueness(count):
    """All IDs in a batch should be unique."""
    ids = [SNID.new() for _ in range(count)]
    unique_ids = set(ids)
    assert len(unique_ids) == count, f"Batch uniqueness violated: {len(unique_ids)} != {count}"


@given(st.integers(min_value=0, max_value=2**48 - 1))
@settings(max_examples=100)
def test_id_length(timestamp_ms):
    """ID should always be 16 bytes."""
    id1 = SNID.from_timestamp(timestamp_ms)
    assert len(id1.to_bytes()) == 16, f"ID length incorrect: {len(id1.to_bytes())} != 16"


@given(st.integers(min_value=0, max_value=2**48 - 1))
@settings(max_examples=100)
def test_version_bits(timestamp_ms):
    """Version bits should be set correctly for UUIDv7 compatibility."""
    id1 = SNID.from_timestamp(timestamp_ms)
    bytes_repr = id1.to_bytes()
    # Version is in bits 48-51 (byte 6, bits 4-7)
    version = (bytes_repr[6] >> 4) & 0x0F
    assert version == 0x7, f"Version bits incorrect: {version} != 0x7"


@given(st.integers(min_value=0, max_value=2**48 - 1))
@settings(max_examples=100)
def test_variant_bits(timestamp_ms):
    """Variant bits should be set correctly for RFC 4122."""
    id1 = SNID.from_timestamp(timestamp_ms)
    bytes_repr = id1.to_bytes()
    # Variant is in bits 64-65 (byte 8, bits 6-7)
    variant = (bytes_repr[8] >> 6) & 0b11
    assert variant == 0b10, f"Variant bits incorrect: {variant} != 0b10"


# =============================================================================
# Test Suite
# =============================================================================

if __name__ == "__main__":
    pytest.main([__file__, "-v"])

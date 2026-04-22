#!/usr/bin/env python3
"""
Python Ecosystem Benchmark for ID Generation Libraries
Benchmarks 8 popular Python ID packages against SNID.
"""

import pytest
import uuid as stdlib_uuid
import sys

# Add parent directory to path for snid import
from pathlib import Path
sys.path.insert(0, str(Path(__file__).parent.parent / "python"))

try:
    from snid import SNID
except ImportError:
    SNID = None

try:
    import ulid
except ImportError:
    ulid = None

try:
    import ksuid
except ImportError:
    ksuid = None

try:
    import nanoid
except ImportError:
    nanoid = None

try:
    import cuid2
except ImportError:
    cuid2 = None

try:
    import snowflake
except ImportError:
    snowflake = None

try:
    import timeflake
except ImportError:
    timeflake = None


# Industry Standard Baseline: UUIDv7
@pytest.mark.benchmark
def test_stdlib_uuid7(benchmark):
    """Benchmark stdlib UUID v7 (Python 3.14+)."""
    if not hasattr(stdlib_uuid, 'uuid7'):
        pytest.skip("UUIDv7 requires Python 3.14+")
    benchmark(stdlib_uuid.uuid7)


# SNID Baseline
@pytest.mark.benchmark
def test_snid_new(benchmark):
    """Benchmark SNID generation."""
    if SNID is None:
        pytest.skip("SNID not available")
    benchmark(SNID.new)


@pytest.mark.benchmark
def test_snid_new_fast(benchmark):
    """Benchmark SNID fast generation."""
    if SNID is None:
        pytest.skip("SNID not available")
    benchmark(SNID.new_fast)


@pytest.mark.benchmark
def test_stdlib_uuid4(benchmark):
    """Benchmark stdlib UUID v4."""
    benchmark(stdlib_uuid.uuid4)


@pytest.mark.benchmark
def test_ulid_new(benchmark):
    """Benchmark ULID generation."""
    if ulid is None:
        pytest.skip("ulid-py not available")
    benchmark(ulid.new)


@pytest.mark.benchmark
def test_ksuid_new(benchmark):
    """Benchmark KSUID generation."""
    if ksuid is None:
        pytest.skip("ksuid not available")
    benchmark(ksuid.ksuid)


@pytest.mark.benchmark
def test_nanoid_generate(benchmark):
    """Benchmark NanoID generation."""
    if nanoid is None:
        pytest.skip("nanoid not available")
    benchmark(nanoid.generate)


@pytest.mark.benchmark
def test_cuid2_generate(benchmark):
    """Benchmark CUID2 generation."""
    if cuid2 is None:
        pytest.skip("cuid2 not available")
    benchmark(cuid2.cuid)


@pytest.mark.benchmark
def test_snowflake_id(benchmark):
    """Benchmark Snowflake ID generation."""
    if snowflake is None:
        pytest.skip("snowflake-id not available")
    generator = snowflake.SnowflakeGenerator(42)
    benchmark(generator.generate_id)


@pytest.mark.benchmark
def test_timeflake_new(benchmark):
    """Benchmark Timeflake generation."""
    if timeflake is None:
        pytest.skip("timeflake not available")
    benchmark(timeflake.Timeflake.random)

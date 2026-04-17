from __future__ import annotations

import pytest

try:
    from snid import SNID
except RuntimeError:  # pragma: no cover
    SNID = None

try:
    import numpy as np  # noqa: F401
except ImportError:  # pragma: no cover
    np = None

pytestmark = pytest.mark.skipif(SNID is None, reason="snid_native is not built")


def test_bench_snid_new_fast(benchmark):
    result = benchmark(SNID.new_fast)
    assert result.to_bytes()


def test_bench_snid_to_wire(benchmark):
    id_obj = SNID.new_fast()
    result = benchmark(id_obj.to_wire, "MAT")
    assert result.startswith("MAT:")


def test_bench_snid_to_llm_format(benchmark):
    id_obj = SNID.new_fast()
    result = benchmark(id_obj.to_llm_format, "MAT")
    assert result[0] == "MAT"


def test_bench_batch_tensor(benchmark):
    batch_size = 100_000
    result = benchmark(SNID.generate_batch, batch_size, backend="tensor")
    assert len(result) == batch_size
    assert all(len(pair) == 2 for pair in result[:8])


@pytest.mark.skipif(np is None, reason="numpy unavailable")
def test_bench_batch_numpy(benchmark):
    batch_size = 100_000
    result = benchmark(SNID.generate_batch, batch_size, backend="numpy")
    assert result.shape == (batch_size, 2)
    assert str(result.dtype) == ">i8"


@pytest.mark.skipif(np is None, reason="numpy unavailable")
def test_bench_tensor_time_delta(benchmark):
    tensors_a = SNID.generate_batch(100_000, backend="numpy")
    tensors_b = SNID.generate_batch(100_000, backend="numpy")
    result = benchmark(SNID.tensor_time_delta, tensors_a, tensors_b)
    assert result.shape == (100_000,)

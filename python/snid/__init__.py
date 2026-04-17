from __future__ import annotations

from pathlib import Path
import json
import struct
from typing import Any

try:
    from .snid_native import (
        BID,
        EID,
        KID,
        LID,
        NID,
        SGID,
        SNID,
        WID,
        XID,
        decode_fixed64_pair as _decode_fixed64_pair,
        encode_fixed64_pair as _encode_fixed64_pair,
        generate_batch_bytes as _generate_batch_bytes,
        generate_batch_tensor_bytes as _generate_batch_tensor_bytes,
    )
except Exception as exc:  # pragma: no cover
    raise RuntimeError(
        "snid_native is not built. Run `maturin develop` or `maturin build` in the SNID `python/` directory."
    ) from exc

from . import neo4j


def load_vectors(path: str | None = None) -> dict:
    repo_root = Path(__file__).resolve().parents[2]
    vector_path = Path(path) if path else repo_root / "conformance" / "vectors.json"
    return json.loads(vector_path.read_text())


def _extract_high_word(value: Any) -> Any:
    try:
        import numpy as np  # type: ignore

        if isinstance(value, np.ndarray):
            if value.shape[-1] != 2:
                raise ValueError("expected trailing tensor dimension of size 2")
            return value[..., 0]
    except Exception:
        pass

    if isinstance(value, (list, tuple)):
        if len(value) == 2 and all(isinstance(v, int) for v in value):
            return value[0]
        return [_extract_high_word(item) for item in value]
    raise TypeError("expected a tensor pair, sequence of pairs, or numpy ndarray")


def _tensor_time_delta(left: Any, right: Any) -> Any:
    try:
        import numpy as np  # type: ignore

        if isinstance(left, np.ndarray) or isinstance(right, np.ndarray):
            left_hi = _extract_high_word(left)
            right_hi = _extract_high_word(right)
            return (left_hi >> 16) - (right_hi >> 16)
    except Exception:
        pass

    left_hi = _extract_high_word(left)
    right_hi = _extract_high_word(right)
    if isinstance(left_hi, list) and isinstance(right_hi, list):
        if len(left_hi) != len(right_hi):
            raise ValueError("left and right must have the same length")
        return [_tensor_time_delta((l, 0), (r, 0)) for l, r in zip(left_hi, right_hi)]
    return (int(left_hi) >> 16) - (int(right_hi) >> 16)


def _generate_batch(n: int, *, format: str = "snid", atom: str = "MAT", backend: str | None = None) -> Any:
    if n < 0:
        raise ValueError("n must be non-negative")
    backend = backend or format

    if backend == "bytes":
        return _generate_batch_bytes(n)

    if backend == "tensor":
        raw = _generate_batch_tensor_bytes(n)
        return list(struct.iter_unpack(">qq", raw))

    if backend == "numpy":
        import numpy as np  # type: ignore

        raw = _generate_batch_tensor_bytes(n)
        return np.frombuffer(raw, dtype=">i8").reshape(n, 2)

    if backend == "pyarrow":
        import pyarrow as pa  # type: ignore

        raw = _generate_batch_bytes(n)
        return pa.array([raw[i : i + 16] for i in range(0, len(raw), 16)], type=pa.binary(16))

    if backend == "polars":
        import polars as pl  # type: ignore

        raw = _generate_batch_bytes(n)
        return pl.Series("snid", [raw[i : i + 16] for i in range(0, len(raw), 16)], dtype=pl.Binary)

    if backend == "snid":
        raw = _generate_batch_bytes(n)
        return [SNID.from_bytes(raw[i : i + 16]) for i in range(0, len(raw), 16)]

    if backend == "llm":
        return [item.to_llm_format(atom) for item in _generate_batch(n, backend="snid")]

    raise ValueError(f"unsupported batch backend: {backend}")


SNID.tensor_time_delta = staticmethod(_tensor_time_delta)
SNID.generate_batch = staticmethod(_generate_batch)
SNID.encode_fixed64_pair = staticmethod(_encode_fixed64_pair)
SNID.decode_fixed64_pair = staticmethod(_decode_fixed64_pair)


__all__ = ["SNID", "SGID", "NID", "WID", "XID", "KID", "LID", "EID", "BID", "load_vectors", "neo4j"]

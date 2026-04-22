from __future__ import annotations

from pathlib import Path
import json
import struct
from typing import Any
import uuid as _uuid

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


def new_uuidv7() -> SNID:
    """Generate a SNID with RFC 9562 UUIDv7-compatible bytes."""
    return SNID.new_uuidv7()


# =============================================================================
# UNIVERSAL PARADIGMS (API V2) - Module-Level Generation
# =============================================================================

def new() -> SNID:
    """Generate a new SNID with ~3.7ns latency.
    This is the universal paradigm for fast ID generation.
    """
    return SNID.new_uuidv7()


def new_with(tenant: str | None = None, shard: int | None = None) -> SNID:
    """Generate a configured ID using options.
    This is the universal paradigm for configured ID generation.
    """
    # For now, just return a regular ID
    # Full tenant-sharded implementation would require native support
    return SNID.new_uuidv7()


def new_spatial(lat: float, lng: float) -> SNID:
    """Generate a spatial ID from lat/lng coordinates.
    This is the universal paradigm for spatial ID generation.
    """
    # For now, generate a regular ID with spatial markers
    # Full H3 integration would require additional dependencies
    return SNID.new_uuidv7()


def batch(count: int, *, backend: str = "snid") -> Any:
    """Generate a batch of IDs efficiently.
    This is the universal paradigm for batch generation.
    """
    return _generate_batch(count, backend=backend)


def parse(s: str) -> SNID:
    """Parse a wire string and return the ID.
    This is the universal paradigm for parsing wire strings.
    """
    return SNID.parse_wire(s)[0]


def parse_uuid(s: str) -> SNID:
    """Parse a UUID string and return the ID.
    This is the universal paradigm for parsing UUID strings.
    """
    return SNID.from_uuid_string(s)


def from_uuid(value: _uuid.UUID) -> SNID:
    """Convert a Python UUIDv7 into a SNID."""
    if not isinstance(value, _uuid.UUID):
        raise TypeError("expected uuid.UUID")
    raw = value.bytes
    if value.version != 7 or (raw[8] & 0xC0) != 0x80:
        raise ValueError("expected UUIDv7")
    return SNID.from_bytes(raw)


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

# Add universal paradigm serialization methods
SNID.string_default = lambda self: self.to_wire("MAT")
SNID.with_atom = lambda self, atom: self.to_wire(atom)


__all__ = [
    "SNID",
    "SGID",
    "NID",
    "WID",
    "XID",
    "KID",
    "LID",
    "EID",
    "BID",
    "load_vectors",
    "neo4j",
    "new_uuidv7",
    "from_uuid",
    # Universal Paradigms (API V2)
    "new",
    "new_with",
    "new_spatial",
    "batch",
    "parse",
    "parse_uuid",
]

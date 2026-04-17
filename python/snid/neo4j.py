from __future__ import annotations

from typing import Any

from .snid_native import KID, LID, NID, SNID, WID, XID


def to_neo4j_value(value: SNID | NID | WID | XID | KID | LID | bytes | bytearray) -> bytes:
    if isinstance(value, SNID):
        return bytes(value.to_bytes())
    if isinstance(value, (NID, WID, XID, KID, LID)):
        return bytes(value.to_bytes())
    if isinstance(value, (bytes, bytearray)):
        if len(value) not in (16, 32):
            raise ValueError("expected 16-byte or 32-byte binary value")
        return bytes(value)
    raise TypeError(f"unsupported Neo4j value type: {type(value)!r}")


def from_neo4j_value(value: Any) -> SNID:
    if isinstance(value, SNID):
        return value
    if isinstance(value, (bytes, bytearray)):
        data = bytes(value)
        if len(data) != 16:
            raise ValueError("expected 16-byte binary value")
        return SNID.from_bytes(data)
    if isinstance(value, str):
        if len(value) == 32:
            return SNID.from_bytes(bytes.fromhex(value))
        parsed, _atom = SNID.parse_wire(value)
        return parsed
    raise TypeError(f"unsupported Neo4j value type: {type(value)!r}")


def bind_params(params: dict[str, Any] | None = None, **values: Any) -> dict[str, Any]:
    out = dict(params or {})
    for key, value in values.items():
        out[key] = to_neo4j_value(value)
    return out

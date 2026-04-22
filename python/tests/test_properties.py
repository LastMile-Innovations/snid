#!/usr/bin/env python3
"""SNID property tests against the current Python binding API."""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

try:
    from snid import SNID
except ImportError:
    print("SNID module not available. Run: cd python && maturin develop")
    sys.exit(1)


TIMESTAMPS = [0, 1, 1_700_000_000_123, 2**48 - 1]


def id_at(timestamp_ms: int, seed: bytes = b"property") -> SNID:
    return SNID.from_hash_with_timestamp(timestamp_ms, seed)


def test_timestamp_roundtrip() -> None:
    for timestamp_ms in TIMESTAMPS:
        id_obj = id_at(timestamp_ms)
        assert id_obj.timestamp_millis() == timestamp_ms


def test_timestamp_sorting_invariant() -> None:
    for timestamp_ms in TIMESTAMPS[:-1]:
        left = id_at(timestamp_ms).to_bytes()
        right = id_at(timestamp_ms + 1).to_bytes()
        assert left < right


def test_wire_roundtrip() -> None:
    for timestamp_ms in TIMESTAMPS:
        id_obj = id_at(timestamp_ms)
        wire = id_obj.to_wire("MAT")
        parsed, atom = SNID.parse_wire(wire)
        assert atom == "MAT"
        assert parsed.to_bytes() == id_obj.to_bytes()


def test_bytes_roundtrip() -> None:
    for timestamp_ms in TIMESTAMPS:
        id_obj = id_at(timestamp_ms)
        assert SNID.from_bytes(id_obj.to_bytes()).to_bytes() == id_obj.to_bytes()


def test_batch_uniqueness() -> None:
    raw = SNID.generate_batch(256, backend="bytes")
    chunks = [raw[i : i + 16] for i in range(0, len(raw), 16)]
    assert len(chunks) == 256
    assert len(set(chunks)) == 256


def test_uuidv7_bits_and_string_roundtrip() -> None:
    id_obj = SNID.new_uuidv7()
    raw = id_obj.to_bytes()
    assert len(raw) == 16
    assert (raw[6] >> 4) & 0x0F == 7
    assert (raw[8] >> 6) & 0b11 == 0b10

    uuid_string = id_obj.to_uuid_string()
    assert len(uuid_string) == 36
    assert [uuid_string[i] for i in (8, 13, 18, 23)] == ["-", "-", "-", "-"]
    assert SNID.parse_uuid_string(uuid_string).to_bytes() == raw


def test_parse_uuid_string_rejects_non_v7() -> None:
    try:
        SNID.parse_uuid_string("018f1c3e-5a7b-4c8d-9e0f-1a2b3c4d5e6f")
    except ValueError:
        return
    raise AssertionError("expected non-v7 UUID string to be rejected")

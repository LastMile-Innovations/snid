from __future__ import annotations

import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "python"))

from snid import BID, EID, KID, LID, NID, WID, XID, SNID, load_vectors  # noqa: E402


def main() -> int:
    vectors = load_vectors(str(ROOT / "conformance" / "vectors.json"))
    core = []
    for case in vectors["core"]:
        snid = SNID.from_bytes(bytes.fromhex(case["bytes_hex"]))
        atom, timestamp_millis, machine_or_shard, sequence = snid.to_llm_format(case["atom"])
        tensor_hi, tensor_lo = snid.to_tensor()
        core.append(
            {
                "bytes_hex": snid.to_bytes().hex(),
                "wire": snid.to_wire(case["atom"]),
                "tensor_hi": tensor_hi,
                "tensor_lo": tensor_lo,
                "atom": atom,
                "timestamp_millis": timestamp_millis,
                "machine_or_shard": machine_or_shard,
                "sequence": sequence,
            }
        )

    spatial = vectors["spatial"]
    sgid = SNID.from_bytes(bytes.fromhex(spatial["bytes_hex"]))
    sgid_v2 = sgid.to_llm_format_v2(spatial["atom"])
    neural = vectors["neural"]
    nid = NID.from_parts(
        SNID.from_bytes(bytes.fromhex(neural["head_hex"])),
        bytes.fromhex(neural["semantic_hex"]),
    )
    ledger = vectors["ledger"]
    lid = LID.from_parts(
        SNID.from_bytes(bytes.fromhex(ledger["head_hex"])),
        bytes.fromhex(ledger["prev_hex"]),
        bytes.fromhex(ledger["payload_hex"]),
        bytes.fromhex(ledger["key_hex"]),
    )
    world = vectors["world"]
    wid = WID.from_parts(
        SNID.from_bytes(bytes.fromhex(world["head_hex"])),
        bytes.fromhex(world["scenario_hex"]),
    )
    edge = vectors["edge"]
    xid = XID.from_parts(
        SNID.from_bytes(bytes.fromhex(edge["head_hex"])),
        bytes.fromhex(edge["edge_hex"]),
    )
    capability = vectors["capability"]
    actor = SNID.from_bytes(bytes.fromhex(capability["actor_hex"]))
    kid = KID.from_parts(
        SNID.from_bytes(bytes.fromhex(capability["head_hex"])),
        actor,
        bytes.fromhex(capability["resource_hex"]),
        bytes.fromhex(capability["capability_hex"]),
        bytes.fromhex(capability["key_hex"]),
    )
    ephemeral = vectors["ephemeral"]
    eid = EID.from_parts(ephemeral["timestamp_millis"], ephemeral["counter"])
    bid_case = vectors["bid"]
    bid = BID.from_parts(
        SNID.from_bytes(bytes.fromhex(bid_case["topology_hex"])),
        bytes.fromhex(bid_case["content_hex"]),
    )

    out = {
        "core": core,
        "spatial": {
            "bytes_hex": sgid.to_bytes().hex(),
            "wire": sgid.to_wire(spatial["atom"]),
            "h3_cell_hex": format(sgid_v2[3], "x"),
        },
        "neural": {
            "bytes_hex": bytes(nid.to_bytes()).hex(),
            "hamming_to_zero": sum(bin(byte).count("1") for byte in bytes(nid.to_bytes())[16:]),
        },
        "ledger": {
            "bytes_hex": bytes(lid.to_bytes()).hex(),
        },
        "world": {
            "bytes_hex": bytes(wid.to_bytes()).hex(),
            "tensor_words": list(wid.to_tensor()),
        },
        "edge": {
            "bytes_hex": bytes(xid.to_bytes()).hex(),
            "tensor_words": list(xid.to_tensor()),
        },
        "capability": {
            "bytes_hex": bytes(kid.to_bytes()).hex(),
            "tensor_words": list(kid.to_tensor()),
            "verified": kid.verify(
                actor,
                bytes.fromhex(capability["resource_hex"]),
                bytes.fromhex(capability["capability_hex"]),
                bytes.fromhex(capability["key_hex"]),
            ),
        },
        "ephemeral": {
            "bytes_hex": bytes(eid.to_bytes()).hex(),
            "timestamp_millis": eid.timestamp_millis(),
            "counter": eid.counter(),
        },
        "bid": {
            "wire": bid.to_wire(),
            "r2_key": bid.r2_key(),
            "neo4j_id": SNID.from_bytes(bytes.fromhex(bid_case["topology_hex"])).to_wire("MAT"),
        },
    }
    print(json.dumps(out, indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

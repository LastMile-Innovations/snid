import json
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
PYTHON_ROOT = REPO_ROOT / "python"
VECTORS_PATH = REPO_ROOT / "conformance" / "vectors.json"

try:
    import sys

    sys.path.insert(0, str(PYTHON_ROOT))
    from snid import KID, SNID, WID, XID, load_vectors, neo4j
except Exception:  # pragma: no cover
    KID = None
    SNID = None
    WID = None
    XID = None
    load_vectors = None
    neo4j = None


class VectorSchemaTest(unittest.TestCase):
    def test_vectors_file_exists(self) -> None:
        self.assertTrue(VECTORS_PATH.exists(), "expected generated vectors.json")
        payload = json.loads(VECTORS_PATH.read_text())
        self.assertIn("core", payload)
        self.assertIn("spatial", payload)
        self.assertIn("neural", payload)
        self.assertIn("ledger", payload)
        self.assertIn("world", payload)
        self.assertIn("edge", payload)
        self.assertIn("capability", payload)
        self.assertIn("ephemeral", payload)
        self.assertIn("bid", payload)
        self.assertIn("compatibility", payload)
        self.assertIn("negative", payload)

    @unittest.skipIf(SNID is None, "snid extension not built")
    def test_boundary_vectors_match_python_bindings(self) -> None:
        payload = load_vectors()
        core = payload["core"][0]
        snid = SNID.from_bytes(bytes.fromhex(core["bytes_hex"]))
        self.assertEqual(snid.to_tensor(), (core["tensor_hi"], core["tensor_lo"]))
        llm = core["llm_format"]
        self.assertEqual(
            snid.to_llm_format(core["atom"]),
            (llm["atom"], llm["timestamp_millis"], llm["machine_or_shard"], llm["sequence"]),
        )
        llm2 = core["llm_format_v2"]
        self.assertEqual(
            snid.to_llm_format_v2(core["atom"]),
            (
                llm2["kind"],
                llm2["atom"],
                llm2["timestamp_millis"],
                llm2["spatial_anchor"],
                llm2["machine_or_shard"],
                llm2["sequence"],
                llm2["ghosted"],
            ),
        )
        self.assertEqual(snid.time_bin(3_600_000), core["time_bin_hour"])
        self.assertTrue(snid.with_ghost_bit(True).is_ghosted())

        parsed, atom = SNID.parse_wire(core["wire"])
        self.assertEqual(atom, core["atom"])
        self.assertEqual(parsed.to_bytes(), snid.to_bytes())

    @unittest.skipIf(SNID is None, "snid extension not built")
    def test_tensor_time_delta_and_neo4j_helpers(self) -> None:
        left = SNID.from_hash_with_timestamp(1700000000123, b"alpha").to_tensor()
        right = SNID.from_hash_with_timestamp(1700000000000, b"alpha").to_tensor()
        self.assertEqual(SNID.tensor_time_delta(left, right), 123)

        node_id = SNID.from_hash_with_timestamp(1700000000123, b"graph")
        bound = neo4j.bind_params(id=node_id)
        self.assertEqual(len(bound["id"]), 16)
        round_trip = neo4j.from_neo4j_value(bound["id"])
        self.assertEqual(round_trip.to_bytes(), node_id.to_bytes())
        encoded = SNID.encode_fixed64_pair(*node_id.to_tensor())
        self.assertEqual(SNID.decode_fixed64_pair(encoded), node_id.to_tensor())

    @unittest.skipIf(SNID is None, "snid extension not built")
    def test_native_batch_buffers(self) -> None:
        raw = SNID.generate_batch(4, backend="bytes")
        self.assertIsInstance(raw, bytes)
        self.assertEqual(len(raw), 64)

        pairs = SNID.generate_batch(4, backend="tensor")
        self.assertEqual(len(pairs), 4)
        self.assertTrue(all(len(pair) == 2 for pair in pairs))

    @unittest.skipIf(SNID is None or WID is None or XID is None or KID is None, "snid extension not built")
    def test_composite_targets(self) -> None:
        payload = load_vectors()
        head = SNID.from_bytes(bytes.fromhex(payload["world"]["head_hex"]))
        wid = WID.from_parts(head, bytes.fromhex(payload["world"]["scenario_hex"]))
        self.assertEqual(wid.to_bytes(), bytes.fromhex(payload["world"]["bytes_hex"]))
        self.assertEqual(tuple(wid.to_tensor()), tuple(payload["world"]["tensor_words"]))

        xid = XID.from_parts(head, bytes.fromhex(payload["edge"]["edge_hex"]))
        self.assertEqual(xid.to_bytes(), bytes.fromhex(payload["edge"]["bytes_hex"]))
        self.assertEqual(tuple(xid.to_tensor()), tuple(payload["edge"]["tensor_words"]))

        actor = SNID.from_bytes(bytes.fromhex(payload["capability"]["actor_hex"]))
        kid = KID.from_parts(
            head,
            actor,
            bytes.fromhex(payload["capability"]["resource_hex"]),
            bytes.fromhex(payload["capability"]["capability_hex"]),
            bytes.fromhex(payload["capability"]["key_hex"]),
        )
        self.assertEqual(kid.to_bytes(), bytes.fromhex(payload["capability"]["bytes_hex"]))
        self.assertEqual(tuple(kid.to_tensor()), tuple(payload["capability"]["tensor_words"]))
        self.assertTrue(
            kid.verify(
                actor,
                bytes.fromhex(payload["capability"]["resource_hex"]),
                bytes.fromhex(payload["capability"]["capability_hex"]),
                bytes.fromhex(payload["capability"]["key_hex"]),
            )
        )


if __name__ == "__main__":
    unittest.main()

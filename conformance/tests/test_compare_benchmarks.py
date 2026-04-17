from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "compare_benchmarks.py"


class CompareBenchmarksTest(unittest.TestCase):
    def test_go_passes_with_small_delta(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            base = Path(tmp) / "base.txt"
            cur = Path(tmp) / "cur.txt"
            base.write_text("BenchmarkSNIDNewFastParallel-8    1  10.0 ns/op  0 B/op  0 allocs/op\n")
            cur.write_text("BenchmarkSNIDNewFastParallel-8    1  10.4 ns/op  0 B/op  0 allocs/op\n")
            subprocess.run(
                ["python3", str(SCRIPT), "--language", "go", "--baseline", str(base), "--current", str(cur)],
                check=True,
            )

    def test_go_fails_with_large_delta(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            base = Path(tmp) / "base.txt"
            cur = Path(tmp) / "cur.txt"
            base.write_text("BenchmarkSNIDNewFastParallel-8    1  10.0 ns/op  0 B/op  0 allocs/op\n")
            cur.write_text("BenchmarkSNIDNewFastParallel-8    1  11.0 ns/op  0 B/op  0 allocs/op\n")
            proc = subprocess.run(
                ["python3", str(SCRIPT), "--language", "go", "--baseline", str(base), "--current", str(cur)],
                check=False,
            )
            self.assertNotEqual(proc.returncode, 0)

    def test_rust_passes_with_small_delta(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            base = Path(tmp) / "base.json"
            cur = Path(tmp) / "cur.json"
            payload = {"benchmarks": {"Snid::new_fast": {"ns_per_op": 10.0}}}
            base.write_text(json.dumps(payload))
            payload["benchmarks"]["Snid::new_fast"]["ns_per_op"] = 10.4
            cur.write_text(json.dumps(payload))
            subprocess.run(
                ["python3", str(SCRIPT), "--language", "rust", "--baseline", str(base), "--current", str(cur)],
                check=True,
            )

    def test_rust_fails_with_large_delta(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            base = Path(tmp) / "base.json"
            cur = Path(tmp) / "cur.json"
            base.write_text(json.dumps({"benchmarks": {"Snid::new_fast": {"ns_per_op": 10.0}}}))
            cur.write_text(json.dumps({"benchmarks": {"Snid::new_fast": {"ns_per_op": 11.0}}}))
            proc = subprocess.run(
                ["python3", str(SCRIPT), "--language", "rust", "--baseline", str(base), "--current", str(cur)],
                check=False,
            )
            self.assertNotEqual(proc.returncode, 0)


if __name__ == "__main__":
    unittest.main()

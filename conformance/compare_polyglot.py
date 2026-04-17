from __future__ import annotations

import json
import os
import subprocess
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


def run_json(*args: str, cwd: Path, env: dict[str, str] | None = None) -> dict:
    proc = subprocess.run(args, cwd=cwd, check=True, capture_output=True, text=True, env=env)
    return json.loads(proc.stdout)


def main() -> int:
    go_cwd = ROOT / "conformance" / "cmd" / "project_go"
    go_env = os.environ.copy()
    go_env["GOCACHE"] = str(go_cwd / ".gocache")
    go_env["GOTMPDIR"] = str(go_cwd / ".gotmp")
    (go_cwd / ".gocache").mkdir(exist_ok=True)
    (go_cwd / ".gotmp").mkdir(exist_ok=True)

    go = run_json("go", "run", ".", cwd=go_cwd, env=go_env)
    rust = run_json("cargo", "run", "--quiet", "--example", "project_vectors", cwd=ROOT / "rust")
    py = run_json(sys.executable, "scripts/project_vectors.py", cwd=ROOT / "python")

    if go != rust or go != py:
        print("polyglot projection mismatch", file=sys.stderr)
        if go != rust:
            print("Go != Rust", file=sys.stderr)
        if go != py:
            print("Go != Python", file=sys.stderr)
        return 1
    print("polyglot projection outputs match")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

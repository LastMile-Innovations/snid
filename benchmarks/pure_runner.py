#!/usr/bin/env python3
"""
SNID Pure Benchmark Runner
Isolated benchmark execution with zero harness overhead.
This script runs benchmarks without any FastAPI, logging, or dashboard code loaded.
"""

import subprocess
import json
import os
import sys
import time
from datetime import datetime
from pathlib import Path

ROOT = Path(__file__).parent.parent
RESULTS_DIR = Path(os.getenv("RESULTS_DIR", str(Path(__file__).parent / "results")))


def run_go_benchmarks():
    """Run Go benchmarks with minimal overhead."""
    result = subprocess.run(
        ["go", "test", "-bench=.", "-benchmem", "-count=5"],
        cwd=ROOT / "go",
        capture_output=True,
        text=True,
    )
    return {
        "language": "go",
        "returncode": result.returncode,
        "stdout": result.stdout,
        "stderr": result.stderr,
    }


def run_rust_benchmarks():
    """Run Rust benchmarks with minimal overhead."""
    result = subprocess.run(
        ["cargo", "bench", "--", "--save-baseline", "main"],
        cwd=ROOT / "rust",
        capture_output=True,
        text=True,
    )
    return {
        "language": "rust",
        "returncode": result.returncode,
        "stdout": result.stdout,
        "stderr": result.stderr,
    }


def run_python_benchmarks():
    """Run Python benchmarks with minimal overhead."""
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)
    output_file = RESULTS_DIR / "python_bench.json"
    
    result = subprocess.run(
        [
            "python",
            "-m",
            "pytest",
            "tests/test_bench.py",
            "--benchmark-only",
            f"--benchmark-json={output_file}",
        ],
        cwd=ROOT / "python",
        capture_output=True,
        text=True,
    )
    return {
        "language": "python",
        "returncode": result.returncode,
        "stdout": result.stdout,
        "stderr": result.stderr,
    }


def run_conformance_benchmark():
    """Run conformance validation with minimal overhead."""
    result = subprocess.run(
        ["python3", "benchmarks/conformance_runner.py"],
        cwd=ROOT,
        capture_output=True,
        text=True,
    )
    return {
        "language": "conformance",
        "returncode": result.returncode,
        "stdout": result.stdout,
        "stderr": result.stderr,
    }


def main():
    """Run benchmarks in pure mode - no harness overhead."""
    suites = {
        "go": run_go_benchmarks,
        "rust": run_rust_benchmarks,
        "python": run_python_benchmarks,
        "conformance": run_conformance_benchmark,
    }

    # Parse arguments
    if len(sys.argv) > 1:
        requested = sys.argv[1:]
        if requested[0] == "all":
            requested = list(suites.keys())
    else:
        requested = ["go", "rust", "python"]

    results = {
        "timestamp": datetime.now().isoformat(),
        "mode": "pure",
        "suites": {},
    }

    for suite_name in requested:
        if suite_name not in suites:
            continue
        results["suites"][suite_name] = suites[suite_name]()

    # Write results ONLY after all benchmarks complete
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    output_file = RESULTS_DIR / f"pure_run_{timestamp}.json"
    
    with open(output_file, "w") as f:
        json.dump(results, f, indent=2)

    # Print only the output file path (minimal output)
    print(output_file)

    return 0 if all(s["returncode"] == 0 for s in results["suites"].values()) else 1


if __name__ == "__main__":
    sys.exit(main())

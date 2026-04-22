#!/usr/bin/env python3
"""
SNID 2026 Master Benchmark Runner
Orchestrates all benchmark suites with statistical rigor and result storage.
Supports both web (JSON output) and CLI (human-readable) modes.
"""

import subprocess
import json
import os
import sys
from datetime import datetime
from pathlib import Path
import statistics

ROOT = Path(__file__).parent.parent
RESULTS_DIR = Path(os.getenv("RESULTS_DIR", str(Path(__file__).parent / "results")))
BENCH_MODE = os.getenv("BENCH_MODE", "cli")
BENCH_PURE_MODE = os.getenv("BENCH_PURE_MODE", "false").lower() == "true"
REGRESSION_THRESHOLD = float(os.getenv("REGRESSION_THRESHOLD", "10"))


def run_command(cmd, cwd=None):
    """Run a command and capture output."""
    if BENCH_MODE == "cli":
        print(f"\n▶ Running: {cmd}")
    if cwd is None:
        cwd = ROOT
    result = subprocess.run(cmd, shell=True, cwd=cwd, capture_output=True, text=True)
    return {
        "command": cmd,
        "stdout": result.stdout[-2000:] if result.stdout else "",
        "stderr": result.stderr[-2000:] if result.stderr else "",
        "returncode": result.returncode,
    }


def calculate_statistics(values):
    """Calculate statistical metrics for a list of values."""
    if not values:
        return {}
    sorted_values = sorted(values)
    n = len(values)
    return {
        "mean": statistics.mean(values),
        "median": statistics.median(values),
        "min": min(values),
        "max": max(values),
        "stddev": statistics.stdev(values) if n > 1 else 0,
        "p95": sorted_values[int(n * 0.95)] if n >= 20 else max(values),
        "p99": sorted_values[int(n * 0.99)] if n >= 100 else max(values),
    }


def run_go_benchmarks():
    """Run Go benchmarks with statistical sampling."""
    return run_command("cd go && go test -bench=. -benchmem -count=5")


def run_rust_benchmarks():
    """Run Rust benchmarks with Criterion."""
    return run_command("cd rust && cargo bench -- --save-baseline main")


def run_python_benchmarks():
    """Run Python benchmarks with pytest-benchmark."""
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)
    return run_command(
        f"cd python && python -m pytest tests/test_bench.py --benchmark-only --benchmark-json={RESULTS_DIR}/python_bench.json"
    )


def run_comparison_benchmarks():
    """Run comparison benchmarks against UUIDv7, ULID, NanoID."""
    return run_command("python benchmarks/comparison_benchmark.py")


def run_llm_benchmark():
    """Run LLM token efficiency benchmark."""
    return run_command("python benchmarks/llm_token_benchmark.py")


def run_ecosystem_benchmark():
    """Run ecosystem comparison benchmark against 20+ ID packages."""
    return run_command("python benchmarks/ecosystem_comparison.py")


def main():
    """Main entry point for benchmark runner."""
    # In pure mode, delegate to isolated pure_runner subprocess
    if BENCH_PURE_MODE:
        pure_runner_path = Path(__file__).parent / "pure_runner.py"
        result = subprocess.run(
            ["python3", str(pure_runner_path)] + sys.argv[1:],
            capture_output=True,
            text=True,
        )
        print(result.stdout.strip())
        return result.returncode

    suites = {
        "go": run_go_benchmarks,
        "rust": run_rust_benchmarks,
        "python": run_python_benchmarks,
        "comparison": run_comparison_benchmarks,
        "llm": run_llm_benchmark,
        "ecosystem": run_ecosystem_benchmark,
    }

    # Parse arguments
    if len(sys.argv) > 1:
        requested = sys.argv[1:]
        if requested[0] == "all":
            requested = list(suites.keys())
    else:
        # Check BENCH_SUITES env var for CLI mode
        bench_suites = os.getenv("BENCH_SUITES", "go,rust,python")
        if bench_suites == "all":
            requested = list(suites.keys())
        else:
            requested = bench_suites.split(",")

    results = {
        "timestamp": datetime.now().isoformat(),
        "mode": BENCH_MODE,
        "suites": {},
    }

    for suite_name in requested:
        if suite_name not in suites:
            if BENCH_MODE == "cli":
                print(f"⚠️  Unknown suite: {suite_name}")
            continue
        if BENCH_MODE == "cli":
            print(f"\n{'='*60}")
            print(f"Running {suite_name} benchmark suite")
            print(f"{'='*60}")
        results["suites"][suite_name] = suites[suite_name]()

    # Save results ONLY after all benchmarks complete
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    output_file = RESULTS_DIR / f"full_run_{timestamp}.json"
    with open(output_file, "w") as f:
        json.dump(results, f, indent=2)

    if BENCH_MODE == "cli":
        print(f"\n{'='*60}")
        print(f"✅ All benchmarks completed")
        print(f"📁 Results saved to: {output_file}")
        print(f"{'='*60}")

        # Print summary
        for suite_name, suite_result in results["suites"].items():
            status = "✅ PASSED" if suite_result["returncode"] == 0 else "❌ FAILED"
            print(f"{suite_name:12} {status}")
    else:
        # Web mode: return JSON path for API
        print(json.dumps({"output_file": str(output_file), "results": results}))

    return 0 if all(s["returncode"] == 0 for s in results["suites"].values()) else 1


if __name__ == "__main__":
    main()

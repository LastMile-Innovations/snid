#!/usr/bin/env python3
"""
SNID Polyglot Performance Benchmark
Runs Go, Rust, and Python benchmarks and aggregates results with statistical analysis.
"""

import subprocess
import json
import time
import os
import statistics
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Any, Optional

ROOT = Path(__file__).parent.parent
RESULTS_DIR = Path(os.getenv("RESULTS_DIR", str(Path(__file__).parent / "results")))


def calculate_statistics(values: List[float]) -> Dict[str, float]:
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
        "count": n,
    }


def calculate_confidence_interval(values: List[float], confidence: float = 0.95) -> Dict[str, float]:
    """Calculate confidence interval using t-distribution."""
    import math
    if len(values) < 2:
        return {"lower": 0, "upper": 0}
    
    n = len(values)
    mean = statistics.mean(values)
    stddev = statistics.stdev(values) if n > 1 else 0
    
    # Approximate t-value for 95% confidence (for large n, approaches 1.96)
    t_value = 1.96 if n >= 30 else 2.0  # Conservative estimate for small samples
    
    margin_of_error = t_value * (stddev / math.sqrt(n))
    
    return {
        "lower": mean - margin_of_error,
        "upper": mean + margin_of_error,
        "margin": margin_of_error,
        "confidence": confidence,
    }


def run_go_benchmarks():
    """Run Go benchmarks and parse output with statistics."""
    print("\n▶ Running Go benchmarks...")
    start = time.time()
    result = subprocess.run(
        ["go", "test", "-bench=.", "-benchmem", "-count=5"],
        cwd=ROOT / "go",
        capture_output=True,
        text=True,
    )
    duration = time.time() - start

    # Parse Go benchmark output with multiple samples
    benchmark_samples: Dict[str, List[float]] = {}
    for line in result.stdout.splitlines():
        if "Benchmark" in line and "ns/op" in line:
            parts = line.split()
            if len(parts) >= 3:
                name = parts[0]
                ns_per_op = float(parts[2])
                if name not in benchmark_samples:
                    benchmark_samples[name] = []
                benchmark_samples[name].append(ns_per_op)

    # Calculate statistics for each benchmark
    benchmarks = {}
    for name, samples in benchmark_samples.items():
        stats = calculate_statistics(samples)
        ci = calculate_confidence_interval(samples)
        benchmarks[name] = {
            "ns_per_op": stats.get("mean", 0),
            "statistics": stats,
            "confidence_interval": ci,
        }

    return {
        "language": "go",
        "duration_seconds": round(duration, 2),
        "benchmarks": benchmarks,
        "stdout": result.stdout[-1000:] if result.stdout else "",
        "returncode": result.returncode,
    }


def run_rust_benchmarks():
    """Run Rust benchmarks with Criterion and parse full statistics."""
    print("\n▶ Running Rust benchmarks...")
    start = time.time()
    result = subprocess.run(
        ["cargo", "bench", "--", "--save-baseline", "main"],
        cwd=ROOT / "rust",
        capture_output=True,
        text=True,
    )
    duration = time.time() - start

    # Parse Criterion output from JSON with full statistics
    benchmarks = {}
    criterion_dir = ROOT / "rust" / "target" / "criterion"
    if criterion_dir.exists():
        for estimates_file in criterion_dir.glob("**/new/estimates.json"):
            try:
                data = json.loads(estimates_file.read_text())
                bench_name = estimates_file.parts[-4]
                
                # Extract Criterion statistics
                mean_point = data["mean"]["point_estimate"]
                mean_ci = data["mean"]["confidence_interval"]
                stddev = data.get("std_dev", {}).get("point_estimate", 0)
                
                benchmarks[bench_name] = {
                    "ns_per_op": mean_point,
                    "statistics": {
                        "mean": mean_point,
                        "stddev": stddev,
                        "min": data.get("min", {}).get("point_estimate", mean_point),
                        "max": data.get("max", {}).get("point_estimate", mean_point),
                    },
                    "confidence_interval": {
                        "lower": mean_ci["lower_bound"],
                        "upper": mean_ci["upper_bound"],
                        "confidence": 0.95,
                    },
                }
            except (json.JSONDecodeError, KeyError):
                pass

    return {
        "language": "rust",
        "duration_seconds": round(duration, 2),
        "benchmarks": benchmarks,
        "stdout": result.stdout[-1000:] if result.stdout else "",
        "returncode": result.returncode,
    }


def run_python_benchmarks():
    """Run Python benchmarks with pytest-benchmark and parse full statistics."""
    print("\n▶ Running Python benchmarks...")
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)
    output_file = RESULTS_DIR / "python_bench.json"

    start = time.time()
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
    duration = time.time() - start

    # Parse pytest-benchmark JSON output with full statistics
    benchmarks = {}
    if output_file.exists():
        try:
            data = json.loads(output_file.read_text())
            for bench in data.get("benchmarks", []):
                name = bench["name"]
                stats = bench.get("stats", {})
                
                # Convert seconds to nanoseconds
                mean_ns = stats.get("mean", 0) * 1e9
                stddev_ns = stats.get("stddev", 0) * 1e9
                min_ns = stats.get("min", 0) * 1e9
                max_ns = stats.get("max", 0) * 1e9
                
                benchmarks[name] = {
                    "ns_per_op": mean_ns,
                    "statistics": {
                        "mean": mean_ns,
                        "median": stats.get("median", 0) * 1e9,
                        "min": min_ns,
                        "max": max_ns,
                        "stddev": stddev_ns,
                        "count": stats.get("rounds", 0),
                    },
                    "confidence_interval": {
                        "lower": stats.get("min", 0) * 1e9,
                        "upper": stats.get("max", 0) * 1e9,
                        "confidence": 0.95,
                    },
                }
        except (json.JSONDecodeError, KeyError):
            pass

    return {
        "language": "python",
        "duration_seconds": round(duration, 2),
        "benchmarks": benchmarks,
        "stdout": result.stdout[-1000:] if result.stdout else "",
        "returncode": result.returncode,
    }


def main():
    """Run all polyglot performance benchmarks."""
    print("=" * 60)
    print("SNID Polyglot Performance Benchmark")
    print("=" * 60)

    results = {
        "timestamp": datetime.now().isoformat(),
        "languages": {},
    }

    # Run each language
    results["languages"]["go"] = run_go_benchmarks()
    results["languages"]["rust"] = run_rust_benchmarks()
    results["languages"]["python"] = run_python_benchmarks()

    # Save results
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    output_file = RESULTS_DIR / f"performance_{timestamp}.json"
    with open(output_file, "w") as f:
        json.dump(results, f, indent=2)

    print(f"\n{'='*60}")
    print(f"✅ Performance benchmarks completed")
    print(f"📁 Results saved to: {output_file}")
    print(f"{'='*60}")

    # Print summary
    for lang, data in results["languages"].items():
        status = "✅ PASSED" if data["returncode"] == 0 else "❌ FAILED"
        print(f"{lang:8} {status} ({data['duration_seconds']:.2f}s)")

    return 0 if all(d["returncode"] == 0 for d in results["languages"].values()) else 1


if __name__ == "__main__":
    import sys

    sys.exit(main())

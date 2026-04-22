#!/usr/bin/env python3
"""
SNID Comparison Benchmark
Compares SNID performance against UUIDv7, ULID, NanoID, and KSUID.
Includes visual tables and statistical significance tests.
"""

import json
import sys
import os
import timeit
import statistics
from datetime import datetime
from pathlib import Path
from typing import Dict, Any, List, Optional

# Add parent directory to path for snid import
sys.path.insert(0, str(Path(__file__).parent.parent / "python"))

try:
    from snid import SNID
except ImportError:
    print("⚠️  SNID module not available. Run: cd python && maturin develop")
    SNID = None

RESULTS_DIR = Path(os.getenv("RESULTS_DIR", str(Path(__file__).parent / "results")))


def benchmark_snid():
    """Benchmark SNID generation with multiple samples."""
    if SNID is None:
        return {"error": "SNID module not available"}
    samples = []
    for _ in range(5):
        samples.append(timeit.timeit("SNID.new_fast()", setup="from snid import SNID", number=100_000))
    return {
        "samples": samples,
        "mean": statistics.mean(samples),
        "stddev": statistics.stdev(samples) if len(samples) > 1 else 0,
    }


def benchmark_uuidv7():
    """Benchmark UUIDv7 generation (Python 3.14+) with multiple samples."""
    try:
        import uuid
        samples = []
        for _ in range(5):
            samples.append(timeit.timeit("uuid.uuid7()", setup="import uuid", number=100_000))
        return {
            "samples": samples,
            "mean": statistics.mean(samples),
            "stddev": statistics.stdev(samples) if len(samples) > 1 else 0,
        }
    except AttributeError:
        return {"error": "UUIDv7 not available (requires Python 3.14+)"}


def benchmark_ulid():
    """Benchmark ULID generation with multiple samples."""
    try:
        import ulid
        samples = []
        for _ in range(5):
            samples.append(timeit.timeit("ulid.new()", setup="import ulid", number=100_000))
        return {
            "samples": samples,
            "mean": statistics.mean(samples),
            "stddev": statistics.stdev(samples) if len(samples) > 1 else 0,
        }
    except ImportError:
        return {"error": "ULID not available (pip install ulid-py)"}


def benchmark_nanoid():
    """Benchmark NanoID generation with multiple samples."""
    try:
        import nanoid
        samples = []
        for _ in range(5):
            samples.append(timeit.timeit("nanoid.generate()", setup="import nanoid", number=100_000))
        return {
            "samples": samples,
            "mean": statistics.mean(samples),
            "stddev": statistics.stdev(samples) if len(samples) > 1 else 0,
        }
    except ImportError:
        return {"error": "NanoID not available (pip install nanoid)"}


def benchmark_ksuid():
    """Benchmark KSUID generation with multiple samples."""
    try:
        import ksuid
        samples = []
        for _ in range(5):
            samples.append(timeit.timeit("ksuid.ksuid()", setup="import ksuid", number=100_000))
        return {
            "samples": samples,
            "mean": statistics.mean(samples),
            "stddev": statistics.stdev(samples) if len(samples) > 1 else 0,
        }
    except ImportError:
        return {"error": "KSUID not available (pip install ksuid)"}


def main():
    """Run comparison benchmarks with visual tables and statistical analysis."""
    print("=" * 60)
    print("SNID Comparison Benchmark")
    print("=" * 60)

    results = {
        "snid": benchmark_snid(),
        "uuidv7": benchmark_uuidv7(),
        "ulid": benchmark_ulid(),
        "nanoid": benchmark_nanoid(),
        "ksuid": benchmark_ksuid(),
    }

    # Print visual table
    print("\n┌────────────┬──────────────┬──────────────┬─────────────┐")
    print("│ Library    │ Time (s)     │ Ops/sec      │ StdDev      │")
    print("├────────────┼──────────────┼──────────────┼─────────────┤")
    
    for name, value in results.items():
        if isinstance(value, dict) and "error" in value:
            print(f"│ {name:10} │ {'ERROR':12} │ {'N/A':12} │ {'N/A':9} │")
        else:
            mean_time = value.get("mean", value) if isinstance(value, dict) else value
            stddev = value.get("stddev", 0) if isinstance(value, dict) else 0
            ops_per_sec = 100_000 / mean_time
            print(f"│ {name:10} │ {mean_time:12.4f} │ {ops_per_sec:12.0f} │ {stddev:9.4f} │")
    
    print("└────────────┴──────────────┴──────────────┴─────────────┘")

    # Calculate relative performance
    if isinstance(results["snid"], dict) and "mean" in results["snid"]:
        snid_time = results["snid"]["mean"]
        print(f"\n📈 Relative Performance (SNID = 1.0x):")
        print("┌────────────┬──────────────┐")
        print("│ Library    │ Relative     │")
        print("├────────────┼──────────────┤")
        for name, value in results.items():
            if isinstance(value, dict) and "mean" in value:
                relative = value["mean"] / snid_time
                print(f"│ {name:10} │ {relative:12.2f}x │")
        print("└────────────┴──────────────┘")

    # Statistical significance test (SNID vs others)
    if isinstance(results["snid"], dict) and "samples" in results["snid"]:
        print(f"\n📊 Statistical Significance (t-test vs SNID):")
        from math import sqrt
        snid_samples = results["snid"]["samples"]
        
        for name, value in results.items():
            if name == "snid" or isinstance(value, dict) and "samples" not in value:
                continue
            if isinstance(value, dict) and "samples" in value:
                other_samples = value["samples"]
                
                # Simple t-test approximation
                n1, n2 = len(snid_samples), len(other_samples)
                mean1, mean2 = statistics.mean(snid_samples), statistics.mean(other_samples)
                var1, var2 = statistics.variance(snid_samples) if n1 > 1 else 0, statistics.variance(other_samples) if n2 > 1 else 0
                
                pooled_std = sqrt(((n1-1)*var1 + (n2-1)*var2) / (n1+n2-2)) if n1+n2 > 2 else 0
                t_stat = (mean1 - mean2) / (pooled_std * sqrt(1/n1 + 1/n2)) if pooled_std > 0 else 0
                
                significance = "significant" if abs(t_stat) > 2.0 else "not significant"
                print(f"  {name:10} t={t_stat:6.2f} ({significance})")

    # Save results
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    output_file = RESULTS_DIR / f"comparison_{timestamp}.json"
    
    # Add metadata
    results_with_meta = {
        "timestamp": datetime.now().isoformat(),
        "iterations": 100_000,
        "samples_per_benchmark": 5,
        "results": results,
    }
    
    with open(output_file, "w") as f:
        json.dump(results_with_meta, f, indent=2)

    print(f"\n📁 Results saved to: {output_file}")
    return 0


if __name__ == "__main__":
    sys.exit(main())

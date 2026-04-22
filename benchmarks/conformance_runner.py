#!/usr/bin/env python3
"""
SNID Conformance Benchmark Runner
Runs conformance validation and measures performance across all languages.
"""

import subprocess
import json
import time
from datetime import datetime
from pathlib import Path

ROOT = Path(__file__).parent.parent
RESULTS_DIR = Path(__file__).parent / "results"


def generate_vectors():
    """Generate conformance vectors using Go."""
    print("\n▶ Generating conformance vectors with Go...")
    start = time.time()
    result = subprocess.run(
        ["go", "run", ".", "--out", "../../vectors.json"],
        cwd=ROOT / "conformance" / "cmd" / "generate_vectors",
        capture_output=True,
        text=True,
    )
    duration = time.time() - start
    return {
        "duration_seconds": round(duration, 2),
        "returncode": result.returncode,
        "stdout": result.stdout[-500:] if result.stdout else "",
        "stderr": result.stderr[-500:] if result.stderr else "",
    }


def validate_go():
    """Validate Go implementation against vectors."""
    print("\n▶ Validating Go implementation...")
    start = time.time()
    result = subprocess.run(
        ["go", "test", "./..."],
        cwd=ROOT / "go",
        capture_output=True,
        text=True,
    )
    duration = time.time() - start
    return {
        "duration_seconds": round(duration, 2),
        "returncode": result.returncode,
        "stdout": result.stdout[-500:] if result.stdout else "",
        "stderr": result.stderr[-500:] if result.stderr else "",
    }


def validate_rust():
    """Validate Rust implementation against vectors."""
    print("\n▶ Validating Rust implementation...")
    start = time.time()
    result = subprocess.run(
        ["cargo", "test"],
        cwd=ROOT / "rust",
        capture_output=True,
        text=True,
    )
    duration = time.time() - start
    return {
        "duration_seconds": round(duration, 2),
        "returncode": result.returncode,
        "stdout": result.stdout[-500:] if result.stdout else "",
        "stderr": result.stderr[-500:] if result.stderr else "",
    }


def validate_python():
    """Validate Python implementation against vectors."""
    print("\n▶ Validating Python implementation...")
    start = time.time()
    result = subprocess.run(
        ["python", "-m", "unittest", "discover", "-s", "tests"],
        cwd=ROOT / "python",
        capture_output=True,
        text=True,
    )
    duration = time.time() - start
    return {
        "duration_seconds": round(duration, 2),
        "returncode": result.returncode,
        "stdout": result.stdout[-500:] if result.stdout else "",
        "stderr": result.stderr[-500:] if result.stderr else "",
    }


def main():
    """Run conformance validation with performance metrics."""
    print("=" * 60)
    print("SNID Conformance Validation Benchmark")
    print("=" * 60)

    results = {
        "timestamp": datetime.now().isoformat(),
        "steps": {},
    }

    # Run conformance pipeline
    results["steps"]["generate_vectors"] = generate_vectors()
    results["steps"]["validate_go"] = validate_go()
    results["steps"]["validate_rust"] = validate_rust()
    results["steps"]["validate_python"] = validate_python()

    # Calculate total duration
    total_duration = sum(
        step["duration_seconds"] for step in results["steps"].values()
    )
    results["total_duration_seconds"] = round(total_duration, 2)

    # Save results
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    output_file = RESULTS_DIR / f"conformance_{timestamp}.json"
    with open(output_file, "w") as f:
        json.dump(results, f, indent=2)

    print(f"\n{'='*60}")
    print(f"✅ Conformance validation completed")
    print(f"📁 Results saved to: {output_file}")
    print(f"⏱️  Total duration: {total_duration:.2f}s")
    print(f"{'='*60}")

    # Print summary
    for step_name, step_data in results["steps"].items():
        status = "✅ PASSED" if step_data["returncode"] == 0 else "❌ FAILED"
        print(f"{step_name:20} {status} ({step_data['duration_seconds']:.2f}s)")

    # Overall status
    all_passed = all(step["returncode"] == 0 for step in results["steps"].values())
    if all_passed:
        print(f"\n✅ All implementations byte-identical")
        return 0
    else:
        print(f"\n❌ Conformance validation failed")
        return 1


if __name__ == "__main__":
    import sys

    sys.exit(main())

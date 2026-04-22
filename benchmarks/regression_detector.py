#!/usr/bin/env python3
"""
SNID Benchmark Regression Detector
Compares current benchmark results against baseline and flags regressions.
"""

import json
import os
import sys
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Any, Optional

RESULTS_DIR = Path(os.getenv("RESULTS_DIR", str(Path(__file__).parent / "results")))
DEFAULT_THRESHOLD = float(os.getenv("REGRESSION_THRESHOLD", "10"))


def load_results(filepath: Path) -> Optional[Dict[str, Any]]:
    """Load benchmark results from JSON file."""
    try:
        with open(filepath) as f:
            return json.load(f)
    except (json.JSONDecodeError, IOError):
        return None


def find_baseline() -> Optional[Path]:
    """Find the most recent baseline (last successful run before current)."""
    files = sorted(RESULTS_DIR.glob("performance_*.json"), key=lambda f: f.stat().st_mtime, reverse=True)
    if len(files) >= 2:
        return files[1]  # Second most recent is baseline
    return None


def detect_regressions(
    current: Dict[str, Any],
    baseline: Dict[str, Any],
    threshold: float = DEFAULT_THRESHOLD,
) -> List[Dict[str, Any]]:
    """Detect performance regressions compared to baseline."""
    regressions = []
    
    current_langs = current.get("languages", {})
    baseline_langs = baseline.get("languages", {})
    
    for lang in current_langs:
        if lang not in baseline_langs:
            continue
            
        current_benchmarks = current_langs[lang].get("benchmarks", {})
        baseline_benchmarks = baseline_langs[lang].get("benchmarks", {})
        
        for bench_name in current_benchmarks:
            if bench_name not in baseline_benchmarks:
                continue
            
            current_ns = current_benchmarks[bench_name].get("ns_per_op", 0)
            baseline_ns = baseline_benchmarks[bench_name].get("ns_per_op", 0)
            
            if baseline_ns == 0:
                continue
            
            percent_change = ((current_ns - baseline_ns) / baseline_ns) * 100
            
            if percent_change > threshold:
                regressions.append({
                    "language": lang,
                    "benchmark": bench_name,
                    "current_ns": current_ns,
                    "baseline_ns": baseline_ns,
                    "percent_change": percent_change,
                    "severity": "high" if percent_change > 20 else "medium",
                    "threshold": threshold,
                })
    
    return regressions


def generate_regression_report(
    regressions: List[Dict[str, Any]],
    current_file: Path,
    baseline_file: Path,
) -> Dict[str, Any]:
    """Generate a regression report."""
    return {
        "timestamp": datetime.now().isoformat(),
        "current_file": str(current_file),
        "baseline_file": str(baseline_file),
        "threshold": DEFAULT_THRESHOLD,
        "regressions_detected": len(regressions) > 0,
        "regressions": regressions,
        "summary": {
            "total": len(regressions),
            "high_severity": sum(1 for r in regressions if r["severity"] == "high"),
            "medium_severity": sum(1 for r in regressions if r["severity"] == "medium"),
        },
    }


def main():
    """Main entry point for regression detection."""
    if len(sys.argv) > 1:
        current_file = Path(sys.argv[1])
    else:
        # Find most recent results
        files = sorted(RESULTS_DIR.glob("performance_*.json"), key=lambda f: f.stat().st_mtime, reverse=True)
        if not files:
            print("❌ No performance results found")
            return 1
        current_file = files[0]
    
    print(f"Loading current results from: {current_file}")
    current = load_results(current_file)
    if not current:
        print(f"❌ Failed to load current results from {current_file}")
        return 1
    
    baseline_file = find_baseline()
    if not baseline_file:
        print("⚠️  No baseline found (need at least 2 runs)")
        return 0
    
    print(f"Loading baseline from: {baseline_file}")
    baseline = load_results(baseline_file)
    if not baseline:
        print(f"❌ Failed to load baseline from {baseline_file}")
        return 1
    
    threshold = DEFAULT_THRESHOLD
    print(f"Detecting regressions with threshold: {threshold}%")
    
    regressions = detect_regressions(current, baseline, threshold)
    
    if regressions:
        print(f"\n⚠️  Detected {len(regressions)} regression(s):\n")
        for reg in regressions:
            print(f"  [{reg['severity'].upper()}] {reg['language']}/{reg['benchmark']}")
            print(f"    Current: {reg['current_ns']:.2f} ns/op")
            print(f"    Baseline: {reg['baseline_ns']:.2f} ns/op")
            print(f"    Change: +{reg['percent_change']:.1f}%")
            print()
    else:
        print("✅ No regressions detected")
    
    # Save regression report
    report = generate_regression_report(regressions, current_file, baseline_file)
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    report_file = RESULTS_DIR / f"regression_{timestamp}.json"
    
    with open(report_file, "w") as f:
        json.dump(report, f, indent=2)
    
    print(f"📁 Regression report saved to: {report_file}")
    
    # Exit with error code if regressions detected
    return 1 if regressions else 0


if __name__ == "__main__":
    sys.exit(main())

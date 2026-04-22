#!/usr/bin/env python3
"""
SNID Benchmark Report Generator
Generates HTML reports with statistical analysis, trend charts, and regression highlights.
"""

import json
import os
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Any, Optional
from jinja2 import Template
import base64
from io import BytesIO
import matplotlib
matplotlib.use('Agg')  # Non-interactive backend
import matplotlib.pyplot as plt

RESULTS_DIR = Path(os.getenv("RESULTS_DIR", str(Path(__file__).parent / "results")))


def load_results(filepath: Path) -> Dict[str, Any]:
    """Load benchmark results from JSON file."""
    with open(filepath) as f:
        return json.load(f)


def get_historical_results() -> List[Dict[str, Any]]:
    """Load all historical benchmark results."""
    files = sorted(RESULTS_DIR.glob("performance_*.json"), key=lambda f: f.stat().st_mtime)
    results = []
    for f in files:
        try:
            results.append(load_results(f))
        except (json.JSONDecodeError, IOError):
            continue
    return results


def detect_regressions(current: Dict[str, Any], baseline: Dict[str, Any], threshold: float = 10.0) -> List[Dict[str, Any]]:
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
                })
    
    return regressions


def generate_trend_chart(historical: List[Dict[str, Any]], benchmark_name: str, language: str) -> str:
    """Generate a trend chart as base64-encoded image."""
    timestamps = []
    values = []
    
    for result in historical:
        timestamp = datetime.fromisoformat(result.get("timestamp", ""))
        timestamps.append(timestamp)
        
        lang_data = result.get("languages", {}).get(language, {})
        bench_data = lang_data.get("benchmarks", {}).get(benchmark_name, {})
        values.append(bench_data.get("ns_per_op", 0))
    
    if not values:
        return ""
    
    fig, ax = plt.subplots(figsize=(10, 4))
    ax.plot(timestamps, values, marker='o', linewidth=2)
    ax.set_title(f"{language} - {benchmark_name}")
    ax.set_ylabel("ns/op")
    ax.set_xlabel("Time")
    ax.grid(True, alpha=0.3)
    fig.autofmt_xdate()
    
    # Convert to base64
    buffer = BytesIO()
    fig.savefig(buffer, format='png', dpi=100, bbox_inches='tight')
    buffer.seek(0)
    image_base64 = base64.b64encode(buffer.read()).decode()
    plt.close(fig)
    
    return f"data:image/png;base64,{image_base64}"


def generate_html_report(results: Dict[str, Any], output_path: Path) -> None:
    """Generate HTML report from benchmark results."""
    historical = get_historical_results()
    
    # Find baseline (last successful run before current)
    baseline = historical[-2] if len(historical) > 1 else None
    regressions = detect_regressions(results, baseline) if baseline else []
    
    # Prepare data for template
    template_data = {
        "timestamp": results.get("timestamp"),
        "languages": results.get("languages", {}),
        "regressions": regressions,
        "has_regressions": len(regressions) > 0,
        "historical_count": len(historical),
    }
    
    # Generate trend charts for key benchmarks
    template_data["charts"] = []
    if len(historical) > 1:
        for lang in ["go", "rust", "python"]:
            lang_data = results.get("languages", {}).get(lang, {})
            for bench_name in list(lang_data.get("benchmarks", {}).keys())[:3]:  # Top 3 per language
                chart_data = generate_trend_chart(historical, bench_name, lang)
                if chart_data:
                    template_data["charts"].append({
                        "language": lang,
                        "benchmark": bench_name,
                        "image": chart_data,
                    })
    
    # Load and render template
    template_path = Path(__file__).parent / "templates" / "report.html"
    with open(template_path) as f:
        template = Template(f.read())
    
    html = template.render(**template_data)
    
    with open(output_path, "w") as f:
        f.write(html)


def main():
    """Generate HTML report for the latest benchmark results."""
    # Find latest performance results
    files = sorted(RESULTS_DIR.glob("performance_*.json"), key=lambda f: f.stat().st_mtime, reverse=True)
    
    if not files:
        print("No performance results found in:", RESULTS_DIR)
        return 1
    
    latest_file = files[0]
    print(f"Loading results from: {latest_file}")
    
    results = load_results(latest_file)
    
    # Generate HTML report
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    output_path = RESULTS_DIR / f"report_{timestamp}.html"
    
    print(f"Generating HTML report: {output_path}")
    generate_html_report(results, output_path)
    
    print(f"✅ Report generated: {output_path}")
    
    # Print regression summary
    historical = get_historical_results()
    if len(historical) > 1:
        baseline = historical[-2]
        regressions = detect_regressions(results, baseline)
        if regressions:
            print(f"\n⚠️  Detected {len(regressions)} regression(s):")
            for reg in regressions:
                print(f"  - {reg['language']}/{reg['benchmark']}: +{reg['percent_change']:.1f}%")
        else:
            print("\n✅ No regressions detected")
    
    return 0


if __name__ == "__main__":
    import sys
    sys.exit(main())

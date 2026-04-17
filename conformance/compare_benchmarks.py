from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path


GO_LINE = re.compile(
    r"^(?P<name>Benchmark\S+)(?:-\d+)?\s+\d+\s+(?P<ns>[0-9.]+)\s+ns/op(?:\s+(?P<bop>[0-9.]+)\s+B/op\s+(?P<allocs>[0-9.]+)\s+allocs/op)?$"
)


def load_go(path: Path) -> dict[str, dict[str, float]]:
    benchmarks: dict[str, dict[str, float]] = {}
    for raw_line in path.read_text().splitlines():
        line = raw_line.strip()
        match = GO_LINE.match(line)
        if not match:
            continue
        benchmarks[match.group("name")] = {
            "ns_per_op": float(match.group("ns")),
            "b_per_op": float(match.group("bop") or 0.0),
            "allocs_per_op": float(match.group("allocs") or 0.0),
        }
    if not benchmarks:
        raise ValueError(f"no Go benchmark lines found in {path}")
    return benchmarks


def load_rust(path: Path) -> dict[str, dict[str, float]]:
    payload = json.loads(path.read_text())
    benchmarks = payload.get("benchmarks", {})
    if not benchmarks:
        raise ValueError(f"no Rust benchmarks found in {path}")
    return {
        name: {"ns_per_op": float(metrics["ns_per_op"])}
        for name, metrics in benchmarks.items()
    }


def compare(
    baseline: dict[str, dict[str, float]],
    current: dict[str, dict[str, float]],
    threshold_pct: float,
    metric: str,
) -> tuple[list[str], list[str]]:
    failures: list[str] = []
    summaries: list[str] = []
    threshold = threshold_pct / 100.0

    for name, base_metrics in baseline.items():
        if name not in current:
            failures.append(f"missing benchmark in current run: {name}")
            continue
        base_value = base_metrics[metric]
        curr_value = current[name][metric]
        if base_value == 0:
            summaries.append(f"{name}: baseline {metric}=0, skipped comparison")
            continue
        delta = (curr_value - base_value) / base_value
        summaries.append(
            f"{name}: baseline={base_value:.3f} current={curr_value:.3f} delta={delta * 100:.2f}%"
        )
        if delta > threshold:
            failures.append(
                f"{name} regressed by {delta * 100:.2f}% on {metric} (threshold {threshold_pct:.2f}%)"
            )
    return failures, summaries


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--language", choices=["go", "rust"], required=True)
    parser.add_argument("--baseline", type=Path, required=True)
    parser.add_argument("--current", type=Path, required=True)
    parser.add_argument("--threshold-pct", type=float, default=5.0)
    args = parser.parse_args()

    if args.language == "go":
        baseline = load_go(args.baseline)
        current = load_go(args.current)
    else:
        baseline = load_rust(args.baseline)
        current = load_rust(args.current)

    failures, summaries = compare(baseline, current, args.threshold_pct, "ns_per_op")
    for line in summaries:
        print(line)
    if failures:
        for line in failures:
            print(line, file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${1:-$ROOT/conformance/artifacts/go}"
COUNT="${BENCH_COUNT:-6}"
BENCH_REGEX="${BENCH_REGEX:-BenchmarkSNID(NewFastParallel|StringMatterParallel|ToTensorWords|ToLLMFormat|DeterministicIngestID)|BenchmarkLIDVerifyParallel|BenchmarkBIDWireFormat|BenchmarkEIDNew}"
CACHE_DIR="$ROOT/conformance/cmd/project_go/.gocache"
TMP_DIR="$ROOT/conformance/cmd/project_go/.gotmp"

mkdir -p "$OUT_DIR"
mkdir -p "$CACHE_DIR" "$TMP_DIR"

cd "$ROOT/go"
GOCACHE="$CACHE_DIR" GOTMPDIR="$TMP_DIR" go test -run '^$' -bench "$BENCH_REGEX" -benchmem -count "$COUNT" ./... | tee "$OUT_DIR/bench.txt"

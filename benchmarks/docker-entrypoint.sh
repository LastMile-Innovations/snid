#!/bin/bash
set -e

# Docker entrypoint for SNID Benchmarking Platform
# Supports both web dashboard and CLI modes based on BENCH_MODE

echo "🚀 SNID Benchmarking Platform"
echo "Mode: ${BENCH_MODE:-web}"
echo "Results Dir: ${RESULTS_DIR:-/app/results}"
echo "Port: ${PORT:-8080}"

# Ensure results directory exists
mkdir -p "${RESULTS_DIR:-/app/results}"

if [ "${BENCH_MODE:-web}" = "cli" ]; then
    # CLI mode: Run benchmarks directly
    echo "Running in CLI mode..."
    exec python3 /app/benchmarks/runner.py ${BENCH_SUITES:-all}
else
    # Web mode: Start FastAPI dashboard
    echo "Starting web dashboard..."
    exec uvicorn benchmarks.app:app --host 0.0.0.0 --port ${PORT:-8080}
fi

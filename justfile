# SNID — Modern polyglot identifier protocol
set shell := ["bash", "-c"]

# Default recipe
default:
    @just --list

# =============================================================================
# Development
# =============================================================================

install:
    cd go && go mod download
    cd rust && cargo fetch
    cd python && maturin develop

# =============================================================================
# Testing
# =============================================================================

test:
    @echo "Running all tests..."
    just test-go
    just test-rust
    just test-python
    @echo "✅ All tests passed"

test-go:
    @echo "Testing Go..."
    cd go && go test ./...

test-rust:
    @echo "Testing Rust..."
    cd rust && cargo test
    cd rust && cargo test --test property_tests

test-python:
    @echo "Testing Python..."
    cd python && python3 -m pytest tests/
    cd python && python3 -m pytest tests/test_properties.py

# =============================================================================
# Conformance (The Most Important Command)
# =============================================================================

conformance:
    @echo "Regenerating vectors with Go (authoritative)..."
    cd conformance/cmd/generate_vectors && go run . --out ../../vectors.json
    @echo "Validating Rust..."
    cd rust && cargo test
    @echo "Validating Python..."
    cd python && python3 -m unittest discover -s tests
    @echo "✅ All implementations byte-identical"

# =============================================================================
# Benchmarking (2026 Best Practices)
# =============================================================================

bench:
    @echo "Running all benchmarks..."
    just bench-go
    just bench-rust
    just bench-python

bench-go:
    cd go && go test -bench=. -benchmem -count=5

bench-rust:
    cd rust && cargo bench -- --save-baseline main

bench-python:
    cd python && python -m pytest tests/test_bench.py --benchmark-only --benchmark-json=../benchmarks/results/python_bench.json

bench-all:
    just bench
    just bench-comparison
    just bench-llm

bench-comparison:
    @echo "Running comparison benchmarks (vs UUIDv7, ULID, NanoID)..."
    python benchmarks/comparison_benchmark.py

bench-llm:
    @echo "Running LLM token efficiency benchmarks..."
    python benchmarks/llm_token_benchmark.py

# =============================================================================
# Railway Commands (Phase 2 - Implemented)
# =============================================================================

railway-deploy:
    @echo "Building and deploying to Railway..."
    cd benchmarks && docker build -f Dockerfile -t registry.railway.app/$(railway service id):latest .
    railway up

railway-run:
    @echo "Running one-off benchmark on Railway..."
    railway run --env BENCH_MODE=cli --env BENCH_SUITES=all python benchmarks/runner.py all

railway-logs:
    @echo "Viewing Railway service logs..."
    railway logs

railway-dashboard:
    @echo "Opening Railway dashboard..."
    railway open

# =============================================================================
# Formatting
# =============================================================================

fmt:
    @echo "Formatting all code..."
    cd go && gofmt -w $(find . -name '*.go' -not -path './.gomodcache/*')
    cd rust && cargo fmt
    cd python && ruff format .
    @echo "✅ All code formatted"

fmt-check:
    @echo "Checking formatting..."
    cd go && gofmt -l .
    cd rust && cargo fmt -- --check
    cd python && ruff format --check .

# =============================================================================
# Linting
# =============================================================================

lint:
    @echo "Linting all code..."
    cd go && golangci-lint run || true
    cd rust && cargo clippy
    cd python && ruff check .
    @echo "✅ Linting complete"

# =============================================================================
# Release Preparation
# =============================================================================

release-prep:
    @echo "⚠️  Manual step: Bump versions in lockstep"
    @echo "   - go/go.mod"
    @echo "   - rust/Cargo.toml"
    @echo "   - python/pyproject.toml"
    @echo "   - Update CHANGELOG.md"
    @echo "   - Regenerate conformance vectors"

# =============================================================================
# Documentation
# =============================================================================

docs-serve:
    @echo "Serve documentation (requires mdbook or similar)"
    @echo "Install: cargo install mdbook"
    @echo "Run: mdbook serve docs"

docs-build:
    @echo "Build documentation"
    @echo "Install: cargo install mdbook"
    @echo "Run: mdbook build docs"

# =============================================================================
# Clean
# =============================================================================

clean:
    @echo "Cleaning build artifacts..."
    cd go && go clean -cache -modcache -testcache
    cd rust && cargo clean
    cd python && rm -rf .venv __pycache__ .pytest_cache target
    @echo "✅ Clean complete"

# =============================================================================
# CI Helpers
# =============================================================================

ci: fmt-check lint test conformance
    @echo "✅ CI checks passed"

# =============================================================================
# Full Quality Gate (Run Before PR)
# =============================================================================

check:
    just fmt
    just lint
    just test
    just conformance
    @echo "✅ All quality gates passed"

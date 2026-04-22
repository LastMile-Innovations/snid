# Contributing to SNID

Thank you for your interest in contributing to SNID! This document provides guidelines for contributing to the project.

## Development Setup

### Prerequisites

- Go 1.24+
- Rust 1.70+ (with Cargo)
- Python 3.10+ (with maturin for Python bindings)
- Git
- [just](https://github.com/casey/just) - Command runner (optional but recommended)
- [mise](https://mise.jdx.dev/) - Dev environment manager (optional but recommended)
- [pre-commit](https://pre-commit.com/) - Git hooks (optional but recommended)

### Quick Setup (Recommended)

```bash
# Clone repository
git clone https://github.com/LastMile-Innovations/snid.git
cd snid

# Install development tools
cargo install just
curl https://mise.run | sh
pip install pre-commit

# Install dependencies
just install

# Set up pre-commit hooks
pre-commit install

# Verify setup
just test
just conformance
```

### Platform-Specific Setup

#### macOS

```bash
# Install mise using Homebrew
brew install mise

# Install Go
brew install go

# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env

# Install Python
brew install python@3.10

# Install just
cargo install just

# Install pre-commit
pip3 install pre-commit

# Set up repository
cd snid
mise install
just install
pre-commit install
```

#### Linux (Ubuntu/Debian)

```bash
# Install mise
curl https://mise.run | sh

# Install Go
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env

# Install Python
sudo apt update
sudo apt install python3.10 python3-pip python3.10-venv

# Install just
cargo install just

# Install pre-commit
pip3 install pre-commit

# Set up repository
cd snid
mise install
just install
pre-commit install
```

#### Windows

```bash
# Install mise using winget
winget install jdx.mise

# Install Go from https://go.dev/dl/

# Install Rust from https://rustup.rs/

# Install Python from https://python.org/

# Install just
cargo install just

# Install pre-commit
pip install pre-commit

# Set up repository
cd snid
mise install
just install
pre-commit install
```

### Repository Structure

```
snid/
├── go/          # Go reference implementation
├── rust/        # Rust core library
├── python/      # Python bindings (PyO3)
├── conformance/ # Cross-language conformance tests
├── docs/        # Protocol specification and guides
├── examples/    # Runnable code examples
├── cli/         # Unified CLI tool (coming soon)
└── justfile     # Unified command runner
```

### Go Development

```bash
cd go
go mod tidy
go test ./...
go test -bench=. -benchmem
```

Or using just:
```bash
just test-go
just bench-go
```

### Rust Development

```bash
cd rust
cargo test
cargo test --release
cargo bench
cargo clippy
```

Or using just:
```bash
just test-rust
just bench-rust
```

### Python Development

```bash
cd python
maturin develop
python -m pytest tests/
python bench_batch.py
```

Or using just:
```bash
just test-python
just bench-python
```

## Conformance Testing

The conformance suite is the release gate for all implementations. Before submitting changes:

### Using just (Recommended)

```bash
just conformance
```

This automatically:
1. Generates vectors with Go
2. Validates with Rust
3. Validates with Python
4. Reports conformance status

### Manual Steps

1. Generate new test vectors with Go:
```bash
cd conformance/cmd/generate_vectors
go run . --out ../../vectors.json
```

2. Validate with Rust:
```bash
cd rust
cargo test
```

3. Validate with Python:
```bash
cd python
python -m unittest discover -s tests
```

All three implementations must pass the conformance suite before changes are merged.

### When to Regenerate Vectors

Regenerate conformance vectors when:
- Protocol changes are made (byte layout, wire format)
- New identifier families are added
- New boundary projections are implemented
- Encoding/decoding logic changes

**Never commit `conformance/vectors.json` without regenerating from Go.**

## Code Style

### Using Pre-Commit Hooks (Recommended)

Pre-commit hooks automatically format and lint your code before each commit:

```bash
# Install pre-commit
pip install pre-commit

# Set up hooks
pre-commit install

# Run hooks manually
pre-commit run --all-files
```

The pre-commit configuration includes:
- Go: gofmt, golangci-lint
- Rust: cargo fmt, cargo clippy
- Python: ruff format, ruff check
- General: trailing whitespace, YAML/TOML validation, markdown lint

### Manual Formatting

#### Go

```bash
# Format
gofmt -w .

# Lint
golangci-lint run

# Or using just
just fmt
just lint
```

- Follow standard Go formatting (`gofmt`)
- Use `golangci-lint` for linting
- Add godoc comments to exported functions

#### Rust

```bash
# Format
cargo fmt

# Lint
cargo clippy

# Or using just
just fmt
just lint
```

- Use `cargo fmt` for formatting
- Use `cargo clippy` for linting
- Add rustdoc comments to public APIs

#### Python

```bash
# Format
ruff format .

# Lint
ruff check .

# Or using just
just fmt
just lint
```

- Follow PEP 8 style guide
- Use `ruff` for linting
- Add docstrings to public functions

## Submitting Changes

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Run formatting: `just fmt`
5. Run linting: `just lint`
6. Run tests: `just test`
7. Run conformance: `just conformance`
8. Commit your changes (`git commit -am 'Add feature'`)
9. Push to the branch (`git push origin feature/my-feature`)
10. Create a Pull Request

## Pull Request Guidelines

- Use the PR template when creating your PR
- Describe your changes in the PR description
- Reference any related issues (e.g., "Fixes #123")
- Ensure all tests pass (CI will check this)
- Update documentation if needed
- Add tests for new functionality
- Run `just fmt` and `just lint` before pushing
- Run `just conformance` for any protocol changes

## Testing Strategy

### Unit Tests

Each implementation should have comprehensive unit tests:
- Go: `go test ./...`
- Rust: `cargo test`
- Python: `python -m pytest tests/`

### Conformance Tests

Conformance tests ensure byte-identical behavior across implementations:
- Run with: `just conformance`
- Must pass before merging
- Regenerate vectors for protocol changes

### Benchmark Tests

Performance benchmarks for optimization validation:
- Go: `go test -bench=. -benchmem`
- Rust: `cargo bench`
- Python: `python bench_batch.py`

### Integration Tests

Integration tests for:
- Database storage contracts
- API compatibility
- Wire format roundtrips

## Code Review Guidelines

### For Reviewers

- Check for conformance test compliance
- Verify documentation updates
- Ensure code style consistency
- Review performance impact for optimizations
- Check for security considerations
- Validate protocol changes against SPEC.md

### For Contributors

- Address all review comments
- Update tests for new functionality
- Keep PRs focused and small
- Respond to review feedback promptly
- Squash commits before final merge

## Release Process

### Version Bumping

When preparing a release:

1. Update version in all three implementations:
   - `go/go.mod`
   - `rust/Cargo.toml`
   - `python/pyproject.toml`

2. Update CHANGELOG.md with notable changes

3. Regenerate conformance vectors if protocol changed

4. Run full test suite: `just test && just conformance`

5. Create release commit with version bump

### Publishing

Publish to all registries in coordinated release:
- Go: `go publish` (or manual)
- Rust: `cargo publish`
- Python: `maturin publish`

### Release Checklist

- [ ] All tests pass
- [ ] Conformance suite passes
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version bumped in all implementations
- [ ] Release notes prepared
- [ ] Security review if applicable

## Protocol Changes

Protocol changes require:

1. Update to `docs/SPEC.md`
2. Version bump in all three implementations
3. Update conformance vectors
4. Migration notes in `CHANGELOG.md`

Protocol changes must remain additive unless accompanied by a version bump.

## Implementation Changes

Implementation changes (optimizations, refactoring) that don't affect the protocol:

- Must pass conformance tests
- Should include benchmarks for performance changes
- Should maintain backward compatibility

## Questions?

Open an issue on GitHub for questions or discussion.

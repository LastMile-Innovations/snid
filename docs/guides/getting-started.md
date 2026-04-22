# Installation Guide

Detailed installation instructions for SNID across all supported languages.

## Prerequisites

### System Requirements

- **Go**: 1.24 or later
- **Rust**: 1.70 or later
- **Python**: 3.10 or later

### Development Tools (Optional)

For contributing to SNID:

- [just](https://github.com/casey/just) - Command runner
- [mise](https://mise.jdx.dev/) - Dev environment manager
- [pre-commit](https://pre-commit.com/) - Git hooks

## Installation

### Go

#### Using go get

```bash
go get github.com/neighbor/snid
```

#### Manual installation

```bash
git clone https://github.com/neighbor/snid.git
cd snid/go
go mod download
go test ./...
```

### Rust

#### Using cargo

```bash
cargo add snid
```

#### Manual installation

```bash
git clone https://github.com/neighbor/snid.git
cd snid/rust
cargo build
cargo test
```

### Python

#### Using pip

```bash
pip install snid
```

#### With data science dependencies

```bash
pip install snid[data]
```

This installs optional dependencies for NumPy, PyArrow, and Polars integration.

#### Manual installation

```bash
git clone https://github.com/neighbor/snid.git
cd snid/python
maturin develop
python -m pytest tests/
```

## Development Setup

### Install development tools

```bash
# Install just
cargo install just

# Install mise
curl https://mise.run | sh

# Install pre-commit
pip install pre-commit
```

### Set up the repository

```bash
git clone https://github.com/neighbor/snid.git
cd snid

# Install language-specific dependencies
just install

# Set up pre-commit hooks
pre-commit install
```

### Verify installation

```bash
# Run all tests
just test

# Run conformance suite
just conformance

# Run benchmarks
just bench
```

## Platform-Specific Notes

### macOS

```bash
# Install mise using Homebrew
brew install mise

# Ensure Go is installed
brew install go

# Ensure Rust is installed
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Ensure Python is installed
brew install python@3.10
```

### Linux (Ubuntu/Debian)

```bash
# Install mise
curl https://mise.run | sh

# Install Go
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz

# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Install Python
sudo apt update
sudo apt install python3.10 python3-pip
```

### Windows

```bash
# Install mise using winget
winget install jdx.mise

# Install Go from https://go.dev/dl/

# Install Rust from https://rustup.rs/

# Install Python from https://python.org/
```

## Troubleshooting

### Go: module not found

```bash
cd go
go mod download
go mod tidy
```

### Rust: cargo not found

```bash
# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env
```

### Python: maturin not found

```bash
pip install maturin
cd python
maturin develop
```

### Pre-commit hooks not running

```bash
# Reinstall pre-commit
pip install --upgrade pre-commit
pre-commit install --force
```

## Next Steps

- [Quick Start Guide](quick-start.md) - Your first SNID
- [Basic Usage](basic-usage.md) - Common patterns
- [Contributing](../../CONTRIBUTING.md) - Development guidelines

# Publishing Guide

This guide covers the complete process for publishing SNID packages to public registries (crates.io, PyPI, proxy.golang.org) via GitHub Actions.

## Prerequisites

Before publishing, ensure you have:

1. **Required tooling installed**
   - **Rust and Cargo**: Required for Rust package publishing
     ```bash
     # On Linux and macOS
     curl https://sh.rustup.rs -sSf | sh

     # On Windows
     # Download and run rustup-init.exe from https://rustup.rs/
     ```
   - **Go**: Required for Go module verification (install from https://go.dev/dl/)
   - **Python and maturin**: Required for Python package publishing
     ```bash
     pip install maturin
     ```

2. **API tokens configured** in GitHub Secrets
   - `CRATES_IO_TOKEN` - See [deploy/TOKEN_SETUP.md](deploy/TOKEN_SETUP.md#cratesio-token-setup)
   - `PYPI_API_TOKEN` - See [deploy/TOKEN_SETUP.md](deploy/TOKEN_SETUP.md#pypi-token-setup)
   - Quick reference: [deploy/GITHUB_SECRETS.md](deploy/GITHUB_SECRETS.md)

3. **Repository permissions**
   - You must be a repository admin or have write access
   - GitHub Actions must be enabled for the repository

4. **Registry accounts**
   - crates.io account with ownership of the `snid` crate
   - PyPI account with the `snid` package reserved

5. **Clean working state**
   - All tests passing: `just test`
   - Conformance suite passing: `just conformance`
   - No uncommitted changes

## Cargo Basics (For New Users)

If you're new to Cargo, here's a quick reference for the essential commands:

### Creating a New Package

```bash
# Create a binary program
cargo new hello_world

# Create a library
cargo new hello_world --lib
```

### Building and Running

```bash
# Build the package
cargo build

# Build for release (optimized)
cargo build --release

# Run the binary
cargo run

# Run tests
cargo test

# Run benchmarks
cargo bench
```

### Publishing Commands

```bash
# Login to crates.io (prompts for API token)
cargo login

# Publish to crates.io
cargo publish

# Publish with dry-run (test without actually publishing)
cargo publish --dry-run

# Yank a published version (deprecate it)
cargo yank --vers 0.2.1 snid
```

### Cargo.toml Structure

The `Cargo.toml` file is the manifest for your package:

```toml
[package]
name = "snid"
version = "0.2.0"
edition = "2021"
rust-version = "1.95"
license = "MIT OR Apache-2.0"
description = "Polyglot sortable identifier protocol"
repository = "https://github.com/LastMile-Innovations/snid"

[dependencies]
# List your dependencies here

[dev-dependencies]
# Dependencies for testing and benchmarking
```

### Useful Commands for SNID

```bash
# Check for linting issues
cargo clippy

# Format code
cargo fmt

# Update dependencies
cargo update

# Check if package can be published
cargo publish --dry-run
```

## Version Management

### Current Version Status

- **Rust**: `rust/Cargo.toml` - version = "0.2.1"
- **Python**: `python/pyproject.toml` - version = "0.2.1"
- **Go**: `go/go.mod` - No version field (uses git tags)

### Version Bumping Process

When preparing a new release, update versions in this order:

1. **Rust** - Edit `rust/Cargo.toml`:
   ```toml
   [package]
   version = "0.2.1"  # Update this
   ```

2. **Python** - Edit `python/pyproject.toml`:
   ```toml
   [project]
   version = "0.2.1"  # Update this
   ```

3. **Go** - No version file needed. Go uses git tags directly.

4. **Commit the changes**:
   ```bash
   git add rust/Cargo.toml python/pyproject.toml
   git commit -m "Bump version to 0.2.1"
   ```

5. **Push to main**:
   ```bash
   git push origin main
   ```

## Release Process

### Step 1: Run Full Test Suite

Before releasing, ensure everything passes:

```bash
# Run all tests
just test

# Run conformance suite (critical)
just conformance

# Run benchmarks (optional but recommended)
just bench
```

### Step 2: Update CHANGELOG.md

Document the changes in `CHANGELOG.md`:

```markdown
## [0.2.1] - 2026-04-22

### Added
- New feature A
- New feature B

### Changed
- Performance improvement C
- API change D

### Fixed
- Bug fix E
```

Commit the changelog:

```bash
git add CHANGELOG.md
git commit -m "Update CHANGELOG for 0.2.1"
git push origin main
```

### Step 3: Create and Push Git Tag

Create an annotated tag for the release:

```bash
# Create annotated tag
git tag -a v0.2.1 -m "Release v0.2.1"

# Push tag to trigger release workflow
git push origin v0.2.1
```

**Important:**
- Tags must follow semantic versioning: `vX.Y.Z`
- Use annotated tags (`-a`) for proper release notes
- The tag format must match the version in Cargo.toml and pyproject.toml

### Step 4: Monitor Release Workflow

1. Go to the "Actions" tab in GitHub
2. Select the "Release" workflow run (triggered by the tag)
3. Monitor the following jobs:
   - `release` - Creates GitHub release
   - `publish-go` - Verifies Go module accessibility
   - `publish-rust` - Publishes to crates.io
   - `publish-python` - Publishes to PyPI

### Step 5: Verify Publication

After the workflow completes, verify each package:

**Go:**
```bash
# Verify Go module is accessible
GOPROXY=proxy.golang.org go list -m github.com/LastMile-Innovations/snid@v0.2.1
```

**Rust:**
```bash
# Check crates.io
cargo search snid

# Or visit https://crates.io/crates/snid
```

**Python:**
```bash
# Check PyPI
pip index versions snid

# Or visit https://pypi.org/project/snid/
```

## Release Workflow Details

The `.github/workflows/release.yml` workflow runs on git tags matching `v*`:

### Jobs

1. **release** - Creates GitHub release with auto-generated notes
2. **publish-go** - Verifies Go module accessibility via proxy.golang.org
3. **publish-rust** - Publishes Rust crate to crates.io
4. **publish-python** - Builds wheels and publishes to PyPI

### Go Module Publishing

Go modules are unique - they don't have a "publish" command. The Go module proxy (proxy.golang.org) automatically fetches and serves modules when:
- Repository is public
- Git tags follow semantic versioning (vX.Y.Z)
- go.mod has correct module path

The workflow verifies the module is accessible after the tag is pushed.

## Rollback Procedure

If a release has issues, you can roll back:

### 1. Yank Rust Crate (crates.io)

```bash
cargo yank --vers 0.2.1 snid
```

### 2. Yank Python Package (PyPI)

PyPI doesn't support yanking, but you can:
- Release a new version with fixes
- Mark the old version as "yanked" in the project metadata

### 3. Go Module

Go modules cannot be deleted from the proxy, but you can:
- Release a new version with fixes
- Users can pin to specific versions in go.mod

### 4. Delete GitHub Release

1. Go to the "Releases" page in GitHub
2. Delete the release (this doesn't delete the tag)
3. Delete the tag locally and remotely:
   ```bash
   git tag -d v0.2.1
   git push origin :refs/tags/v0.2.1
   ```

## Post-Release Tasks

After a successful release:

1. **Update documentation** if API changes occurred
2. **Announce the release** (blog post, social media, etc.)
3. **Close related GitHub issues** with the release tag
4. **Update examples** if new features were added
5. **Plan next release** based on open issues and roadmap

## Troubleshooting

### Workflow Fails on Go Verification

**Error:** Go module not accessible via proxy.golang.org

**Solution:**
- Wait a few minutes for the proxy to index the new tag
- Verify the repository is public
- Check the tag format matches semantic versioning

### Workflow Fails on Rust Publish

**Error:** `cargo publish` fails

**Solution:**
- Verify `CRATES_IO_TOKEN` secret is set correctly
- Check you have ownership of the `snid` crate on crates.io
- Ensure version in Cargo.toml hasn't been published already

### Workflow Fails on Python Publish

**Error:** `maturin publish` fails

**Solution:**
- Verify `PYPI_API_TOKEN` secret is set correctly
- Check 2FA is enabled on your PyPI account
- Ensure version in pyproject.toml hasn't been published already
- Verify the package name `snid` is reserved on PyPI

### Tag Already Exists

**Error:** `tag 'v0.2.1' already exists`

**Solution:**
- Delete the existing tag if it was a mistake:
  ```bash
  git tag -d v0.2.1
  git push origin :refs/tags/v0.2.1
  ```
- Or use a different version number

## Manual Publishing (Fallback)

If the GitHub Actions workflow fails, you can publish manually:

### Rust

```bash
cd rust
cargo login
# Enter your crates.io token
cargo publish
```

### Python

```bash
cd python
maturin build --release
maturin publish --username __token__
# Enter your PyPI token when prompted
```

### Go

Go modules are published automatically when tags are pushed. No manual action needed.

## Related Documentation

- [deploy/TOKEN_SETUP.md](deploy/TOKEN_SETUP.md) - API token setup instructions
- [deploy/GITHUB_SECRETS.md](deploy/GITHUB_SECRETS.md) - GitHub secrets configuration
- [deploy/CI.md](deploy/CI.md) - CI/CD integration guide
- [CHANGELOG.md](CHANGELOG.md) - Release changelog
- [CONTRIBUTING.md](CONTRIBUTING.md) - Development guidelines

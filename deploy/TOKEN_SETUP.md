# API Token Setup Guide

This guide explains how to set up the required API tokens for publishing SNID packages to public registries via GitHub Actions.

## Required GitHub Secrets

The release workflow requires the following secrets to be configured in your GitHub repository:

- `CRATES_IO_TOKEN` - For publishing Rust crate to crates.io
- `PYPI_API_TOKEN` - For publishing Python package to PyPI

## crates.io Token Setup

### 1. Create a crates.io Account

1. Go to [https://crates.io](https://crates.io)
2. Click "Log In" and sign up with GitHub OAuth
3. Verify your email address

### 2. Claim the Package Name

1. Go to [https://crates.io/new](https://crates.io/new)
2. Enter package name: `snid`
3. Click "Check" to verify the name is available
4. If available, click "Create" to claim it
5. Note: You'll need to be an owner of the `snid` crate to publish

### 3. Generate API Token

1. Go to [https://crates.io/me](https://crates.io/me)
2. Click "API Tokens" in the left sidebar
3. Click "New API Token"
4. Enter a token name (e.g., "GitHub Actions - SNID")
5. Select scope: "Publish new crates" and "Update existing crates"
6. Click "Create"
7. **Copy the token immediately** - you won't see it again

### 4. Add to GitHub Secrets

1. Go to your GitHub repository: `https://github.com/LastMile-Innovations/snid`
2. Click "Settings" tab
3. Click "Secrets and variables" → "Actions"
4. Click "New repository secret"
5. Name: `CRATES_IO_TOKEN`
6. Value: Paste the token from step 3
7. Click "Add secret"

## PyPI Token Setup

### 1. Create a PyPI Account

1. Go to [https://pypi.org](https://pypi.org)
2. Click "Register" in the top right
3. Fill in username, email, and password
4. Verify your email address

### 2. Enable 2FA (Required for API Tokens)

1. Go to [https://pypi.org/manage/account](https://pypi.org/manage/account)
2. Scroll to "Two-factor authentication"
3. Click "Enable 2FA"
4. Follow the setup process (TOTP app recommended)
5. **Note**: PyPI requires 2FA to be enabled before you can create API tokens

### 3. Reserve the Package Name (Optional but Recommended)

1. Go to [https://pypi.org/manage/projects/](https://pypi.org/manage/projects/)
2. Click "Create project"
3. Enter name: `snid`
4. Click "Create"
5. This reserves the name and prevents others from claiming it

### 4. Generate API Token

1. Go to [https://pypi.org/manage/account/token/](https://pypi.org/manage/account/token/)
2. Click "Add API token"
3. Enter token name: "GitHub Actions - SNID"
4. Scope: Select "Entire account" (for publishing)
5. Click "Add token"
6. **Copy the token immediately** - it starts with `pypi-` and won't be shown again

### 5. Add to GitHub Secrets

1. Go to your GitHub repository: `https://github.com/LastMile-Innovations/snid`
2. Click "Settings" tab
3. Click "Secrets and variables" → "Actions"
4. Click "New repository secret"
5. Name: `PYPI_API_TOKEN`
6. Value: Paste the token from step 4
7. Click "Add secret"

## Test Publishing Locally (Optional)

### Test Rust Publishing

```bash
cd rust
cargo login
# Enter your crates.io token when prompted
cargo publish --dry-run
```

### Test Python Publishing

```bash
cd python
maturin build --release
maturin publish --username __token__ --password <your-token>
```

## Security Best Practices

- **Never commit tokens to git** - Always use GitHub Secrets
- **Use scoped tokens** - Only grant necessary permissions
- **Rotate tokens regularly** - Revoke and regenerate tokens periodically
- **Monitor usage** - Check token usage logs on crates.io and PyPI
- **Limit access** - Only give repository maintainers access to secrets

## Troubleshooting

### crates.io Issues

**Error: "you don't have permission to publish this crate"**
- Ensure you're listed as an owner of the `snid` crate on crates.io
- Contact existing owners to add you: `cargo owner --add your-username snid`

**Error: "crate name already taken"**
- The package name is already claimed by another user
- Consider a different name or contact the current owner

### PyPI Issues

**Error: "403 Forbidden"**
- Verify 2FA is enabled on your PyPI account
- Ensure the token has the correct scope
- Check that the token hasn't been revoked

**Error: "Project name already exists"**
- The package name is already taken on PyPI
- You may need to use a different name or contact the current maintainer

### GitHub Secrets Issues

**Secrets not available in workflow**
- Ensure secrets are added to the correct repository
- Check that the workflow has `contents: write` permission
- Verify the secret names match exactly: `CRATES_IO_TOKEN`, `PYPI_API_TOKEN`

## Additional Resources

- [crates.io API Tokens Documentation](https://crates.io/settings/tokens)
- [PyPI API Tokens Documentation](https://pypi.org/help/#apitoken)
- [GitHub Actions Secrets Documentation](https://docs.github.com/en/actions/security-guides/encrypted-secrets)

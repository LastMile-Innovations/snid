# GitHub Secrets Configuration

This guide lists the required GitHub secrets for the SNID release workflow and provides quick setup references.

## Required Secrets

The following secrets must be configured in your GitHub repository for the release workflow to function:

| Secret Name | Purpose | Registry | Setup Guide |
|-------------|---------|----------|-------------|
| `CRATES_IO_TOKEN` | Publish Rust crate to crates.io | crates.io | [TOKEN_SETUP.md](TOKEN_SETUP.md#cratesio-token-setup) |
| `PYPI_API_TOKEN` | Publish Python package to PyPI | PyPI | [TOKEN_SETUP.md](TOKEN_SETUP.md#pypi-token-setup) |

## Quick Setup

### 1. Navigate to Repository Settings

1. Go to `https://github.com/LastMile-Innovations/snid`
2. Click the "Settings" tab
3. Click "Secrets and variables" → "Actions" in the left sidebar

### 2. Add Each Secret

For each secret listed above:

1. Click "New repository secret"
2. Enter the secret name (exactly as shown in the table)
3. Paste the token value
4. Click "Add secret"

### 3. Verify Configuration

After adding all secrets:

1. Scroll to the "Actions secrets" section
2. Verify all three secrets are listed
3. Ensure no typos in secret names

## Secret Access and Permissions

### Who Can Access Secrets?

- Repository administrators can view and modify secrets
- GitHub Actions workflows can use secrets (but cannot view them)
- Collaborators with write access cannot view secret values

### Workflow Permissions

The release workflow requires the following permissions in `.github/workflows/release.yml`:

```yaml
permissions:
  contents: write  # For creating GitHub releases
```

This permission is already configured in the workflow file.

## Security Best Practices

- **Never commit secrets to git** - Always use GitHub Secrets
- **Use scoped tokens** - Only grant necessary permissions on crates.io/PyPI
- **Rotate tokens regularly** - Revoke and regenerate tokens periodically
- **Monitor token usage** - Check usage logs on crates.io and PyPI
- **Limit secret access** - Only repository admins should manage secrets
- **Use environment-specific secrets** - Consider separate tokens for staging/production

## Testing Secret Configuration

### Test Locally (Dry Run)

Before triggering a real release, test locally with your tokens:

```bash
# Test Rust publishing
cd rust
cargo login
# Enter your CRATES_IO_TOKEN
cargo publish --dry-run

# Test Python publishing
cd python
maturin build --release
maturin publish --username __token__ --password <PYPI_API_TOKEN> --dry-run
```

### Test Workflow (Manual Trigger)

You can test the release workflow without creating a real release:

1. Go to "Actions" tab in GitHub
2. Select "Release" workflow
3. Click "Run workflow"
4. This will run the workflow but won't publish if secrets are misconfigured

## Troubleshooting

### Secret Not Found in Workflow

**Error:** `Input required and not supplied: CRATES_IO_TOKEN`

**Solution:**
- Verify secret name matches exactly (case-sensitive)
- Ensure secret is added to the correct repository
- Check that the workflow has access to secrets

### Invalid Token

**Error:** `403 Forbidden` or `authentication failed`

**Solution:**
- Verify token hasn't expired or been revoked
- Check token has correct permissions on crates.io/PyPI
- Regenerate token if necessary

### Workflow Permission Denied

**Error:** `Resource not accessible by integration`

**Solution:**
- Ensure workflow has `contents: write` permission
- Check repository settings allow Actions to create releases
- Verify GitHub Actions is enabled for the repository

## Related Documentation

- [TOKEN_SETUP.md](TOKEN_SETUP.md) - Detailed token setup instructions
- [PUBLISHING.md](../PUBLISHING.md) - Complete release process guide
- [GitHub Actions Secrets Docs](https://docs.github.com/en/actions/security-guides/encrypted-secrets)

# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| 0.2.x   | Yes       |
| 0.1.x   | No        |

## Reporting a Vulnerability

If you discover a security vulnerability in SNID, please report it responsibly.

### How to Report

Send an email to: oss@neighbor.com

Please include:
- A description of the vulnerability
- Steps to reproduce the issue
- Potential impact if exploited
- Any suggested mitigation (if known)

### What to Expect

- We will acknowledge receipt of your report within 48 hours
- We will provide a detailed response within 7 days
- We will work with you to understand and validate the issue
- We will coordinate a release schedule for fixes
- We will credit you in the release notes (unless you request otherwise)

### Disclosure Policy

We follow responsible disclosure practices:
- We will not disclose vulnerabilities before a fix is released
- We will provide advance notice to users for critical vulnerabilities
- We will publish security advisories for all fixed vulnerabilities

## Security Best Practices

### ID Generation

- SNID IDs are not secrets and can be safely exposed in APIs
- The entropy in IDs provides collision resistance, not cryptographic security
- For cryptographic use cases, use dedicated cryptographic random number generators

### Storage

- Store SNID IDs as raw bytes when possible (16 bytes for SNID/SGID, 32 bytes for extended families)
- Hex fallback is acceptable when raw bytes are not supported
- Never rely on ID structure for access control - use proper authorization mechanisms

### Verification

- LID and KID families include HMAC-based verification tails
- Keep verification keys secure and rotate them regularly
- Use strong, randomly generated keys for HMAC operations

### AKID Credentials

- AKID combines a public SNID head with an opaque secret
- Treat the secret portion as sensitive credentials
- Use AKID for API keys and similar credential scenarios
- Rotate AKID credentials regularly

## Dependency Security

We regularly update dependencies to address security vulnerabilities:

- Go dependencies are updated via `go get -u`
- Rust dependencies are updated via `cargo update`
- Python dependencies are updated via `pip install --upgrade`

## Security Audits

Formal security audits are planned for future releases. Results will be published in this repository.

## Contact

For general security questions: oss@neighbor.com

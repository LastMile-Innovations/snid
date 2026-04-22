# SNID Security Analysis (2026)

## Executive Summary

SNID shares the same security characteristics as UUIDv7 for its core 128-bit time-ordered format: it leaks timestamp and machine information but provides excellent entropy and collision resistance. For security-sensitive or public-facing use cases, SNID should be used with the same precautions as UUIDv7.

**Security Posture:**
- **Internal/Backend Use**: ✅ Excellent (same as UUIDv7)
- **Public-Facing APIs**: ⚠️ Requires caution (timestamp leakage)
- **Session Tokens**: ❌ Not recommended (use UUIDv4 or dedicated token systems)
- **User-Facing IDs**: ⚠️ Use dual-ID strategy (SNID internally + random externally)

## SNID vs Security Vulnerability Categories

| Vulnerability | SNID Status | Details | Mitigation |
|---------------|-------------|---------|------------|
| **Timestamp Leakage** | ⚠️ Present | 48-bit Unix timestamp in MSB (same as UUIDv7) | Use internally only; don't expose in public APIs |
| **Machine Fingerprinting** | ⚠️ Present | 24-bit machine/process fingerprint field | Use internally only; field is optional in some modes |
| **Enumeration/Predictability** | ⚠️ Moderate | 14-bit monotonic sequence + timestamp enables narrow-window guessing | Rate limiting, authorization checks, dual-ID strategy |
| **Weak Randomness** | ✅ Secure | Uses CSPRNG in all implementations (Go crypto/rand, Rust rand, Python secrets) | Monitor for CSPRNG failures (like CVE-2025-66630) |
| **Insufficient Entropy** | ✅ Excellent | 122+ bits of entropy (far above minimum) | No action needed |
| **Information Disclosure** | ⚠️ Present | Timestamp + machine field exposed in wire format | Use dual-ID strategy for public APIs |
| **IDOR Risk** | ✅ Low | Not sequential; time-based but high entropy | Always enforce authorization checks |

## Detailed Security Analysis

### 1. Timestamp Leakage

**SNID Layout:**
- Bits 0-47: Unix timestamp in milliseconds (48 bits)
- This is identical to UUIDv7's timestamp field

**Risk:**
- Attackers can determine exact creation time (millisecond precision)
- Enables user activity profiling, account age inference
- Can narrow ID guessing windows in high-concurrency scenarios

**Mitigation:**
- Use SNID internally for database primary keys
- For public-facing APIs, use a dual-ID strategy:
  - Internal: SNID (time-ordered for DB performance)
  - External: UUIDv4 or sufficiently long NanoID (random for privacy)

### 2. Machine/Node Fingerprinting

**SNID Layout:**
- Bits 66-89: Machine/process fingerprint or projected shard field (24 bits)

**Risk:**
- Can reveal infrastructure scale and topology
- May expose datacenter or worker information
- Could enable targeted infrastructure attacks

**Mitigation:**
- The machine field is implementation-defined
- Can be randomized or omitted in security-sensitive deployments
- Use UUIDv7-compatible mode which uses random bits instead

### 3. Enumeration and Predictability

**SNID Layout:**
- Bits 52-65: Monotonic sequence (14 bits)
- Combined with timestamp, enables narrow-window guessing

**Risk:**
- In high-concurrency scenarios, attacker could guess valid IDs within a short time window
- 14-bit sequence provides 16,384 possible values per millisecond
- With rate limiting, this risk is manageable but non-zero

**Mitigation:**
- Implement strict rate limiting on ID-accepting endpoints
- Use anomaly detection for suspicious patterns
- Always enforce authorization checks (never trust ID obscurity)
- Consider dual-ID strategy for public-facing resources

### 4. Randomness Quality

**SNID Implementations:**
- **Go**: Uses `crypto/rand` (CSPRNG)
- **Rust**: Uses `rand` crate with CSPRNG
- **Python**: Uses `secrets` module (CSPRNG)

**Risk:**
- CSPRNG failures could lead to predictable values (as in CVE-2025-66630)
- Fallback to non-crypto RNG would be catastrophic

**Mitigation:**
- Monitor for CSPRNG failures in production
- Never use `math/rand` (Go) or `random` (Python) for ID generation
- Add health checks for RNG quality in CI/CD

### 5. Entropy and Collision Resistance

**SNID Entropy:**
- Total: 122+ bits of entropy
- Timestamp: 48 bits (not random, but adds uniqueness)
- Sequence: 14 bits (monotonic, not random)
- Machine: 24 bits (can be random or deterministic)
- Entropy tail: 38 bits (random)

**Assessment:**
- Excellent collision resistance
- Birthday paradox requires ~2.7 × 10¹⁸ IDs for 50% collision chance
- Far above minimum for any realistic use case

**Mitigation:**
- No action needed; entropy is sufficient

### 6. Information Disclosure in Public APIs

**SNID Wire Format:**
- `<ATOM>:<payload>` where payload is Base58-encoded 16-byte SNID + checksum
- Base58 encoding is reversible (not one-way)

**Risk:**
- Exposing SNID in URLs or API responses reveals timestamp and machine field
- Can be used for reconnaissance and privacy violations
- May violate GDPR/CCPA if not documented

**Mitigation:**
- Use dual-ID strategy for public APIs:
  ```go
  // Internal database
  type User struct {
      ID       SNID    // Time-ordered for DB performance
      PublicID string  // Random UUIDv4 or NanoID for APIs
  }
  ```
- Never expose SNID in public URLs or responses
- Document any time-based IDs in privacy policies

### 7. IDOR (Insecure Direct Object References)

**SNID Characteristics:**
- Not sequential (unlike auto-increment)
- Time-ordered but high entropy
- Not trivially enumerable

**Risk:**
- While better than sequential IDs, time-based IDs can still be guessed within windows
- Authorization bypass is possible if checks are weak

**Mitigation:**
- **Always enforce proper authorization checks**
- Never trust ID obscurity for security
- Implement rate limiting and anomaly detection
- Use dual-ID strategy for public resources

## Comparison with Other ID Types

| ID Type | Timestamp Leakage | Machine Leakage | Entropy | Predictability | Public API Safety |
|---------|------------------|-----------------|---------|----------------|-------------------|
| **SNID** | ⚠️ Yes (48-bit ms) | ⚠️ Yes (24-bit) | ✅ 122+ bits | ⚠️ Moderate (time-based) | ⚠️ Use dual-ID |
| **UUIDv7** | ⚠️ Yes (48-bit ms) | ✅ No | ✅ 122+ bits | ⚠️ Moderate (time-based) | ⚠️ Use dual-ID |
| **UUIDv4** | ✅ No | ✅ No | ✅ 122 bits | ✅ Very low | ✅ Safe |
| **ULID** | ⚠️ Yes (48-bit ms) | ✅ No | ✅ 122+ bits | ⚠️ Moderate (time-based) | ⚠️ Use dual-ID |
| **NanoID** (21 chars) | ✅ No | ✅ No | ✅ ~126 bits | ✅ Very low | ✅ Safe |
| **CUID2** | ⚠️ Partial (obfuscated) | ⚠️ Partial (fingerprint) | ✅ High | ✅ Low | ✅ Generally safe |
| **Sequential** | ✅ No | ✅ No | ❌ Very low | ❌ Very high | ❌ Unsafe |
| **Snowflake** | ⚠️ Yes | ⚠️ Yes (worker ID) | ⚠️ ~64 bits | ⚠️ Moderate | ⚠️ Use dual-ID |

## Recommended Security Practices

### 1. Dual-ID Strategy (Recommended for Public APIs)

```go
// Internal database
type Resource struct {
    ID       SNID    // Time-ordered for DB performance
    Slug     string  // Human-readable URL slug
    PublicID string  // Random UUIDv4 for API responses
}

// API response
type ResourceResponse struct {
    ID       string  // PublicID (UUIDv4)
    Slug     string  // Human-readable
    // ... other fields
}
```

**Benefits:**
- Internal: DB performance from time-ordered SNID
- External: Privacy from random public ID
- No timestamp leakage in public APIs

### 2. Authorization Checks (Critical)

```go
// NEVER do this (insecure)
func getResource(id SNID) (*Resource, error) {
    return db.Get(id)
}

// ALWAYS do this (secure)
func getResource(id SNID, user User) (*Resource, error) {
    resource, err := db.Get(id)
    if err != nil {
        return nil, err
    }
    if !user.CanAccess(resource) {
        return nil, ErrUnauthorized
    }
    return resource, nil
}
```

### 3. Rate Limiting and Detection

```go
// Implement strict rate limiting
func apiHandler(w http.ResponseWriter, r *http.Request) {
    if rateLimiter.Allow(r.RemoteAddr) == false {
        http.Error(w, "Too many requests", 429)
        return
    }
    // ... handle request
}

// Detect suspicious patterns
func detectSuspiciousPattern(requests []Request) bool {
    // Check for sequential ID guessing
    // Check for rapid time-window scanning
    // Check for anomalous user agents
}
```

### 4. CSPRNG Monitoring

```go
// Add health check for RNG quality
func checkRNGHealth() error {
    // Generate test IDs and check for randomness
    // Monitor for CSPRNG failures
    // Alert if fallback to non-crypto RNG detected
}
```

### 5. Privacy Compliance

- Document any time-based IDs in privacy policies
- Justify timestamp leakage for internal systems
- Provide data retention policies for time-ordered data
- Consider GDPR/CCPA implications when exposing IDs

## Specialized SNID Families

### AKID (Access Key ID)

**Security Posture:**
- Dual-part: public SNID + opaque secret
- Public SNID can be exposed (timestamp leakage acceptable)
- Secret must be treated as credential (never exposed)

**Best Practices:**
- Store secret in secure storage (HashiCorp Vault, AWS Secrets Manager)
- Use constant-time comparison for secret verification
- Rotate secrets regularly
- Never log or expose secret in error messages

### LID (Log ID with Verification)

**Security Posture:**
- 256-bit: 16-byte SNID head + 16-byte HMAC verification tail
- Verification tail enables tamper detection
- Timestamp leakage acceptable for immutable logs

**Best Practices:**
- Use strong HMAC key (256-bit or better)
- Store verification key securely
- Rotate keys periodically
- Use for audit trails, not for access control

### KID (Capability ID)

**Security Posture:**
- 256-bit: 16-byte SNID head + 16-byte MAC tail
- Binds actor + resource + capability
- Self-verifying capability grant

**Best Practices:**
- Use strong MAC key
- Implement proper capability revocation
- Monitor for capability abuse
- Use for authorization, not authentication

## Conclusion

SNID provides excellent security characteristics for internal/backend use cases, with the same trade-offs as UUIDv7:

**Strengths:**
- ✅ Excellent entropy (122+ bits)
- ✅ Uses CSPRNG in all implementations
- ✅ Not sequential (better than auto-increment)
- ✅ Collision resistance at any realistic scale

**Weaknesses:**
- ⚠️ Timestamp leakage (same as UUIDv7)
- ⚠️ Machine fingerprinting (optional field)
- ⚠️ Moderate predictability within time windows

**Recommendation:**
- Use SNID internally for database primary keys (excellent performance + security)
- Use dual-ID strategy for public APIs (SNID internally + random externally)
- Never rely on ID obscurity for security (always enforce authorization)
- Monitor for CSPRNG failures and suspicious patterns

For maximum security in public-facing scenarios, use UUIDv4 or sufficiently long NanoID (21+ chars) for external IDs, while keeping SNID for internal database performance.

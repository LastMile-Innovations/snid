//! Helper functions for hashing, encoding, and string manipulation.

// FNV-1a 32-bit hash
pub fn fnv1a(s: &str) -> u32 {
    let mut h: u32 = 2166136261;
    for b in s.bytes() {
        h ^= b as u32;
        h = h.wrapping_mul(16777619);
    }
    h
}

#[allow(dead_code)]
// FNV-1a 64-bit hash
pub fn fnv1a64(s: &str) -> u64 {
    let mut h: u64 = 14695981039346656037;
    for b in s.bytes() {
        h ^= b as u64;
        h = h.wrapping_mul(1099511628211);
    }
    h
}

#[allow(dead_code)]
// FNV-1a 32-bit hash with ASCII uppercasing
pub fn fnv1a32_upper(s: &str) -> u32 {
    let mut h: u32 = 2166136261;
    for b in s.bytes() {
        let c = if b.is_ascii_lowercase() { b - 32 } else { b };
        h ^= c as u32;
        h = h.wrapping_mul(16777619);
    }
    h
}

#[allow(dead_code)]
// FNV-1a 64-bit hash with ASCII uppercasing
pub fn fnv1a64_upper(s: &str) -> u64 {
    let mut h: u64 = 14695981039346656037;
    for b in s.bytes() {
        let c = if b.is_ascii_lowercase() { b - 32 } else { b };
        h ^= c as u64;
        h = h.wrapping_mul(1099511628211);
    }
    h
}

#[allow(dead_code)]
// FNV-1a 64-bit hash with ASCII lowercasing
pub fn fnv1a64_lower(s: &str) -> u64 {
    let mut h: u64 = 14695981039346656037;
    for b in s.bytes() {
        let c = if b.is_ascii_uppercase() { b + 32 } else { b };
        h ^= c as u64;
        h = h.wrapping_mul(1099511628211);
    }
    h
}

// Sanitize alias: keep alphanumeric, -, _, convert spaces to -
pub fn sanitize_alias(s: &str) -> String {
    s.chars()
        .map(|c| {
            if c.is_ascii_lowercase()
                || c.is_ascii_uppercase()
                || c.is_ascii_digit()
                || c == '-'
                || c == '_'
            {
                c
            } else if c == ' ' {
                '-'
            } else {
                '\0'
            }
        })
        .filter(|&c| c != '\0')
        .collect()
}

// Zero-alloc hex encoding for TraceID
pub const HEX_CHARS: &[u8; 16] = b"0123456789abcdef";
#[allow(dead_code)]
pub const HEX_CHARS_UPPER: &[u8; 16] = b"0123456789ABCDEF";

pub fn hex_encode_fast(bytes: &[u8]) -> String {
    let mut out = Vec::with_capacity(bytes.len() * 2);
    for &b in bytes {
        out.push(HEX_CHARS[(b >> 4) as usize]);
        out.push(HEX_CHARS[(b & 0x0F) as usize]);
    }
    String::from_utf8(out).unwrap()
}

pub fn hex_encode_to<'a>(bytes: &[u8], out: &'a mut [u8]) -> &'a str {
    debug_assert!(out.len() >= bytes.len() * 2);
    let mut cursor = 0usize;
    for &byte in bytes {
        out[cursor] = HEX_CHARS[(byte >> 4) as usize];
        out[cursor + 1] = HEX_CHARS[(byte & 0x0F) as usize];
        cursor += 2;
    }
    // SAFETY: every written byte comes from the lowercase ASCII hex table.
    unsafe { std::str::from_utf8_unchecked(&out[..cursor]) }
}

pub fn hex_decode_to(input: &str, out: &mut [u8]) -> Result<(), crate::error::Error> {
    let bytes = input.as_bytes();
    if bytes.len() != out.len() * 2 {
        return Err(crate::error::Error::InvalidLength);
    }
    for (idx, slot) in out.iter_mut().enumerate() {
        let hi = hex_value(bytes[idx * 2])?;
        let lo = hex_value(bytes[idx * 2 + 1])?;
        *slot = (hi << 4) | lo;
    }
    Ok(())
}

#[allow(dead_code)]
pub fn hex_decode_vec(input: &str) -> Result<Vec<u8>, crate::error::Error> {
    if !input.len().is_multiple_of(2) {
        return Err(crate::error::Error::InvalidLength);
    }
    let mut out = vec![0u8; input.len() / 2];
    hex_decode_to(input, &mut out)?;
    Ok(out)
}

#[inline(always)]
fn hex_value(byte: u8) -> Result<u8, crate::error::Error> {
    match byte {
        b'0'..=b'9' => Ok(byte - b'0'),
        b'a'..=b'f' => Ok(byte - b'a' + 10),
        b'A'..=b'F' => Ok(byte - b'A' + 10),
        _ => Err(crate::error::Error::InvalidFormat),
    }
}

pub fn splitmix64(seed: &mut u64) -> u64 {
    *seed = seed.wrapping_add(0x9e3779b97f4a7c15);
    let mut z = *seed;
    z = (z ^ (z >> 30)).wrapping_mul(0xbf58476d1ce4e5b9);
    z = (z ^ (z >> 27)).wrapping_mul(0x94d049bb133111eb);
    z ^ (z >> 31)
}

pub fn expand_hash_material(hash: &[u8]) -> [u8; 16] {
    let mut out = [0u8; 16];
    if hash.is_empty() {
        return out;
    }
    for i in 0..16 {
        out[i] = hash[i % hash.len()];
    }
    out
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_fnv1a() {
        let hash = fnv1a("test");
        assert_ne!(hash, 0);
    }

    #[test]
    fn test_fnv1a_empty() {
        let hash = fnv1a("");
        assert_eq!(hash, 2166136261);
    }

    #[test]
    fn test_fnv1a_consistent() {
        let hash1 = fnv1a("test");
        let hash2 = fnv1a("test");
        assert_eq!(hash1, hash2);
    }

    #[test]
    fn test_fnv1a64() {
        let hash = fnv1a64("test");
        assert_ne!(hash, 0);
    }

    #[test]
    fn test_fnv1a64_empty() {
        let hash = fnv1a64("");
        assert_eq!(hash, 14695981039346656037);
    }

    #[test]
    fn test_fnv1a64_consistent() {
        let hash1 = fnv1a64("test");
        let hash2 = fnv1a64("test");
        assert_eq!(hash1, hash2);
    }

    #[test]
    fn test_fnv1a32_upper() {
        let hash = fnv1a32_upper("test");
        assert_ne!(hash, 0);
    }

    #[test]
    fn test_fnv1a32_upper_mixed_case() {
        let hash1 = fnv1a32_upper("TEST");
        let hash2 = fnv1a32_upper("test");
        assert_eq!(hash1, hash2);
    }

    #[test]
    fn test_fnv1a64_upper() {
        let hash = fnv1a64_upper("test");
        assert_ne!(hash, 0);
    }

    #[test]
    fn test_fnv1a64_upper_mixed_case() {
        let hash1 = fnv1a64_upper("TEST");
        let hash2 = fnv1a64_upper("test");
        assert_eq!(hash1, hash2);
    }

    #[test]
    fn test_fnv1a64_lower() {
        let hash = fnv1a64_lower("TEST");
        assert_ne!(hash, 0);
    }

    #[test]
    fn test_fnv1a64_lower_mixed_case() {
        let hash1 = fnv1a64_lower("TeSt");
        let hash2 = fnv1a64_lower("TEST");
        assert_eq!(hash1, hash2);
    }

    #[test]
    fn test_sanitize_alias() {
        let sanitized = sanitize_alias("test-alias_123");
        assert_eq!(sanitized, "test-alias_123");
    }

    #[test]
    fn test_sanitize_alias_spaces() {
        let sanitized = sanitize_alias("test alias");
        assert_eq!(sanitized, "test-alias");
    }

    #[test]
    fn test_sanitize_alias_special_chars() {
        let sanitized = sanitize_alias("test@alias#123");
        assert_eq!(sanitized, "testalias123");
    }

    #[test]
    fn test_sanitize_alias_empty() {
        let sanitized = sanitize_alias("");
        assert_eq!(sanitized, "");
    }

    #[test]
    fn test_hex_encode_fast() {
        let bytes = [0x12, 0x34, 0x56, 0x78];
        let encoded = hex_encode_fast(&bytes);
        assert_eq!(encoded, "12345678");
    }

    #[test]
    fn test_hex_encode_fast_empty() {
        let bytes: [u8; 0] = [];
        let encoded = hex_encode_fast(&bytes);
        assert_eq!(encoded, "");
    }

    #[test]
    fn test_hex_encode_fast_all_zeros() {
        let bytes = [0u8; 4];
        let encoded = hex_encode_fast(&bytes);
        assert_eq!(encoded, "00000000");
    }

    #[test]
    fn test_hex_encode_fast_all_ones() {
        let bytes = [0xFFu8; 4];
        let encoded = hex_encode_fast(&bytes);
        assert_eq!(encoded, "ffffffff");
    }

    #[test]
    fn test_splitmix64() {
        let mut seed = 12345u64;
        let hash1 = splitmix64(&mut seed);
        let hash2 = splitmix64(&mut seed);
        assert_ne!(hash1, hash2);
    }

    #[test]
    fn test_splitmix64_consistent() {
        let mut seed1 = 12345u64;
        let mut seed2 = 12345u64;
        let hash1 = splitmix64(&mut seed1);
        let hash2 = splitmix64(&mut seed2);
        assert_eq!(hash1, hash2);
    }

    #[test]
    fn test_expand_hash_material() {
        let hash = b"test";
        let expanded = expand_hash_material(hash);
        assert_ne!(expanded, [0u8; 16]);
    }

    #[test]
    fn test_expand_hash_material_empty() {
        let hash = b"";
        let expanded = expand_hash_material(hash);
        assert_eq!(expanded, [0u8; 16]);
    }

    #[test]
    fn test_expand_hash_material_short() {
        let hash = b"ab";
        let expanded = expand_hash_material(hash);
        // Should repeat the pattern
        assert_eq!(expanded[0], b'a');
        assert_eq!(expanded[1], b'b');
        assert_eq!(expanded[2], b'a');
        assert_eq!(expanded[3], b'b');
    }

    #[test]
    fn test_expand_hash_material_long() {
        let hash = b"0123456789abcdef";
        let expanded = expand_hash_material(hash);
        assert_eq!(expanded[0], b'0');
        assert_eq!(expanded[15], b'f');
    }

    #[test]
    fn test_hex_chars() {
        assert_eq!(HEX_CHARS.len(), 16);
        assert_eq!(HEX_CHARS[0], b'0');
        assert_eq!(HEX_CHARS[15], b'f');
    }

    #[test]
    fn test_hex_chars_upper() {
        assert_eq!(HEX_CHARS_UPPER.len(), 16);
        assert_eq!(HEX_CHARS_UPPER[0], b'0');
        assert_eq!(HEX_CHARS_UPPER[15], b'F');
    }
}

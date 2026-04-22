// SNID Property-Based Tests using proptest
// Tests invariants across generated inputs to ensure protocol correctness.

use proptest::prelude::*;
use snid::Snid;

// =============================================================================
// Property Tests
// =============================================================================

proptest! {
    #[test]
    fn prop_bytes_roundtrip(timestamp_ms in 0u64..(1u64 << 48)) {
        // Bytes representation should roundtrip through from_bytes
        let hash = [0u8; 32];
        let id1 = Snid::from_hash_with_timestamp(timestamp_ms, &hash);
        let bytes_repr = id1.to_bytes();
        let id2 = Snid::from_bytes(bytes_repr);
        prop_assert_eq!(id1, id2);
    }

    #[test]
    fn prop_id_length(timestamp_ms in 0u64..(1u64 << 48)) {
        // ID should always be 16 bytes
        let hash = [0u8; 32];
        let id1 = Snid::from_hash_with_timestamp(timestamp_ms, &hash);
        let bytes_repr = id1.to_bytes();
        prop_assert_eq!(bytes_repr.len(), 16);
    }

    #[test]
    fn prop_version_bits(timestamp_ms in 0u64..(1u64 << 48)) {
        // Version bits should be set correctly for UUIDv7 compatibility
        let hash = [0u8; 32];
        let id1 = Snid::from_hash_with_timestamp(timestamp_ms, &hash);
        let bytes_repr = id1.to_bytes();
        // Version is in bits 48-51 (byte 6, bits 4-7)
        let version = (bytes_repr[6] >> 4) & 0x0F;
        prop_assert_eq!(version, 0x7);
    }

    #[test]
    fn prop_variant_bits(timestamp_ms in 0u64..(1u64 << 48)) {
        // Variant bits should be set correctly for RFC 4122
        let hash = [0u8; 32];
        let id1 = Snid::from_hash_with_timestamp(timestamp_ms, &hash);
        let bytes_repr = id1.to_bytes();
        // Variant is in bits 64-65 (byte 8, bits 6-7)
        let variant = (bytes_repr[8] >> 6) & 0b11;
        prop_assert_eq!(variant, 0b10);
    }
}

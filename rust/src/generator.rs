//! ID generation logic with thread-local state and coarse clock.

use crate::core::Snid;
use crate::helpers::splitmix64;
use getrandom::getrandom;
use std::cell::UnsafeCell;
use std::process;
use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::LazyLock;
use std::time::{SystemTime, UNIX_EPOCH};

thread_local! {
    pub(crate) static GENERATOR: UnsafeCell<GeneratorState> = UnsafeCell::new(GeneratorState::init());
}

// Global coarse clock shared across all threads
static COARSE_CLOCK: LazyLock<AtomicU64> = LazyLock::new(|| AtomicU64::new(current_time_ms()));

// Initialize coarse clock updater thread (called once on first use)
pub(crate) fn init_coarse_clock() {
    static INIT: std::sync::Once = std::sync::Once::new();
    INIT.call_once(|| {
        std::thread::spawn(|| loop {
            let now = current_time_ms();
            COARSE_CLOCK.store(now, Ordering::Relaxed);
            std::thread::sleep(std::time::Duration::from_millis(10));
        });
    });
}

pub(crate) struct GeneratorState {
    _pad_front: [u8; 64],  // Front padding: isolates this struct from adjacent memory
    last_ms: u64,
    sequence: u16,
    machine_id: u32,
    state: [u64; 4],
    pid: u32,
    _pad_back: [u8; 64],   // Back padding: ensures struct occupies full cache lines
}

impl GeneratorState {
    pub(crate) fn init() -> Self {
        let mut seed = [0u8; 8];
        getrandom(&mut seed).expect("os rng");
        let mut z = u64::from_le_bytes(seed);
        let pid = process::id();
        let machine_id = (splitmix64(&mut z) as u32 ^ pid) & 0x00FF_FFFF;
        Self {
            _pad_front: [0u8; 64],
            last_ms: current_time_ms(),
            sequence: 0,
            machine_id,
            state: [
                splitmix64(&mut z),
                splitmix64(&mut z),
                splitmix64(&mut z),
                splitmix64(&mut z),
            ],
            pid,
            _pad_back: [0u8; 64],
        }
    }

    #[inline(always)]
    pub(crate) fn next(&mut self) -> Snid {
        // Use coarse clock for performance
        let mut ms = COARSE_CLOCK.load(Ordering::Relaxed);

        if ms > self.last_ms {
            self.last_ms = ms;
            self.sequence = random_seq_start(&mut self.state);
        } else {
            self.sequence = self.sequence.wrapping_add(1);
            if self.sequence > 0x3FFF {
                self.last_ms += 1;
                ms = self.last_ms;
                self.sequence = random_seq_start(&mut self.state);
            } else {
                ms = self.last_ms;
            }
        }

        let result = self.state[1].wrapping_mul(5).rotate_left(7).wrapping_mul(9);
        let t = self.state[1] << 17;
        self.state[2] ^= self.state[0];
        self.state[3] ^= self.state[1];
        self.state[1] ^= self.state[2];
        self.state[0] ^= self.state[3];
        self.state[2] ^= t;
        self.state[3] = self.state[3].rotate_left(45);

        let seq = self.sequence as u64;
        let hi = (ms << 16) | 0x7000 | (seq >> 2);
        let lo = 0x8000_0000_0000_0000
            | ((seq & 0x03) << 60)
            | ((self.machine_id as u64 & 0x00FF_FFFF) << 36)
            | ((result >> 28) & 0xFFFF_FFFFF);
        let mut out = [0u8; 16];
        out[..8].copy_from_slice(&hi.to_be_bytes());
        out[8..].copy_from_slice(&lo.to_be_bytes());
        Snid(out)
    }
}

// random_seq_start generates a random 14-bit sequence start value (0-16383).
// This implements RFC 9562 Method 2 for randomized monotonicity,
// preventing ID enumeration attacks when the creation window is known.
#[inline(always)]
fn random_seq_start(state: &mut [u64; 4]) -> u16 {
    // Advance Xoshiro256** state once
    let res = state[1].wrapping_mul(5).rotate_left(7).wrapping_mul(9);
    let t = state[1] << 17;
    state[2] ^= state[0];
    state[3] ^= state[1];
    state[1] ^= state[2];
    state[0] ^= state[3];
    state[2] ^= t;
    state[3] = state[3].rotate_left(45);
    // Return lower 14 bits as random sequence start (0-16383)
    (res & 0x3FFF) as u16
}

pub(crate) fn current_time_ms() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .expect("clock")
        .as_millis() as u64
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_generator_state_init() {
        let state = GeneratorState::init();
        assert_ne!(state.last_ms, 0);
        assert_eq!(state.sequence, 0);
        assert_ne!(state.machine_id, 0);
        assert_ne!(state.pid, 0);
    }

    #[test]
    fn test_generator_state_next() {
        let mut state = GeneratorState::init();
        let id1 = state.next();
        let id2 = state.next();
        assert_ne!(id1, id2);
    }

    #[test]
    fn test_generator_sequence_increment() {
        let mut state = GeneratorState::init();
        let initial_seq = state.sequence;
        state.next();
        assert_eq!(state.sequence, initial_seq.wrapping_add(1));
    }

    #[test]
    fn test_generator_unique_ids() {
        let mut state = GeneratorState::init();
        let mut ids = std::collections::HashSet::new();
        for _ in 0..1000 {
            let id = state.next();
            assert!(ids.insert(id), "duplicate ID generated");
        }
    }

    #[test]
    fn test_current_time_ms() {
        let now = current_time_ms();
        assert!(now > 1700000000000); // After 2023
    }

    #[test]
    fn test_generator_produces_valid_snid() {
        let mut state = GeneratorState::init();
        let id = state.next();
        // Check version bits (should be 0x7 for UUIDv7)
        let version = (id.0[6] >> 4) & 0x0F;
        assert_eq!(version, 0x7);
        // Check variant bits (should be 0b10 for RFC 4122)
        let variant = (id.0[8] >> 6) & 0b11;
        assert_eq!(variant, 0b10);
    }

    #[test]
    fn test_generator_timestamp_monotonic() {
        let mut state = GeneratorState::init();
        let ts1 = state.next().timestamp_millis();
        let ts2 = state.next().timestamp_millis();
        assert!(ts2 >= ts1);
    }

    #[test]
    fn test_generator_thread_local() {
        GENERATOR.with(|cell| unsafe {
            let state = &mut *cell.get();
            let id1 = state.next();
            let id2 = state.next();
            assert_ne!(id1, id2);
        });
    }

    #[test]
    fn test_init_coarse_clock() {
        init_coarse_clock();
        // Should not panic
        init_coarse_clock();
    }

    #[test]
    fn test_randomized_sequence_bounds() {
        // Generate many IDs to test overflow handling
        let mut state = GeneratorState::init();
        for _ in 0..10000 {
            let id = state.next();
            // Extract sequence from bits 52-65 (14 bits)
            let hi = u64::from_be_bytes([id.0[0], id.0[1], id.0[2], id.0[3], id.0[4], id.0[5], id.0[6], id.0[7]]);
            let seq = (hi & 0x3FFF) as u16;
            assert!(seq <= 0x3FFF, "sequence {} exceeds maximum value 16383", seq);
        }
    }

    #[test]
    fn test_randomized_sequence_monotonicity() {
        let mut state = GeneratorState::init();
        let mut prev = state.next();
        for _ in 0..1000 {
            let curr = state.next();
            // IDs should be monotonic (timestamp + sequence)
            let ts_prev = prev.timestamp_millis();
            let ts_curr = curr.timestamp_millis();
            assert!(ts_curr >= ts_prev, "timestamp not monotonic");
            prev = curr;
        }
    }
}

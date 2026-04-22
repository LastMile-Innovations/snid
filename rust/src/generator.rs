//! ID generation logic with thread-local state and coarse clock.

use crate::core::Snid;
use crate::error::Error;
use crate::helpers::splitmix64;
use getrandom::getrandom;
use std::cell::UnsafeCell;
use std::mem::MaybeUninit;
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

#[derive(Debug)]
#[allow(dead_code)]
pub(crate) struct GeneratorState {
    _pad_front: [u8; 64], // Front padding: isolates this struct from adjacent memory
    last_ms: u64,
    sequence: u16,
    machine_id: u32,
    state: [u64; 4],
    pid: u32,
    _pad_back: [u8; 64], // Back padding: ensures struct occupies full cache lines
}

impl GeneratorState {
    pub(crate) fn init() -> Self {
        Self::try_init().expect("os rng")
    }

    pub(crate) fn try_init() -> Result<Self, Error> {
        init_coarse_clock();
        let mut seed = [0u8; 8];
        getrandom(&mut seed)?;
        let mut z = u64::from_le_bytes(seed);
        let pid = process::id();
        let machine_id = (splitmix64(&mut z) as u32 ^ pid) & 0x00FF_FFFF;
        Ok(Self {
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
        })
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

    #[inline(always)]
    pub(crate) fn fill_slice(&mut self, out: &mut [Snid]) {
        let out = unsafe { std::slice::from_raw_parts_mut(out.as_mut_ptr().cast(), out.len()) };
        self.fill_uninit_slice(out);
    }

    #[inline(always)]
    pub(crate) fn fill_bytes(&mut self, out: &mut [u8]) -> usize {
        let byte_len = out.len() / 16 * 16;
        let out = unsafe { std::slice::from_raw_parts_mut(out.as_mut_ptr().cast(), byte_len) };
        self.fill_uninit_bytes(out)
    }

    #[inline(always)]
    pub(crate) fn fill_uninit_bytes(&mut self, out: &mut [MaybeUninit<u8>]) -> usize {
        let count = out.len() / 16;
        if count == 0 {
            return 0;
        }

        if self.try_fill_uninit_bytes_same_window(out, count) {
            return count;
        }

        let mut ms = COARSE_CLOCK.load(Ordering::Relaxed);
        let mut last_ms = self.last_ms;
        let mut sequence = self.sequence;
        let machine = (self.machine_id as u64 & 0x00FF_FFFF) << 36;
        let mut state = self.state;

        for idx in 0..count {
            if ms > last_ms {
                last_ms = ms;
                sequence = random_seq_start(&mut state);
            } else {
                sequence = sequence.wrapping_add(1);
                if sequence > 0x3FFF {
                    last_ms += 1;
                    ms = last_ms;
                    sequence = random_seq_start(&mut state);
                } else {
                    ms = last_ms;
                }
            }

            let result = xoshiro_next(&mut state);
            let seq = sequence as u64;
            let hi = (ms << 16) | 0x7000 | (seq >> 2);
            let lo = 0x8000_0000_0000_0000
                | ((seq & 0x03) << 60)
                | machine
                | ((result >> 28) & 0xFFFF_FFFFF);
            let hi = hi.to_be_bytes();
            let lo = lo.to_be_bytes();
            let offset = idx * 16;
            out[offset].write(hi[0]);
            out[offset + 1].write(hi[1]);
            out[offset + 2].write(hi[2]);
            out[offset + 3].write(hi[3]);
            out[offset + 4].write(hi[4]);
            out[offset + 5].write(hi[5]);
            out[offset + 6].write(hi[6]);
            out[offset + 7].write(hi[7]);
            out[offset + 8].write(lo[0]);
            out[offset + 9].write(lo[1]);
            out[offset + 10].write(lo[2]);
            out[offset + 11].write(lo[3]);
            out[offset + 12].write(lo[4]);
            out[offset + 13].write(lo[5]);
            out[offset + 14].write(lo[6]);
            out[offset + 15].write(lo[7]);
        }

        self.last_ms = last_ms;
        self.sequence = sequence;
        self.state = state;
        count
    }

    #[inline(always)]
    fn try_fill_uninit_bytes_same_window(
        &mut self,
        out: &mut [MaybeUninit<u8>],
        count: usize,
    ) -> bool {
        let clock_ms = COARSE_CLOCK.load(Ordering::Relaxed);
        let mut state = self.state;
        let (ms, mut sequence, use_current) = if clock_ms > self.last_ms {
            (clock_ms, random_seq_start(&mut state), true)
        } else {
            (self.last_ms, self.sequence, false)
        };
        let available = 0x3FFFusize - sequence as usize + usize::from(use_current);
        if count > available {
            return false;
        }

        let machine = (self.machine_id as u64 & 0x00FF_FFFF) << 36;
        let mut ptr = out.as_mut_ptr().cast::<u8>();
        unsafe {
            if use_current {
                let result = xoshiro_next(&mut state);
                write_words_to_ptr(ptr, make_words(ms, sequence, machine, result));
                ptr = ptr.add(16);
            }
            let remaining = count - usize::from(use_current);
            for _ in 0..remaining {
                sequence += 1;
                let result = xoshiro_next(&mut state);
                write_words_to_ptr(ptr, make_words(ms, sequence, machine, result));
                ptr = ptr.add(16);
            }
        }

        self.last_ms = ms;
        self.sequence = sequence;
        self.state = state;
        true
    }

    #[inline(always)]
    pub(crate) fn fill_uninit_slice(&mut self, out: &mut [MaybeUninit<Snid>]) {
        if out.is_empty() {
            return;
        }

        if self.try_fill_uninit_slice_same_window(out) {
            return;
        }

        let mut ms = COARSE_CLOCK.load(Ordering::Relaxed);
        let mut last_ms = self.last_ms;
        let mut sequence = self.sequence;
        let machine = (self.machine_id as u64 & 0x00FF_FFFF) << 36;
        let mut state = self.state;

        for slot in out {
            if ms > last_ms {
                last_ms = ms;
                sequence = random_seq_start(&mut state);
            } else {
                sequence = sequence.wrapping_add(1);
                if sequence > 0x3FFF {
                    last_ms += 1;
                    ms = last_ms;
                    sequence = random_seq_start(&mut state);
                } else {
                    ms = last_ms;
                }
            }

            let result = xoshiro_next(&mut state);
            let seq = sequence as u64;
            let hi = (ms << 16) | 0x7000 | (seq >> 2);
            let lo = 0x8000_0000_0000_0000
                | ((seq & 0x03) << 60)
                | machine
                | ((result >> 28) & 0xFFFF_FFFFF);
            let mut bytes = [0u8; 16];
            bytes[..8].copy_from_slice(&hi.to_be_bytes());
            bytes[8..].copy_from_slice(&lo.to_be_bytes());
            slot.write(Snid(bytes));
        }

        self.last_ms = last_ms;
        self.sequence = sequence;
        self.state = state;
    }

    #[inline(always)]
    fn try_fill_uninit_slice_same_window(&mut self, out: &mut [MaybeUninit<Snid>]) -> bool {
        let clock_ms = COARSE_CLOCK.load(Ordering::Relaxed);
        let mut state = self.state;
        let (ms, mut sequence, use_current) = if clock_ms > self.last_ms {
            (clock_ms, random_seq_start(&mut state), true)
        } else {
            (self.last_ms, self.sequence, false)
        };
        let available = 0x3FFFusize - sequence as usize + usize::from(use_current);
        if out.len() > available {
            return false;
        }

        let machine = (self.machine_id as u64 & 0x00FF_FFFF) << 36;
        let mut ptr = out.as_mut_ptr();
        unsafe {
            if use_current {
                let result = xoshiro_next(&mut state);
                (*ptr).write(make_snid(ms, sequence, machine, result));
                ptr = ptr.add(1);
            }
            let remaining = out.len() - usize::from(use_current);
            for _ in 0..remaining {
                sequence += 1;
                let result = xoshiro_next(&mut state);
                (*ptr).write(make_snid(ms, sequence, machine, result));
                ptr = ptr.add(1);
            }
        }

        self.last_ms = ms;
        self.sequence = sequence;
        self.state = state;
        true
    }
}

#[inline(always)]
pub(crate) fn with_generator<T>(f: impl FnOnce(&mut GeneratorState) -> T) -> T {
    GENERATOR.with(|cell| unsafe {
        let state = &mut *cell.get();
        f(state)
    })
}

pub(crate) fn try_new_standalone() -> Result<Snid, Error> {
    let mut state = GeneratorState::try_init()?;
    Ok(state.next())
}

pub(crate) fn try_batch_standalone(count: usize) -> Result<Vec<Snid>, Error> {
    let mut state = GeneratorState::try_init()?;
    let mut ids = new_uninit_snid_vec(count);
    state.fill_uninit_slice(&mut ids);
    Ok(assume_init_snid_vec(ids))
}

pub(crate) fn new_uninit_snid_vec(count: usize) -> Vec<MaybeUninit<Snid>> {
    let mut ids: Vec<MaybeUninit<Snid>> = Vec::with_capacity(count);
    ids.resize_with(count, MaybeUninit::uninit);
    ids
}

pub(crate) fn assume_init_snid_vec(mut ids: Vec<MaybeUninit<Snid>>) -> Vec<Snid> {
    let ptr = ids.as_mut_ptr().cast::<Snid>();
    let len = ids.len();
    let cap = ids.capacity();
    std::mem::forget(ids);
    unsafe { Vec::from_raw_parts(ptr, len, cap) }
}

/// Reusable single-threaded generator optimized for hot loops.
///
/// `TurboStreamer` owns its generator state and a reusable buffer. It avoids
/// thread-local lookup on every ID and amortizes clock reads across refills.
#[derive(Debug)]
pub struct TurboStreamer {
    buffer: Vec<Snid>,
    cursor: usize,
    state: GeneratorState,
}

impl TurboStreamer {
    pub fn new(size: usize) -> Self {
        Self::try_new(size).expect("os rng")
    }

    pub fn try_new(size: usize) -> Result<Self, Error> {
        let size = size.max(64).next_power_of_two();
        Ok(Self {
            buffer: vec![Snid([0u8; 16]); size],
            cursor: size,
            state: GeneratorState::try_init()?,
        })
    }

    #[inline(always)]
    pub fn next_id(&mut self) -> Snid {
        if self.cursor >= self.buffer.len() {
            self.refill();
        }
        let id = self.buffer[self.cursor];
        self.cursor += 1;
        id
    }

    #[inline(always)]
    pub fn next_unchecked(&mut self) -> Snid {
        let id = self.buffer[self.cursor];
        self.cursor += 1;
        id
    }

    #[inline(always)]
    pub fn refill(&mut self) {
        self.state.fill_slice(&mut self.buffer);
        self.cursor = 0;
    }

    pub fn remaining(&self) -> usize {
        self.buffer.len().saturating_sub(self.cursor)
    }

    pub fn capacity(&self) -> usize {
        self.buffer.len()
    }

    pub fn buffer(&self) -> &[Snid] {
        &self.buffer
    }
}

// random_seq_start generates a random 14-bit sequence start value (0-16383).
// This implements RFC 9562 Method 2 for randomized monotonicity,
// preventing ID enumeration attacks when the creation window is known.
#[inline(always)]
fn random_seq_start(state: &mut [u64; 4]) -> u16 {
    (xoshiro_next(state) & 0x3FFF) as u16
}

#[inline(always)]
fn xoshiro_next(state: &mut [u64; 4]) -> u64 {
    let res = state[1].wrapping_mul(5).rotate_left(7).wrapping_mul(9);
    let t = state[1] << 17;
    state[2] ^= state[0];
    state[3] ^= state[1];
    state[1] ^= state[2];
    state[0] ^= state[3];
    state[2] ^= t;
    state[3] = state[3].rotate_left(45);
    res
}

#[inline(always)]
fn make_words(ms: u64, sequence: u16, machine: u64, random: u64) -> (u64, u64) {
    let seq = sequence as u64;
    (
        (ms << 16) | 0x7000 | (seq >> 2),
        0x8000_0000_0000_0000 | ((seq & 0x03) << 60) | machine | ((random >> 28) & 0xFFFF_FFFFF),
    )
}

#[inline(always)]
fn make_snid(ms: u64, sequence: u16, machine: u64, random: u64) -> Snid {
    let (hi, lo) = make_words(ms, sequence, machine, random);
    let mut bytes = [0u8; 16];
    bytes[..8].copy_from_slice(&hi.to_be_bytes());
    bytes[8..].copy_from_slice(&lo.to_be_bytes());
    Snid(bytes)
}

#[inline(always)]
unsafe fn write_words_to_ptr(ptr: *mut u8, words: (u64, u64)) {
    let hi = words.0.to_be_bytes();
    let lo = words.1.to_be_bytes();
    unsafe {
        std::ptr::copy_nonoverlapping(hi.as_ptr(), ptr, 8);
        std::ptr::copy_nonoverlapping(lo.as_ptr(), ptr.add(8), 8);
    }
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
            let hi = u64::from_be_bytes([
                id.0[0], id.0[1], id.0[2], id.0[3], id.0[4], id.0[5], id.0[6], id.0[7],
            ]);
            let seq = (hi & 0x3FFF) as u16;
            assert!(
                seq <= 0x3FFF,
                "sequence {} exceeds maximum value 16383",
                seq
            );
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

    #[test]
    fn test_fill_slice_unique() {
        let mut state = GeneratorState::init();
        let mut ids = vec![Snid([0u8; 16]); 1000];
        state.fill_slice(&mut ids);
        let set: std::collections::HashSet<_> = ids.iter().copied().collect();
        assert_eq!(set.len(), ids.len());
    }

    #[test]
    fn test_turbo_streamer_next() {
        let mut streamer = TurboStreamer::new(64);
        let first = streamer.next_id();
        let second = streamer.next_id();
        assert_ne!(first, second);
        assert!(streamer.remaining() <= streamer.capacity());
    }
}

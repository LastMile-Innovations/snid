package snid

import "encoding/binary"

// NewBurst generates 'n' IDs using the unified shard system.
func NewBurst(n int) []ID {
	if n <= 0 {
		return nil
	}
	ids := make([]ID, n)

	s := nextShard()

	s.mu.Lock()
	ms := coarseClock.value.Load()
	if ms > s.lastMS {
		s.lastMS = ms
		s.sequence = 0
	}

	seq := s.sequence
	mid := uint64(s.machineID)
	s0, s1, s2, s3 := s.s0, s.s1, s.s2, s.s3

	for i := range ids {
		seq++
		if seq > 0x3FFF {
			seq = 0
			s.lastMS++
			ms = s.lastMS
		}

		res := rotl(s1*5, 7) * 9
		t := s1 << 17
		s2 ^= s0
		s3 ^= s1
		s1 ^= s2
		s0 ^= s3
		s2 ^= t
		s3 = rotl(s3, 45)

		binary.BigEndian.PutUint64(ids[i][:8], (ms<<16)|0x7000|(uint64(seq)>>2))
		binary.BigEndian.PutUint64(ids[i][8:], 0x8000000000000000|
			((uint64(seq)&0x03)<<60)|
			((mid&0xFFFFFF)<<36)|
			((res>>28)&0xFFFFFFFFF))
	}

	s.sequence = seq
	s.s0, s.s1, s.s2, s.s3 = s0, s1, s2, s3
	s.mu.Unlock()

	return ids
}

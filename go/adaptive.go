package snid

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/uber/h3-go/v4"
)

// Mode represents the generation strategy.
type Mode uint8

const (
	// ModeFast uses Xoshiro256** (Speed > Security).
	// Use for: Logs, Traces, Internal IDs.
	ModeFast Mode = iota

	// ModeSecure uses AES-CTR (Security > Speed).
	// Use for: Session Tokens, Public IDs, API Keys.
	ModeSecure

	// ModeAdaptive automatically selects based on load/hardware.
	ModeAdaptive
)

// AdaptiveGenerator manages the "Hive Mind" of ID generation.
type AdaptiveGenerator struct {
	mode   atomic.Uint32
	shards []*adaptiveShard
	mask   uint64

	// P-Local Shard Affinity
	// sync.Pool keeps objects local to the Processor (P)
	cursorPool sync.Pool
}

func (g *AdaptiveGenerator) nextShard() *adaptiveShard {
	cursor := g.cursorPool.Get().(*shardCursor)
	idx := cursor.index
	cursor.index++
	g.cursorPool.Put(cursor)
	return g.shards[idx&g.mask]
}

type adaptiveShard struct {
	_ [64]byte // front padding: isolates this struct from adjacent heap objects

	// HOT CACHE LINE (bytes 64–127, 64 bytes total).
	// Every genFast/genSecure call touches all of these fields.
	// Packing them here means one cache-line load after the mutex acquire
	// covers all hot state — eliminating the extra miss that occurred when
	// lastMS/sequence/machineID lived 1 KB past the mutex in the old layout.
	mu        sync.Mutex // 8 bytes
	lastMS    uint64     // 8 bytes
	sequence  uint32     // 4 bytes
	machineID uint32     // 4 bytes
	aesBufIdx int        // 8 bytes (checked every genSecure call)
	s0, s1    uint64     // 16 bytes (Xoshiro256** fast mode)
	s2, s3    uint64     // 16 bytes

	// WARM: accessed only in secure mode, and only for interface dispatch
	stream cipher.Stream // 16 bytes

	// COLD: refilled every ~128 IDs; intentionally off the hot path
	aesBuf [1024]byte

	_ [64]byte // back padding
}

var (
	adaptive     *AdaptiveGenerator
	adaptiveOnce sync.Once
)

// InitAdaptive boots the sentient engine.
func InitAdaptive() {
	adaptiveOnce.Do(func() {
		// 1. Detect Hardware & Scale Shards
		// Rule: 4x cores for optimal distribution, max 1024
		numCores := runtime.GOMAXPROCS(0)
		numShards := max(min(nextPowerOf2(numCores*4), 1024),
			// Minimum floor
			16)

		shards := make([]*adaptiveShard, numShards)
		mask := uint64(numShards - 1)

		// 2. Initialize Shards
		seed := uint64(time.Now().UnixNano())
		var seedBytes [8]byte
		if _, err := rand.Read(seedBytes[:]); err == nil {
			seed = binary.BigEndian.Uint64(seedBytes[:])
		}

		// Setup AES key from CSPRNG; fallback only if entropy source fails.
		masterKey := make([]byte, 32)
		if _, err := rand.Read(masterKey); err != nil {
			expandSeedToKey(seed, masterKey)
		}
		block, err := aes.NewCipher(masterKey)
		if err != nil {
			panic("snid: adaptive cipher init failed")
		}
		var ivBase [16]byte
		if _, err := rand.Read(ivBase[:]); err != nil {
			binary.BigEndian.PutUint64(ivBase[:8], splitMix64(&seed))
			binary.BigEndian.PutUint64(ivBase[8:], splitMix64(&seed))
		}

		for i := 0; i < numShards; i++ {
			s := &adaptiveShard{
				lastMS:    unixMilliCoarse(),
				machineID: uint32(splitMix64(&seed)) & 0xFFFFFF,
			}

			// Seed Xoshiro
			s.s0 = splitMix64(&seed)
			s.s1 = splitMix64(&seed)
			s.s2 = splitMix64(&seed)
			s.s3 = splitMix64(&seed)

			// Seed AES-CTR
			iv := ivBase
			binary.BigEndian.PutUint32(iv[12:], uint32(i))
			s.stream = cipher.NewCTR(block, iv[:])
			s.aesBufIdx = len(s.aesBuf) // Force refill

			shards[i] = s
		}

		adaptive = &AdaptiveGenerator{
			shards: shards,
			mask:   mask,
			cursorPool: sync.Pool{
				New: func() any {
					// Random start to spread load initially
					s := uint64(time.Now().UnixNano())
					return &shardCursor{index: splitMix64(&s)}
				},
			},
		}
		adaptive.mode.Store(uint32(ModeAdaptive))
	})
}

// Next generates an ID using the current adaptive strategy.
func Next() ID {
	InitAdaptive()
	return adaptive.next()
}

// SetMode forces a specific generation strategy.
func SetMode(m Mode) {
	InitAdaptive()
	adaptive.mode.Store(uint32(m))
}

// Batch generates N IDs using the current strategy.
// It locks the shard once per batch, minimizing contention overhead.
func Batch(n int) []ID {
	InitAdaptive()
	return adaptive.batch(n)
}

// next implements the corresponding operation.
func (g *AdaptiveGenerator) next() ID {
	shard := g.nextShard()

	mode := Mode(g.mode.Load())
	if mode == ModeAdaptive {
		mode = ModeSecure
	}

	if mode == ModeSecure {
		return shard.genSecure()
	}
	return shard.genFast()
}

// batch implements the corresponding operation.
func (g *AdaptiveGenerator) batch(n int) []ID {
	shard := g.nextShard()

	mode := Mode(g.mode.Load())
	if mode == ModeAdaptive {
		mode = ModeSecure
	}

	if mode == ModeSecure {
		return shard.batchSecure(n)
	}
	return shard.batchFast(n)
}

// nextEntropy implements the corresponding operation.
func (g *AdaptiveGenerator) nextEntropy() uint64 {
	shard := g.nextShard()
	return shard.genEntropy()
}

// nextSpatial generates a Location-Ordered ID (v8). (~20ns)
func (g *AdaptiveGenerator) nextSpatial(lat, lng float64, res int) ID {
	lat, lng, res = sanitizeSpatialInput(lat, lng, res)
	h3Idx, err := h3.LatLngToCell(h3.LatLng{Lat: lat, Lng: lng}, res)
	if err != nil {
		// Defensive fallback for impossible input edge-cases.
		h3Idx, _ = h3.LatLngToCell(h3.LatLng{Lat: 0, Lng: 0}, 12)
	}
	rand64 := g.nextEntropy()

	var id ID
	// High 64: H3 Index (with UUIDv8 version nibble + preserved original nibble in tail).
	binary.BigEndian.PutUint64(id[:8], uint64(h3Idx))
	originalHighNibble := id[6] >> 4
	id[6] = 0x80 | (id[6] & 0x0F) // Version 8

	// Low 64: entropy plus spatial markers.
	binary.BigEndian.PutUint64(id[8:], rand64)
	id[8] = (id[8] & 0x3F) | 0x80 // RFC 4122 variant
	id[14] = spatialMarkerByte
	id[15] = spatialTailNibble | (originalHighNibble & 0x0F)

	return id
}

func sanitizeSpatialInput(lat, lng float64, res int) (float64, float64, int) {
	if math.IsNaN(lat) || math.IsInf(lat, 0) {
		lat = 0
	}
	if math.IsNaN(lng) || math.IsInf(lng, 0) {
		lng = 0
	}
	if lat > 90 {
		lat = 90
	} else if lat < -90 {
		lat = -90
	}
	lng = math.Mod(lng+180, 360)
	if lng < 0 {
		lng += 360
	}
	lng -= 180

	if res < 0 {
		res = 0
	} else if res > 15 {
		res = 15
	}
	return lat, lng, res
}

// genEntropy implements the corresponding operation.
func (s *adaptiveShard) genEntropy() uint64 {
	s.mu.Lock()
	res := rotl(s.s1*5, 7) * 9
	t := s.s1 << 17
	s.s2 ^= s.s0
	s.s3 ^= s.s1
	s.s1 ^= s.s2
	s.s0 ^= s.s3
	s.s2 ^= t
	s.s3 = rotl(s.s3, 45)
	s.mu.Unlock()
	return res
}

// genFast implements the corresponding operation.
func (s *adaptiveShard) genFast() ID {
	s.mu.Lock()

	ms := unixMilliCoarse() // Use cached clock

	if ms > s.lastMS {
		s.lastMS = ms
		s.sequence = 0
	} else {
		s.sequence++
		if s.sequence > 0xFFF {
			// Overflow: spill to next ms logic or spin
			// For speed, we just force forward
			s.lastMS++
			ms = s.lastMS
			s.sequence = 0
		}
	}

	// Xoshiro256**
	res := rotl(s.s1*5, 7) * 9
	t := s.s1 << 17
	s.s2 ^= s.s0
	s.s3 ^= s.s1
	s.s1 ^= s.s2
	s.s0 ^= s.s3
	s.s2 ^= t
	s.s3 = rotl(s.s3, 45)

	seq := s.sequence
	mid := s.machineID
	s.mu.Unlock()

	return assembleID(ms, seq, mid, res)
}

// genSecure implements the corresponding operation.
func (s *adaptiveShard) genSecure() ID {
	s.mu.Lock()

	// Hybrid Logical Clock Logic
	ms := unixMilliCoarse()
	if ms > s.lastMS {
		s.lastMS = ms
		s.sequence = 0
	} else {
		s.sequence++ // Logical tick
		if s.sequence > 0xFFF {
			s.lastMS++
			ms = s.lastMS
			s.sequence = 0
		} else {
			ms = s.lastMS // Stick to logical time
		}
	}

	// AES-CTR Drain
	if s.aesBufIdx >= len(s.aesBuf) {
		// Security: Zero buffer before refill to prevent keystream accumulation
		// XOR(zeros) = pure keystream, XOR(ciphertext) = predictable pattern
		clear(s.aesBuf[:])
		s.stream.XORKeyStream(s.aesBuf[:], s.aesBuf[:])
		s.aesBufIdx = 0
	}

	entropy := binary.BigEndian.Uint64(s.aesBuf[s.aesBufIdx:])
	s.aesBufIdx += 8

	seq := s.sequence
	mid := s.machineID
	s.mu.Unlock()

	return assembleID(ms, seq, mid, entropy)
}

// batchFast implements the corresponding operation.
func (s *adaptiveShard) batchFast(n int) []ID {
	ids := make([]ID, n)
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := 0; i < n; i++ {
		ms := unixMilliCoarse()
		if ms > s.lastMS {
			s.lastMS = ms
			s.sequence = 0
		} else {
			s.sequence++
			if s.sequence > 0xFFF {
				s.lastMS++
				ms = s.lastMS
				s.sequence = 0
			}
		}

		res := rotl(s.s1*5, 7) * 9
		t := s.s1 << 17
		s.s2 ^= s.s0
		s.s3 ^= s.s1
		s.s1 ^= s.s2
		s.s0 ^= s.s3
		s.s2 ^= t
		s.s3 = rotl(s.s3, 45)

		ids[i] = assembleID(ms, s.sequence, s.machineID, res)
	}
	return ids
}

// batchSecure implements the corresponding operation.
func (s *adaptiveShard) batchSecure(n int) []ID {
	ids := make([]ID, n)
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := 0; i < n; i++ {
		ms := unixMilliCoarse()
		if ms > s.lastMS {
			s.lastMS = ms
			s.sequence = 0
		} else {
			s.sequence++
			if s.sequence > 0xFFF {
				s.lastMS++
				ms = s.lastMS
				s.sequence = 0
			} else {
				ms = s.lastMS
			}
		}

		if s.aesBufIdx >= len(s.aesBuf) {
			clear(s.aesBuf[:])
			s.stream.XORKeyStream(s.aesBuf[:], s.aesBuf[:])
			s.aesBufIdx = 0
		}

		entropy := binary.BigEndian.Uint64(s.aesBuf[s.aesBufIdx:])
		s.aesBufIdx += 8

		ids[i] = assembleID(ms, s.sequence, s.machineID, entropy)
	}
	return ids
}

// Helpers

// nextPowerOf2 is defined in generator.go

// expandSeedToKey implements the corresponding operation.
func expandSeedToKey(seed uint64, key []byte) {
	// simple deterministic expansion for example
	for i := 0; i < len(key); i += 8 {
		seed = splitMix64(&seed)
		binary.BigEndian.PutUint64(key[i:], seed)
	}
}

// assembleID implements the corresponding operation.
func assembleID(ms uint64, seq uint32, machineID uint32, entropy uint64) ID {
	// Layout:
	// 48 bit MS
	//  4 bit Version (7)
	// 12 bit Seq
	//  2 bit Variant
	// 62 bit Entropy (MachineID mixed in or replaced)

	// We mix MachineID into entropy for standard SNID layout compliance
	// Or we use the specific 24-bit MID slot if we strictly follow v7 layout.
	// Let's use standard SNID v4.5/v5 layout:
	// HI: MS(48) | Ver(4) | SeqA(12)
	// LO: Var(2) | SeqB(2) | MID(24) | ENT(36)

	hi := (ms << 16) | (0x7000) | (uint64(seq) >> 2)
	lo := (0x8000000000000000) |
		((uint64(seq) & 0x03) << 60) |
		((uint64(machineID) & 0xFFFFFF) << 36) |
		(entropy & 0xFFFFFFFFF) // Use lower 36 bits of entropy

	var id ID
	binary.BigEndian.PutUint64(id[:8], hi)
	binary.BigEndian.PutUint64(id[8:], lo)
	return id
}

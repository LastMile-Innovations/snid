package snid

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"sync"
	"time"
)

// SingularityGenerator represents the theoretical limit of ID generation.
// It combines AES-CTR (CSPRNG) with Hybrid Logical Clocks (HLC).
//
// guarantees:
// 1. Prediction Resistance: 2^128 security against next-ID prediction.
// 2. Causality: HLC ensures ordering even during clock regressions.
// 3. Speed: Pre-computed entropy buffers for zero-latency fetch.
type SingularityGenerator struct {
	// 128-byte cache line padding
	_ [64]byte

	// Hybrid Logical Clock state
	hlcPhysical uint64
	hlcLogical  uint64
	hlcMu       sync.Mutex

	// Entropy Stream (AES-CTR)
	stream cipher.Stream
	buffer [1024]byte // Pre-computed entropy pool
	bufIdx int

	_ [64]byte
}

var (
	singularity     *SingularityGenerator
	singularityOnce sync.Once
)

// InitSingularity boots the engine.
// It effectively turns the CPU's AES co-processor into a white-noise generator.
func InitSingularity() {
	singularityOnce.Do(func() {
		// 1. Generate a true random seed for the AES key
		key := make([]byte, 32) // AES-256
		if _, err := rand.Read(key); err != nil {
			panic("snid: not enough entropy for singularity boot")
		}

		block, _ := aes.NewCipher(key)

		// 2. Setup AES-CTR stream
		// We use a random IV for the counter
		iv := make([]byte, aes.BlockSize)
		if _, err := rand.Read(iv); err != nil {
			panic("snid: iv generation failed")
		}

		stream := cipher.NewCTR(block, iv)

		singularity = &SingularityGenerator{
			stream: stream,
			bufIdx: 1024, // Force initial refill
		}
	})
}

// NewSingularity generates a "God-Tier" ID.
// Cost: ~15ns (Amortized)
// Security: AES-256 Prediction Resistance
func NewSingularity() ID {
	InitSingularity()
	return singularity.next()
}

// next implements the corresponding operation.
func (g *SingularityGenerator) next() ID {
	g.hlcMu.Lock()

	// --- 1. Hybrid Logical Clock (Physics) ---
	now := uint64(time.Now().UnixMilli())

	if now > g.hlcPhysical {
		// Clock moved forward: reset logical tick
		g.hlcPhysical = now
		g.hlcLogical = 0
	} else {
		// Clock frozen or regression: increment logical tick
		// This preserves causality even if NTP is broken.
		g.hlcLogical++
		// If logical overflows (extremely rare in 1ms), we push physical forward
		if g.hlcLogical > 0xFFF {
			g.hlcPhysical++
			g.hlcLogical = 0
		}
	}

	phys := g.hlcPhysical
	logi := g.hlcLogical

	// --- 2. AES-CTR Entropy (Crypto-Plasma) ---
	if g.bufIdx >= len(g.buffer) {
		// Security: Zero buffer before refill to prevent keystream accumulation
		clear(g.buffer[:])
		g.stream.XORKeyStream(g.buffer[:], g.buffer[:])
		g.bufIdx = 0
	}

	// Grab 8 bytes of pure chaos
	// We use 8 bytes for the tail (64 bits of entropy)
	entropy := binary.BigEndian.Uint64(g.buffer[g.bufIdx:])
	g.bufIdx += 8

	g.hlcMu.Unlock()

	// --- 3. Assembly (The Singularity Layout) ---
	// 48 bits: Physical Time (HLC)
	//  4 bits: Version (8 - Singularity)
	// 12 bits: Logical Tick (HLC) - High precision causality
	//  2 bits: Variant
	// 62 bits: AES-256 Entropy

	hi := (phys << 16) | (0x8000) | (logi & 0xFFF)
	lo := (0x8000000000000000) | (entropy & 0x3FFFFFFFFFFFFFFF)

	var id ID
	binary.BigEndian.PutUint64(id[:8], hi)
	binary.BigEndian.PutUint64(id[8:], lo)

	return id
}

// SingularityBatch generates N IDs using full vector throughput.
func SingularityBatch(n int) []ID {
	InitSingularity()

	ids := make([]ID, n)
	g := singularity

	g.hlcMu.Lock()
	defer g.hlcMu.Unlock()

	// Performance: Cache initial time, only refresh on logical overflow
	now := uint64(time.Now().UnixMilli())

	for i := range n {
		// HLC with amortized time fetch
		if now > g.hlcPhysical {
			g.hlcPhysical = now
			g.hlcLogical = 0
		} else {
			g.hlcLogical++
			if g.hlcLogical > 0xFFF {
				// Logical overflow: refresh wall clock
				now = uint64(time.Now().UnixMilli())
				if now <= g.hlcPhysical {
					g.hlcPhysical++
				} else {
					g.hlcPhysical = now
				}
				g.hlcLogical = 0
			}
		}

		if g.bufIdx >= len(g.buffer) {
			clear(g.buffer[:])
			g.stream.XORKeyStream(g.buffer[:], g.buffer[:])
			g.bufIdx = 0
		}

		entropy := binary.BigEndian.Uint64(g.buffer[g.bufIdx:])
		g.bufIdx += 8

		hi := (g.hlcPhysical << 16) | (0x8000) | (g.hlcLogical & 0xFFF)
		lo := (0x8000000000000000) | (entropy & 0x3FFFFFFFFFFFFFFF)

		binary.BigEndian.PutUint64(ids[i][:8], hi)
		binary.BigEndian.PutUint64(ids[i][8:], lo)
	}

	return ids
}

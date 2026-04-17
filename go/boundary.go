package snid

import (
	"crypto/sha256"
	"encoding/binary"
)

// LLMFormatV1 is the canonical AI-facing projection for corpus and model pipelines.
type LLMFormatV1 struct {
	Atom            Atom   `json:"atom"`
	TimestampMillis int64  `json:"timestamp_millis"`
	MachineOrShard  uint32 `json:"machine_or_shard"`
	Sequence        uint16 `json:"sequence"`
}

// ToTensorWords exports the 128-bit ID as two big-endian int64 words.
func (id ID) ToTensorWords() (int64, int64) {
	return int64(binary.BigEndian.Uint64(id[:8])), int64(binary.BigEndian.Uint64(id[8:]))
}

// Sequence returns the full 14-bit sequence reconstructed from the split layout.
func (id ID) Sequence() uint16 {
	hi := uint16(binary.BigEndian.Uint64(id[:8]) & 0x0FFF)
	lo := uint16((binary.BigEndian.Uint64(id[8:]) >> 60) & 0x03)
	return (hi << 2) | lo
}

// MachineOrShard returns the stored 24-bit machine or projected shard field.
func (id ID) MachineOrShard() uint32 {
	return uint32((binary.BigEndian.Uint64(id[8:]) >> 36) & 0xFFFFFF)
}

// ToLLMFormat exports the canonical AI-readable decomposition for this ID.
func (id ID) ToLLMFormat(atom Atom) LLMFormatV1 {
	return LLMFormatV1{
		Atom:            CanonicalAtom(atom),
		TimestampMillis: id.Time().UnixMilli(),
		MachineOrShard:  id.MachineOrShard(),
		Sequence:        id.Sequence(),
	}
}

// NewDeterministicIngestID preserves timestamp locality while deriving the low bits from a stable hash.
func NewDeterministicIngestID(unixMillis uint64, contentHash []byte) ID {
	if len(contentHash) == 0 {
		contentHash = []byte{0}
	}
	sum := sha256.Sum256(contentHash)
	seq := binary.BigEndian.Uint16(sum[0:2]) & 0x3FFF
	machine := uint32(sum[2])<<16 | uint32(sum[3])<<8 | uint32(sum[4])
	entropy := (uint64(sum[5])<<32 | uint64(sum[6])<<24 | uint64(sum[7])<<16 | uint64(sum[8])<<8 | uint64(sum[9])) & 0xFFFFFFFFF
	entropy &^= ghostBitMask

	var id ID
	binary.BigEndian.PutUint64(id[:8], (unixMillis<<16)|0x7000|(uint64(seq)>>2))
	binary.BigEndian.PutUint64(id[8:], 0x8000000000000000|
		((uint64(seq)&0x03)<<60)|
		((uint64(machine)&0xFFFFFF)<<36)|
		entropy)
	return id
}

// TensorTimeDeltaMillis extracts timestamps from two tensor words and returns the millisecond delta.
func TensorTimeDeltaMillis(leftHi int64, rightHi int64) int64 {
	return (leftHi >> 16) - (rightHi >> 16)
}

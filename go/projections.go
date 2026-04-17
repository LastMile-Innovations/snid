package snid

import (
	"encoding/binary"

	"github.com/uber/h3-go/v4"
)

const ghostBitMask uint64 = 1 << 35

// LLMFormatV2 extends the minimal AI projection with topology and masking metadata.
type LLMFormatV2 struct {
	Kind            string `json:"kind"`
	Atom            Atom   `json:"atom"`
	TimestampMillis int64  `json:"timestamp_millis,omitempty"`
	SpatialAnchor   uint64 `json:"spatial_anchor,omitempty"`
	MachineOrShard  uint32 `json:"machine_or_shard,omitempty"`
	Sequence        uint16 `json:"sequence,omitempty"`
	Ghosted         bool   `json:"ghosted"`
}

// ToLLMFormatV2 exports an AI-facing projection that preserves time or spatial topology.
func (id ID) ToLLMFormatV2(atom Atom) LLMFormatV2 {
	out := LLMFormatV2{
		Kind:    "snid",
		Atom:    CanonicalAtom(atom),
		Ghosted: id.IsGhosted(),
	}
	if id.IsSpatial() {
		out.Kind = "sgid"
		out.SpatialAnchor = uint64(id.H3Cell())
		return out
	}
	out.TimestampMillis = id.Time().UnixMilli()
	out.MachineOrShard = id.MachineOrShard()
	out.Sequence = id.Sequence()
	return out
}

// TimeBin truncates the embedded timestamp to the requested millisecond resolution.
func (id ID) TimeBin(resolutionMillis int64) int64 {
	ts := id.Time().UnixMilli()
	if resolutionMillis <= 1 {
		return ts
	}
	return (ts / resolutionMillis) * resolutionMillis
}

// IsGhosted reports whether the reserved tombstone bit is set.
func (id ID) IsGhosted() bool {
	return (binary.BigEndian.Uint64(id[8:]) & ghostBitMask) != 0
}

// WithGhostBit returns a copy with the reserved tombstone bit enabled or disabled.
func (id ID) WithGhostBit(enabled bool) ID {
	lo := binary.BigEndian.Uint64(id[8:])
	if enabled {
		lo |= ghostBitMask
	} else {
		lo &^= ghostBitMask
	}
	binary.BigEndian.PutUint64(id[8:], lo)
	return id
}

// H3FeatureVector expands a spatial anchor into its parent chain from coarse to fine.
func (id ID) H3FeatureVector() []uint64 {
	cell := id.H3Cell()
	if cell == 0 {
		return nil
	}
	res := cell.Resolution()
	out := make([]uint64, 0, res+1)
	for r := 0; r <= res; r++ {
		parent, err := cell.Parent(r)
		if err != nil {
			return out
		}
		out = append(out, uint64(parent))
	}
	return out
}

// EncodeFixed64Pair encodes two int64 values as big-endian fixed64 words.
func EncodeFixed64Pair(hi, lo int64) [16]byte {
	var out [16]byte
	binary.BigEndian.PutUint64(out[:8], uint64(hi))
	binary.BigEndian.PutUint64(out[8:], uint64(lo))
	return out
}

// DecodeFixed64Pair decodes two big-endian fixed64 words.
func DecodeFixed64Pair(raw []byte) (int64, int64, error) {
	if len(raw) != 16 {
		return 0, 0, ErrInvalidLength
	}
	return int64(binary.BigEndian.Uint64(raw[:8])), int64(binary.BigEndian.Uint64(raw[8:])), nil
}

// H3FeatureVectorFromCell expands a raw H3 cell into its hierarchy path.
func H3FeatureVectorFromCell(cell h3.Cell) []uint64 {
	id := NewSpatialFromCell(uint64(cell), 0x8000000000000000)
	return id.H3FeatureVector()
}

package snid

import (
	"encoding/hex"
	"sort"

	"github.com/neighbor/snid/go/internal/vector"
)

// NeuralID is a 256-bit composite key.
type NeuralID [32]byte

// NewNeural creates a Neural ID (Standard ID + Binary Vector).
func NewNeural(base ID, vec []float32) (NeuralID, error) {
	var nid NeuralID
	copy(nid[:16], base[:]) // Head

	// Tail: Binary Quantization of first 128 dimensions using SIMD
	quantized := vector.QuantizeFloat32ToBinary(vec)
	copy(nid[16:], quantized[:])

	return nid, nil
}

// NewNeuralDeterministic preserves the timestamped head while installing an explicit semantic tail.
func NewNeuralDeterministic(unixMillis uint64, contentHash []byte, semanticHash [16]byte) NeuralID {
	return NewNeuralFromHash(NewDeterministicIngestID(unixMillis, contentHash), semanticHash)
}

// Distance returns Hamming Distance (Similarity).
func (n NeuralID) Distance(other NeuralID) int {
	return vector.HammingDistance(n.SemanticHash(), other.SemanticHash())
}

// HammingDistance is an alias for Distance.
func (n NeuralID) HammingDistance(other NeuralID) int {
	return n.Distance(other)
}

// Similarity returns 1.0 for identical hashes and 0.0 for opposite hashes.
func (n NeuralID) Similarity(other NeuralID) float64 {
	return 1.0 - (float64(n.Distance(other)) / 128.0)
}

// IsSimilar reports whether two neural IDs are within maxDistance bits.
func (n NeuralID) IsSimilar(other NeuralID, maxDistance int) bool {
	return n.Distance(other) <= maxDistance
}

// Head returns the 128-bit SNID causality head.
func (n NeuralID) Head() ID {
	var id ID
	copy(id[:], n[:16])
	return id
}

// SemanticHash returns the 128-bit semantic tail.
func (n NeuralID) SemanticHash() [16]byte {
	var out [16]byte
	copy(out[:], n[16:])
	return out
}

// ToTensor256Words exports the full 256-bit neural ID as four big-endian int64 words.
func (n NeuralID) ToTensor256Words() (int64, int64, int64, int64) {
	return int64FromBytes(n[0:8]), int64FromBytes(n[8:16]), int64FromBytes(n[16:24]), int64FromBytes(n[24:32])
}

// BatchHammingDistance computes distances between one query and many candidates.
// The underlying vector path uses SIMD on supported CPUs.
func BatchHammingDistance(query NeuralID, candidates []NeuralID) []int {
	if len(candidates) == 0 {
		return nil
	}
	qHash := query.SemanticHash()
	out := make([]int, len(candidates))
	for i := range candidates {
		out[i] = vector.HammingDistance(qHash, candidates[i].SemanticHash())
	}
	return out
}

// FindSimilar filters candidates by Hamming distance threshold.
func FindSimilar(query NeuralID, candidates []NeuralID, maxDistance int) []NeuralID {
	if len(candidates) == 0 {
		return nil
	}

	qHash := query.SemanticHash()
	out := make([]NeuralID, 0, len(candidates))
	for i := range candidates {
		if vector.HammingDistance(qHash, candidates[i].SemanticHash()) <= maxDistance {
			out = append(out, candidates[i])
		}
	}
	return out
}

// TopK returns the nearest K candidates by Hamming distance.
func TopK(query NeuralID, candidates []NeuralID, k int) []NeuralID {
	if k <= 0 || len(candidates) == 0 {
		return nil
	}

	type scored struct {
		id   NeuralID
		dist int
	}
	qHash := query.SemanticHash()
	scoredIDs := make([]scored, len(candidates))
	for i := range candidates {
		scoredIDs[i] = scored{
			id:   candidates[i],
			dist: vector.HammingDistance(qHash, candidates[i].SemanticHash()),
		}
	}
	sort.Slice(scoredIDs, func(i, j int) bool {
		if scoredIDs[i].dist == scoredIDs[j].dist {
			return scoredIDs[i].id.String() < scoredIDs[j].id.String()
		}
		return scoredIDs[i].dist < scoredIDs[j].dist
	})
	if k > len(scoredIDs) {
		k = len(scoredIDs)
	}
	out := make([]NeuralID, k)
	for i := 0; i < k; i++ {
		out[i] = scoredIDs[i].id
	}
	return out
}

// String encodes NeuralID bytes as lowercase hexadecimal.
func (n NeuralID) String() string {
	return hex.EncodeToString(n[:])
}

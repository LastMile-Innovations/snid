package vector

import (
	"math"
)

// DotInt8 computes the dot product of two int8 vectors.
func DotInt8(a, b []int8) int32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	return dotInt8(a, b)
}

// CosineSimilarityInt8 computes the cosine similarity between two int8 vectors.
func CosineSimilarityInt8(a, b []int8) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	dot := DotInt8(a, b)
	normA := DotInt8(a, a)
	normB := DotInt8(b, b)

	if normA == 0 || normB == 0 {
		return 0
	}

	return float32(float64(dot) / (math.Sqrt(float64(normA)) * math.Sqrt(float64(normB))))
}

// L2DistanceSquaredInt8 computes the squared L2 distance between two int8 vectors.
func L2DistanceSquaredInt8(a, b []int8) int32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	return l2DistanceSquaredInt8(a, b)
}

// HammingDistance computes the Hamming distance between two 16-byte blocks.
func HammingDistance(a, b [16]byte) int {
	return hammingDistance(a, b)
}

// BatchHammingDistance computes the Hamming distances between a query and multiple candidates.
func BatchHammingDistance(query [16]byte, candidates [][16]byte) []int {
	return batchHammingDistance(query, candidates)
}

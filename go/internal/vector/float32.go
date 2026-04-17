package vector

import (
	"math"
)

// DotFloat32 computes the dot product of two float32 vectors.
func DotFloat32(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	return dotFloat32(a, b)
}

// CosineSimilarityFloat32 computes the cosine similarity between two float32 vectors.
func CosineSimilarityFloat32(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	dot := DotFloat32(a, b)
	normA := DotFloat32(a, a)
	normB := DotFloat32(b, b)

	if normA == 0 || normB == 0 {
		return 0
	}

	return float32(float64(dot) / (math.Sqrt(float64(normA)) * math.Sqrt(float64(normB))))
}

// QuantizeFloat32ToBinary converts a float32 vector to a 128-bit binary quantized vector (16 bytes).
// Elements > 0 become 1, elements <= 0 become 0.
// Only the first 128 dimensions are processed.
func QuantizeFloat32ToBinary(vector []float32) [16]byte {
	return quantizeFloat32ToBinary(vector)
}

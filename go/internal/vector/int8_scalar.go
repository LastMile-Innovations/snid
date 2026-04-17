package vector

func dotInt8(a, b []int8) int32 {
	var sum int32
	for i := range a {
		sum += int32(a[i]) * int32(b[i])
	}
	return sum
}

func l2DistanceSquaredInt8(a, b []int8) int32 {
	var sum int32
	for i := range a {
		diff := int32(a[i]) - int32(b[i])
		sum += diff * diff
	}
	return sum
}

func hammingDistance(a, b [16]byte) int {
	dist := 0
	for i := 0; i < 16; i++ {
		xor := a[i] ^ b[i]
		// Kernighan's bit counting
		for xor > 0 {
			xor &= (xor - 1)
			dist++
		}
	}
	return dist
}

func batchHammingDistance(query [16]byte, candidates [][16]byte) []int {
	out := make([]int, len(candidates))
	for i := range candidates {
		out[i] = hammingDistance(query, candidates[i])
	}
	return out
}

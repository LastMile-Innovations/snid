package vector

func dotFloat32(a, b []float32) float32 {
	var sum float32
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

func quantizeFloat32ToBinary(vector []float32) [16]byte {
	var out [16]byte
	for i := 0; i < 128 && i < len(vector); i++ {
		if vector[i] > 0 {
			out[i/8] |= (1 << (7 - (i % 8)))
		}
	}
	return out
}

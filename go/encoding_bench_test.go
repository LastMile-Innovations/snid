package snid

import (
	"bytes"
	"testing"
)

var (
	benchBytes []byte
	benchErr   error
)

func BenchmarkBase58Bytes(b *testing.B) {
	cases := []struct {
		name string
		raw  []byte
	}{
		{name: "encode_8_bytes", raw: []byte{0, 1, 2, 3, 4, 5, 6, 7}},
		{name: "encode_24_bytes", raw: bytes.Repeat([]byte{0x5a}, 24)},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchString = encodeBase58Bytes(tc.raw)
			}
		})

		encoded := encodeBase58Bytes(tc.raw)
		b.Run("decode_"+tc.name[len("encode_"):], func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchBytes, benchErr = decodeBase58Bytes(encoded)
				if benchErr != nil {
					b.Fatalf("decodeBase58Bytes: %v", benchErr)
				}
			}
		})
	}
}

func BenchmarkSNIDWireEncoding(b *testing.B) {
	id := NewFast()
	wire := id.String(Matter)
	compact := id.StringCompact()
	dst := make([]byte, 0, MaxAtomLength+1+MaxPayloadLength)

	b.Run("String", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchString = id.String(Matter)
		}
	})

	b.Run("StringCompact", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchString = id.StringCompact()
		}
	})

	b.Run("AppendTo", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dst = dst[:0]
			dst = id.AppendTo(dst, Matter)
			benchBytes = dst
		}
	})

	b.Run("FromString", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchID, benchAtom, benchErr = FromString(wire)
			if benchErr != nil {
				b.Fatalf("FromString: %v", benchErr)
			}
		}
	})

	b.Run("ParseCompact", func(b *testing.B) {
		var parsed ID
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchErr = parsed.ParseCompact(compact)
			if benchErr != nil {
				b.Fatalf("ParseCompact: %v", benchErr)
			}
			benchID = parsed
		}
	})
}

func BenchmarkAKIDEncoding(b *testing.B) {
	tenantID := NewTenant().TenantString()
	publicID := NewAKIDPublic(tenantID)
	rawSecret := bytes.Repeat([]byte{0x33}, akidSecretBytes)
	secret := EncodeAKIDSecret(rawSecret)
	wire := FormatAKID(publicID, secret)

	b.Run("EncodeAKIDSecret", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchString = EncodeAKIDSecret(rawSecret)
		}
	})

	b.Run("VerifyAKIDSecretChecksum", func(b *testing.B) {
		var ok bool
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchBytes, ok = VerifyAKIDSecretChecksum(secret)
			if !ok {
				b.Fatal("VerifyAKIDSecretChecksum failed")
			}
			benchBool = ok
		}
	})

	b.Run("ParseAKID", func(b *testing.B) {
		var secretOut string
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchID, secretOut, benchErr = ParseAKID(wire)
			if benchErr != nil {
				b.Fatalf("ParseAKID: %v", benchErr)
			}
			benchString = secretOut
		}
	})
}

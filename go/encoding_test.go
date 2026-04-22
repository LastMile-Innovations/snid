package snid

import (
	"bytes"
	"strings"
	"testing"
)

// TestEncodingFunctions tests encoding.go functions for coverage
func TestEncodingFunctions(t *testing.T) {
	// Test base58 encoding/decoding
	id := NewFast()

	// Test AppendTo
	buf := make([]byte, 0, 32)
	buf = id.AppendTo(buf, Matter)
	if len(buf) == 0 {
		t.Fatal("expected non-empty buffer after AppendTo")
	}

	// Test StringCompact
	compact := id.StringCompact()
	if len(compact) == 0 {
		t.Fatal("expected non-empty compact representation")
	}

	// Test ParseCompact
	var parsed ID
	err := parsed.ParseCompact(compact)
	if err != nil {
		t.Fatalf("ParseCompact failed: %v", err)
	}
	if parsed != id {
		t.Fatal("parsed ID mismatch")
	}

	// Test invalid compact format
	err = parsed.ParseCompact("invalid")
	if err == nil {
		t.Fatal("expected error for invalid compact format")
	}
}

// TestBase58Encoding tests base58 encoding functions for coverage
func TestBase58Encoding(t *testing.T) {
	tests := []struct {
		name string
		raw  []byte
		want string
	}{
		{name: "empty", raw: nil, want: ""},
		{name: "single zero", raw: []byte{0x00}, want: "1"},
		{name: "three zeroes", raw: []byte{0x00, 0x00, 0x00}, want: "111"},
		{name: "leading zeroes", raw: []byte{0x00, 0x00, 0x01}, want: "112"},
		{name: "hello world", raw: []byte("Hello World!"), want: "2NEpo7TZRRrLZSi2U"},
		{name: "short id bytes", raw: []byte{0, 1, 2, 3, 4, 5, 6, 7}, want: "13DUyZY2dc"},
		{name: "akid secret bytes", raw: bytes.Repeat([]byte{0x5a}, 24), want: "9Ek46doqep1srpD1W4QaovLWwgPir9Xhb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodeBase58Bytes(tt.raw)
			if got != tt.want {
				t.Fatalf("encodeBase58Bytes mismatch: got %q want %q", got, tt.want)
			}
			decoded, err := decodeBase58Bytes(got)
			if err != nil {
				t.Fatalf("decodeBase58Bytes(%q): %v", got, err)
			}
			if !bytes.Equal(decoded, tt.raw) {
				t.Fatalf("decoded bytes mismatch: got %x want %x", decoded, tt.raw)
			}
		})
	}

	if _, err := decodeBase58Bytes("!!!"); err == nil {
		t.Fatal("expected error for invalid base58")
	}
}

func TestSNIDCompactPayloadRegression(t *testing.T) {
	compact := Zero.StringCompact()
	if want := strings.Repeat("1", 17); compact != want {
		t.Fatalf("zero compact mismatch: got %q want %q", compact, want)
	}
	var parsed ID
	if err := parsed.ParseCompact(compact); err != nil {
		t.Fatalf("ParseCompact zero: %v", err)
	}
	if parsed != Zero {
		t.Fatal("zero compact parsed to non-zero ID")
	}

	var max ID
	for i := range max {
		max[i] = 0xFF
	}
	maxCompact := max.StringCompact()
	if len(maxCompact) != 23 {
		t.Fatalf("max compact length mismatch: got %d want 23", len(maxCompact))
	}
	if err := parsed.ParseCompact(maxCompact); err != nil {
		t.Fatalf("ParseCompact max: %v", err)
	}
	if parsed != max {
		t.Fatal("max compact parsed mismatch")
	}

	var leading ID
	leading[2] = 1
	leadingCompact := leading.StringCompact()
	if !strings.HasPrefix(leadingCompact, "11") {
		t.Fatalf("expected two canonical leading ones, got %q", leadingCompact)
	}
	if err := parsed.ParseCompact(leadingCompact); err != nil {
		t.Fatalf("ParseCompact leading zeroes: %v", err)
	}
	if parsed != leading {
		t.Fatal("leading-zero compact parsed mismatch")
	}
	if err := parsed.ParseCompact("1" + leadingCompact); err != ErrInvalidFormat {
		t.Fatalf("expected ErrInvalidFormat for non-canonical leading one, got %v", err)
	}
}

// Test16Base58Encoding tests 16-byte base58 encoding functions for coverage
func Test16Base58Encoding(t *testing.T) {
	// Test encode16Base58
	var src [16]byte
	for i := range src {
		src[i] = byte(i)
	}
	encoded := encode16Base58(src)
	if len(encoded) == 0 {
		t.Fatal("expected non-empty 16-byte base58 encoding")
	}

	// Test decode16Base58Bytes
	decoded, err := decode16Base58Bytes(encoded)
	if err != nil {
		t.Fatalf("decode16Base58Bytes failed: %v", err)
	}
	if decoded != src {
		t.Fatal("decoded 16-byte array mismatch")
	}

	// Test invalid 16-byte base58
	_, err = decode16Base58Bytes("!!!")
	if err == nil {
		t.Fatal("expected error for invalid 16-byte base58")
	}
}

func TestAKIDSecretBase58RoundTrip(t *testing.T) {
	secretBytes := append([]byte{0, 0}, bytes.Repeat([]byte{0x42}, 22)...)
	secret := EncodeAKIDSecret(secretBytes)
	decoded, ok := VerifyAKIDSecretChecksum(secret)
	if !ok {
		t.Fatal("expected AKID secret checksum to verify")
	}
	if !bytes.Equal(decoded, secretBytes) {
		t.Fatalf("AKID decoded secret mismatch: got %x want %x", decoded, secretBytes)
	}
}

func TestShortIDBase58RoundTrip(t *testing.T) {
	var sid ShortID
	for i := range sid {
		sid[i] = byte(i)
	}
	wire := sid.String(Matter)
	if !strings.HasPrefix(wire, "MAT:") {
		t.Fatalf("ShortID wire prefix mismatch: got %q", wire)
	}
	decoded, err := decodeBase58Bytes(wire[len("MAT:"):])
	if err != nil {
		t.Fatalf("decode ShortID payload: %v", err)
	}
	if !bytes.Equal(decoded, sid[:]) {
		t.Fatalf("ShortID decoded bytes mismatch: got %x want %x", decoded, sid[:])
	}
}

func TestSNIDWireRoundTripRegression(t *testing.T) {
	id := NewFast()
	wire := id.String(Matter)
	parsed, atom, err := FromString(wire)
	if err != nil {
		t.Fatalf("FromString(%q): %v", wire, err)
	}
	if atom != Matter {
		t.Fatalf("atom mismatch: got %s want %s", atom, Matter)
	}
	if parsed != id {
		t.Fatal("SNID wire parsed mismatch")
	}
}

func FuzzBase58BytesRoundTrip(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0})
	f.Add([]byte{0, 0, 1})
	f.Add([]byte("Hello World!"))
	f.Add(bytes.Repeat([]byte{0x5a}, 24))

	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > 64 {
			t.Skip()
		}
		encoded := encodeBase58Bytes(raw)
		decoded, err := decodeBase58Bytes(encoded)
		if err != nil {
			t.Fatalf("decodeBase58Bytes(%q): %v", encoded, err)
		}
		if !bytes.Equal(decoded, raw) {
			t.Fatalf("base58 roundtrip mismatch: got %x want %x", decoded, raw)
		}
	})
}

func FuzzSNIDWireRoundTrip(f *testing.F) {
	f.Add(make([]byte, 16))
	f.Add(bytes.Repeat([]byte{0xFF}, 16))
	f.Add([]byte{0, 0, 1, 2, 3, 4, 5, 6, 0x80, 9, 10, 11, 12, 13, 14, 15})

	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) != 16 {
			t.Skip()
		}
		id, err := FromBytes(raw)
		if err != nil {
			t.Fatalf("FromBytes: %v", err)
		}
		wire := id.String(Matter)
		parsed, atom, err := FromString(wire)
		if err != nil {
			t.Fatalf("FromString(%q): %v", wire, err)
		}
		if atom != Matter {
			t.Fatalf("atom mismatch: got %s want %s", atom, Matter)
		}
		if parsed != id {
			t.Fatalf("SNID wire roundtrip mismatch: got %x want %x", parsed, id)
		}
	})
}

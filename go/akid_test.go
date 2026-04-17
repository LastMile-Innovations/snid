package snid

import (
	"bytes"
	"testing"
)

func TestAKIDRoundTrip(t *testing.T) {
	tenantID := NewTenant().TenantString()
	publicID := NewAKIDPublic(tenantID)
	secretBytes := bytes.Repeat([]byte{0x5a}, 24)
	secret := EncodeAKIDSecret(secretBytes)
	wire := FormatAKID(publicID, secret)

	parsedID, parsedSecret, err := ParseAKID(wire)
	if err != nil {
		t.Fatalf("ParseAKID returned error: %v", err)
	}
	if parsedID != publicID {
		t.Fatalf("public ID mismatch: got=%s want=%s", parsedID.AccessKeyString(), publicID.AccessKeyString())
	}
	if parsedSecret != secret {
		t.Fatalf("secret mismatch: got=%s want=%s", parsedSecret, secret)
	}
	if publicID.TenantHash() == 0 {
		t.Fatal("expected tenant hash to be embedded in AKID public ID")
	}
}

func TestVerifyAKIDSecretChecksum(t *testing.T) {
	secretBytes := bytes.Repeat([]byte{0x42}, 24)
	secret := EncodeAKIDSecret(secretBytes)
	decoded, ok := VerifyAKIDSecretChecksum(secret)
	if !ok {
		t.Fatal("expected valid secret checksum")
	}
	if !bytes.Equal(decoded, secretBytes) {
		t.Fatalf("decoded secret mismatch: got=%x want=%x", decoded, secretBytes)
	}

	invalid := secret[:len(secret)-1] + "1"
	if _, ok := VerifyAKIDSecretChecksum(invalid); ok {
		t.Fatal("expected invalid checksum to fail")
	}
}

func BenchmarkParseAKID(b *testing.B) {
	tenantID := NewTenant().TenantString()
	publicID := NewAKIDPublic(tenantID)
	secret := EncodeAKIDSecret(bytes.Repeat([]byte{0x33}, 24))
	wire := FormatAKID(publicID, secret)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := ParseAKID(wire); err != nil {
			b.Fatalf("ParseAKID returned error: %v", err)
		}
	}
}

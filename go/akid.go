package snid

import (
	"crypto/rand"
	"strings"

	"github.com/mr-tron/base58"
)

const akidSecretBytes = 24

// NewAKIDPublic returns a tenant-projected public SNID for AKID lookup and routing.
func NewAKIDPublic(tenantID string) ID {
	return NewAccessKeyIDForTenant(tenantID)
}

// EncodeAKIDSecret returns the Base58 payload plus a CRC8-derived check character.
func EncodeAKIDSecret(secret []byte) string {
	if len(secret) == 0 {
		return ""
	}
	body := base58.Encode(secret)
	chk := base58Alphabet[crc8(secret)%58]
	return body + string(chk)
}

// NewAKIDSecret returns a freshly generated 24-byte AKID secret string.
func NewAKIDSecret() (string, error) {
	raw := make([]byte, akidSecretBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return EncodeAKIDSecret(raw), nil
}

// VerifyAKIDSecretChecksum validates the AKID secret checksum and returns the decoded bytes.
func VerifyAKIDSecretChecksum(secret string) ([]byte, bool) {
	secret = strings.TrimSpace(secret)
	if len(secret) < 2 {
		return nil, false
	}
	body := secret[:len(secret)-1]
	chk := secret[len(secret)-1]
	decoded, err := base58.Decode(body)
	if err != nil {
		return nil, false
	}
	expected := base58Alphabet[crc8(decoded)%58]
	if chk != expected {
		return nil, false
	}
	return decoded, true
}

// ParseAKID parses the wire AKID format KEY:<public>_<secret>.
func ParseAKID(wire string) (publicID ID, secret string, err error) {
	wire = strings.TrimSpace(wire)
	if !strings.HasPrefix(wire, "KEY:") {
		return Zero, "", ErrInvalidFormat
	}
	parts := strings.SplitN(wire, "_", 2)
	if len(parts) != 2 {
		return Zero, "", ErrInvalidFormat
	}
	atom, parseErr := publicID.Parse(parts[0])
	if parseErr != nil {
		return Zero, "", parseErr
	}
	if atom != Key {
		return Zero, "", ErrInvalidAtom
	}
	if _, ok := VerifyAKIDSecretChecksum(parts[1]); !ok {
		return Zero, "", ErrChecksum
	}
	return publicID, parts[1], nil
}

// FormatAKID joins the public ID and secret into the canonical wire format.
func FormatAKID(publicID ID, secret string) string {
	return publicID.String(Key) + "_" + strings.TrimSpace(secret)
}

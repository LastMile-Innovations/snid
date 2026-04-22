package snid

import (
	"testing"
	"time"
)

// TestCompositeIDVerification tests composite ID verification for coverage
func TestCompositeIDVerification(t *testing.T) {
	head := NewFast()
	actor := NewIdentityID()
	resource := []byte("test-resource")
	capability := []byte("read")
	key := []byte("0123456789abcdef")

	// Test KID verification
	kid, err := NewKIDForCapability(head, actor, resource, capability, key)
	if err != nil {
		t.Fatalf("NewKIDForCapability failed: %v", err)
	}

	if !kid.Verify(actor, resource, capability, key) {
		t.Fatal("KID verify failed")
	}

	// Test KID verification with wrong capability
	if kid.Verify(actor, resource, []byte("write"), key) {
		t.Fatal("KID verify should fail for wrong capability")
	}

	// Test KID verification with empty key
	if kid.Verify(actor, resource, capability, []byte{}) {
		t.Fatal("KID verify should fail with empty key")
	}

	// Test KID verification with wrong resource
	if kid.Verify(actor, []byte("wrong"), capability, key) {
		t.Fatal("KID verify should fail for wrong resource")
	}
}

// TestLIDVerification tests LID verification for coverage
func TestLIDVerification(t *testing.T) {
	head := NewLedger()
	var prev LID
	copy(prev[:16], head[:])
	payload := []byte("test-payload")
	key := []byte("0123456789abcdef")

	lid, err := NewLIDWithHead(head, prev, payload, key)
	if err != nil {
		t.Fatalf("NewLIDWithHead failed: %v", err)
	}

	// Test verification
	if !lid.Verify(prev, payload, key) {
		t.Fatal("LID verify failed")
	}

	// Test verification with wrong payload
	if lid.Verify(prev, []byte("wrong"), key) {
		t.Fatal("LID verify should fail for wrong payload")
	}

	// Test verification with wrong key
	if lid.Verify(prev, payload, []byte("wrongkey")) {
		t.Fatal("LID verify should fail for wrong key")
	}

	// Test verification with empty key
	_, err = NewLID(prev, payload, []byte{})
	if err != ErrInvalidLIDKey {
		t.Fatal("expected ErrInvalidLIDKey for empty key")
	}
}

// TestLIDBLAKE3 tests BLAKE3 variant of LID for coverage
func TestLIDBLAKE3(t *testing.T) {
	head := NewLedger()
	var prev LID
	copy(prev[:16], head[:])
	payload := []byte("test-payload")
	key := []byte("0123456789abcdef")

	blake3LID, err := LIDBLAKE3(head, prev, payload, key)
	if err != nil {
		t.Fatalf("LIDBLAKE3 failed: %v", err)
	}

	if blake3LID.Head() != head {
		t.Fatal("BLAKE3 LID head mismatch")
	}
}

// TestCompositeIDAccessors tests composite ID accessor methods for coverage
func TestCompositeIDAccessors(t *testing.T) {
	head := NewFast()
	var hash [16]byte
	for i := range hash {
		hash[i] = byte(i)
	}

	// Test WID accessors
	wid := NewWID(head, hash)
	if wid.Head() != head {
		t.Fatal("WID head mismatch")
	}
	if wid.ScenarioHash() != hash {
		t.Fatal("WID scenario hash mismatch")
	}

	// Test XID accessors
	var edge [16]byte
	for i := range edge {
		edge[i] = byte(16 - i)
	}
	xid := NewXID(head, edge)
	if xid.Head() != head {
		t.Fatal("XID head mismatch")
	}
	if xid.EdgeHash() != edge {
		t.Fatal("XID edge hash mismatch")
	}
}

// TestGrantIDVerification tests GrantID verification for coverage
func TestGrantIDVerification(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")

	// Test grant with expiration
	grant := NewGrant(Matter, time.Hour, secret)
	if grant.ID == Zero {
		t.Fatal("expected non-zero grant ID")
	}

	if !grant.Verify(secret) {
		t.Fatal("grant verify failed")
	}

	// Test grant with wrong secret
	if grant.Verify([]byte("wrongsecret12345678")) {
		t.Fatal("grant verify should fail for wrong secret")
	}

	// Test grant with invalid secret length
	_ = NewGrant(Matter, time.Hour, []byte("short"))

	// Test permanent grant (no expiration)
	permanentGrant := NewGrant(Tenant, 0, secret)
	if permanentGrant.ID == Zero {
		t.Fatal("expected non-zero permanent grant ID")
	}

	if !permanentGrant.Verify(secret) {
		t.Fatal("permanent grant verify failed")
	}

	// Test ParseGrant
	grantStr := grant.String(Matter)
	parsedGrant, atom, err := ParseGrant(grantStr, secret)
	if err != nil {
		t.Fatalf("ParseGrant failed: %v", err)
	}
	if atom != Matter {
		t.Fatal("grant atom mismatch")
	}
	if parsedGrant.ID != grant.ID {
		t.Fatal("parsed grant ID mismatch")
	}

	// Test ParseGrant with wrong secret
	_, _, err = ParseGrant(grantStr, []byte("wrong"))
	if err != ErrInvalidSignature {
		t.Fatal("expected ErrInvalidSignature for wrong secret")
	}
}

// TestAKIDFunctions tests AKID functions for coverage
func TestAKIDFunctions(t *testing.T) {
	tenantID := NewTenant().TenantString()
	publicID := NewAKIDPublic(tenantID)
	secretBytes := []byte("012345678901234567890123")

	// Test EncodeAKIDSecret
	secret := EncodeAKIDSecret(secretBytes)
	if secret == "" {
		t.Fatal("expected non-empty encoded secret")
	}

	// Test FormatAKID
	wire := FormatAKID(publicID, secret)
	if wire == "" {
		t.Fatal("expected non-empty AKID wire format")
	}

	// Test ParseAKID
	parsedID, parsedSecret, err := ParseAKID(wire)
	if err != nil {
		t.Fatalf("ParseAKID failed: %v", err)
	}
	if parsedID != publicID {
		t.Fatal("parsed AKID public ID mismatch")
	}
	if parsedSecret != secret {
		t.Fatal("parsed AKID secret mismatch")
	}

	// Test VerifyAKIDSecretChecksum
	decoded, ok := VerifyAKIDSecretChecksum(secret)
	if !ok {
		t.Fatal("expected valid secret checksum")
	}
	if len(decoded) != len(secretBytes) {
		t.Fatal("decoded secret length mismatch")
	}

	// Test VerifyAKIDSecretChecksum with invalid checksum
	invalid := secret[:len(secret)-1] + "1"
	if _, ok := VerifyAKIDSecretChecksum(invalid); ok {
		t.Fatal("expected invalid checksum to fail")
	}

	// Test TenantHash
	if publicID.TenantHash() == 0 {
		t.Fatal("expected non-zero tenant hash")
	}
}

// TestTenantID tests tenant ID functions for coverage
func TestTenantID(t *testing.T) {
	tenant := NewTenant()
	if tenant == Zero {
		t.Fatal("expected non-zero tenant ID")
	}

	// Test TenantString
	tenantStr := tenant.TenantString()
	if tenantStr == "" {
		t.Fatal("expected non-empty tenant string")
	}

	// Test TenantHash
	hash := tenant.TenantHash()
	if hash == 0 {
		t.Fatal("expected non-zero tenant hash")
	}
}

// TestEventID tests event ID functions for coverage
func TestEventID(t *testing.T) {
	event := NewEvent()
	if event == Zero {
		t.Fatal("expected non-zero event ID")
	}

	// Test EventString
	eventStr := event.EventString()
	if eventStr == "" {
		t.Fatal("expected non-empty event string")
	}
}

// TestSessionID tests session ID functions for coverage
func TestSessionID(t *testing.T) {
	session := NewSession()
	if session == Zero {
		t.Fatal("expected non-zero session ID")
	}

	// Test SessionString
	sessionStr := session.SessionString()
	if sessionStr == "" {
		t.Fatal("expected non-empty session string")
	}
}

// TestTimeID tests time ID functions for coverage
func TestTimeID(t *testing.T) {
	timeID := NewChronos()
	if timeID == Zero {
		t.Fatal("expected non-zero time ID")
	}

	// Test ChronosString
	timeStr := timeID.ChronosString()
	if timeStr == "" {
		t.Fatal("expected non-empty time string")
	}
}

// TestLegalID tests legal ID functions for coverage
func TestLegalID(t *testing.T) {
	legal := NewLegal()
	if legal == Zero {
		t.Fatal("expected non-zero legal ID")
	}

	// Test LegalString
	legalStr := legal.LegalString()
	if legalStr == "" {
		t.Fatal("expected non-empty legal string")
	}
}

// TestTrustID tests trust ID functions for coverage
func TestTrustID(t *testing.T) {
	trust := NewTrust()
	if trust == Zero {
		t.Fatal("expected non-zero trust ID")
	}

	// Test TrustString
	trustStr := trust.TrustString()
	if trustStr == "" {
		t.Fatal("expected non-empty trust string")
	}
}

// TestKineticID tests kinetic ID functions for coverage
func TestKineticID(t *testing.T) {
	kinetic := NewKinetic()
	if kinetic == Zero {
		t.Fatal("expected non-zero kinetic ID")
	}

	// Test KineticString
	kineticStr := kinetic.KineticString()
	if kineticStr == "" {
		t.Fatal("expected non-empty kinetic string")
	}
}

// TestCognitionID tests cognition ID functions for coverage
func TestCognitionID(t *testing.T) {
	cognition := NewCognition()
	if cognition == Zero {
		t.Fatal("expected non-zero cognition ID")
	}

	// Test CognitionString
	cognitionStr := cognition.CognitionString()
	if cognitionStr == "" {
		t.Fatal("expected non-empty cognition string")
	}
}

// TestSemanticID tests semantic ID functions for coverage
func TestSemanticID(t *testing.T) {
	semantic := NewSemantic()
	if semantic == Zero {
		t.Fatal("expected non-zero semantic ID")
	}

	// Test SemanticString
	semanticStr := semantic.SemanticString()
	if semanticStr == "" {
		t.Fatal("expected non-empty semantic string")
	}
}

// TestSystemID tests system ID functions for coverage
func TestSystemID(t *testing.T) {
	system := NewSystem()
	if system == Zero {
		t.Fatal("expected non-zero system ID")
	}

	// Test SystemString
	systemStr := system.SystemString()
	if systemStr == "" {
		t.Fatal("expected non-empty system string")
	}
}

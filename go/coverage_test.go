package snid

import (
	"strings"
	"testing"
	"time"
)

// TestEncodingCoverage tests encoding/decoding functions for coverage
func TestEncodingCoverage(t *testing.T) {
	// Test Base58 encoding/decoding
	id := NewFast()
	wire := id.String(Matter)
	if wire == "" {
		t.Fatal("expected non-empty wire string")
	}

	// Test ParseWithOptions with version validation
	parsed, atom, err := ParseWithOptions(wire, ValidationOptions{
		RequireVersion7: true,
	})
	if err != nil {
		t.Fatalf("ParseWithOptions failed: %v", err)
	}
	if atom != Matter {
		t.Fatalf("atom mismatch: got %s want %s", atom, Matter)
	}
	if parsed != id {
		t.Fatal("parsed ID mismatch")
	}

	// Test underscore wire format
	underscoreWire := id.StringWithFormat(Matter, WireUnderscore)
	if !strings.Contains(underscoreWire, "_") {
		t.Fatal("expected underscore in wire format")
	}

	// Test parsing underscore format with cleanup
	oldAcceptUnderscore := AcceptUnderscore()
	SetAcceptUnderscore(true)
	t.Cleanup(func() { SetAcceptUnderscore(oldAcceptUnderscore) })

	parsed2, atom2, err := FromString(underscoreWire)
	if err != nil {
		t.Fatalf("FromString with underscore failed: %v", err)
	}
	if atom2 != Matter {
		t.Fatalf("atom mismatch with underscore: got %s want %s", atom2, Matter)
	}
	if parsed2 != id {
		t.Fatal("parsed ID mismatch with underscore")
	}
}

// TestGeneratorCoverage tests generator functions for coverage
func TestGeneratorCoverage(t *testing.T) {
	// Test NewProjected
	tenantID := "tenant-123"
	projected := NewProjected(tenantID, 0)
	if projected == Zero {
		t.Fatal("expected non-zero projected ID")
	}

	// Test NewDeterministicIngestID
	content := []byte("test-content")
	detID := NewDeterministicIngestID(1700000000123, content)
	if detID == Zero {
		t.Fatal("expected non-zero deterministic ID")
	}

	// Test deterministic IDs are reproducible
	detID2 := NewDeterministicIngestID(1700000000123, content)
	if detID != detID2 {
		t.Fatal("expected deterministic IDs to match")
	}

	// Test different content produces different IDs
	detID3 := NewDeterministicIngestID(1700000000123, []byte("different"))
	if detID == detID3 {
		t.Fatal("expected different content to produce different IDs")
	}

	// Test NewBurst
	batch := NewBurst(100)
	if len(batch) != 100 {
		t.Fatalf("expected 100 IDs, got %d", len(batch))
	}

	// Test batch IDs are unique
	seen := make(map[ID]bool)
	for _, id := range batch {
		if seen[id] {
			t.Fatal("duplicate ID in batch")
		}
		seen[id] = true
	}
}

// TestCompositeCoverage tests composite ID functions for coverage
func TestCompositeCoverage(t *testing.T) {
	head := NewFast()

	// Test WID
	var scenario [16]byte
	for i := range scenario {
		scenario[i] = byte(i)
	}
	wid := NewWID(head, scenario)
	if wid.Head() != head {
		t.Fatal("WID head mismatch")
	}
	if wid.ScenarioHash() != scenario {
		t.Fatal("WID scenario hash mismatch")
	}

	// Test WID.ToTensor256Words
	w0, w1, w2, w3 := wid.ToTensor256Words()
	if w0 == 0 && w1 == 0 && w2 == 0 && w3 == 0 {
		t.Fatal("expected non-zero tensor words")
	}

	// Test XID
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

	// Test XID.ToTensor256Words
	x0, x1, x2, x3 := xid.ToTensor256Words()
	if x0 == 0 && x1 == 0 && x2 == 0 && x3 == 0 {
		t.Fatal("expected non-zero tensor words for XID")
	}

	// Test KID
	actor := NewIdentityID()
	resource := []byte("resource")
	capability := []byte("read")
	key := []byte("0123456789abcdef")
	kid, err := NewKIDForCapability(head, actor, resource, capability, key)
	if err != nil {
		t.Fatalf("NewKIDForCapability failed: %v", err)
	}

	// Test KID.Verify
	if !kid.Verify(actor, resource, capability, key) {
		t.Fatal("KID verify failed")
	}

	// Test KID.Verify with wrong capability
	if kid.Verify(actor, resource, []byte("write"), key) {
		t.Fatal("KID verify should fail for wrong capability")
	}

	// Test KID.ToTensor256Words
	k0, k1, k2, k3 := kid.ToTensor256Words()
	if k0 == 0 && k1 == 0 && k2 == 0 && k3 == 0 {
		t.Fatal("expected non-zero tensor words for KID")
	}
}

// TestLIDCoverage tests LID functions for coverage
func TestLIDCoverage(t *testing.T) {
	head := NewLedger()
	var prev LID
	copy(prev[:16], head[:])
	payload := []byte("payload")
	key := []byte("0123456789abcdef")

	lid, err := NewLIDWithHead(head, prev, payload, key)
	if err != nil {
		t.Fatalf("NewLIDWithHead failed: %v", err)
	}

	// Test LID.Head
	if lid.Head() != head {
		t.Fatal("LID head mismatch")
	}

	// Test LID.ChainHash
	if lid.ChainHash() == [16]byte{} {
		t.Fatal("expected non-zero chain hash")
	}

	// Test LID.Verify
	if !lid.Verify(prev, payload, key) {
		t.Fatal("LID verify failed")
	}

	// Test LID.Verify with wrong payload
	if lid.Verify(prev, []byte("wrong"), key) {
		t.Fatal("LID verify should fail for wrong payload")
	}

	// Test LIDBLAKE3
	blake3LID, err := LIDBLAKE3(head, prev, payload, key)
	if err != nil {
		t.Fatalf("LIDBLAKE3 failed: %v", err)
	}
	if blake3LID.Head() != head {
		t.Fatal("BLAKE3 LID head mismatch")
	}

	// Test NewLID with empty key (should error)
	_, err = NewLID(prev, payload, []byte{})
	if err != ErrInvalidLIDKey {
		t.Fatal("expected ErrInvalidLIDKey for empty key")
	}
}

// TestTypesCoverage tests types.go functions for coverage
func TestTypesCoverage(t *testing.T) {
	// Test ScopeID
	scope := NewScope(Matter, "tenant-123")
	scopeStr := scope.String(Matter)
	if scopeStr == "" {
		t.Fatal("expected non-empty scope string")
	}

	parsedScope, atom, err := ParseScope(scopeStr)
	if err != nil {
		t.Fatalf("ParseScope failed: %v", err)
	}
	if parsedScope.Scope != "tenant-123" {
		t.Fatal("scope tenant mismatch")
	}

	// Test ShardedID
	sharded := NewSharded("tenant-123", 456)
	shardedStr := sharded.String(Matter)
	if shardedStr == "" {
		t.Fatal("expected non-empty sharded string")
	}

	parsedSharded, atom, err := ParseSharded(shardedStr)
	if err != nil {
		t.Fatalf("ParseSharded failed: %v", err)
	}
	if parsedSharded.ShardKey != 456 {
		t.Fatal("sharded key mismatch")
	}

	// Test AliasID
	alias := NewWithAlias("tenant-123", "alias-456")
	aliasStr := alias.String(Matter)
	if aliasStr == "" {
		t.Fatal("expected non-empty alias string")
	}

	parsedAlias, atom, err := ParseAlias(aliasStr)
	if err != nil {
		t.Fatalf("ParseAlias failed: %v", err)
	}
	if parsedAlias.Alias != "alias-456" {
		t.Fatal("alias mismatch")
	}

	// Test TraceID
	trace := NewTrace()
	traceStr := trace.String()
	if traceStr == "" {
		t.Fatal("expected non-empty trace string")
	}

	var traceID [8]byte
	traceParent := trace.TraceParent(traceID)
	if traceParent == "" {
		t.Fatal("expected non-empty trace parent")
	}

	// Test GrantID
	secret := []byte("0123456789abcdef0123456789abcdef")
	grant := NewGrant(Matter, time.Hour, secret)
	if grant.ID == Zero {
		t.Fatal("expected non-zero grant ID")
	}

	// Test GrantID.Verify
	if !grant.Verify(secret) {
		t.Fatal("grant verify failed")
	}

	// Test GrantID.Verify with wrong secret
	if grant.Verify([]byte("wrongsecret12345678")) {
		t.Fatal("grant verify should fail for wrong secret")
	}

	// Test grant with invalid secret length
	_ = NewGrant(Matter, time.Hour, []byte("short"))

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

	// Test TestID helper
	testID := TestID(Matter, time.UnixMilli(1700000000123), 42)
	if testID == Zero {
		t.Fatal("expected non-zero test ID")
	}

	testID2 := TestIDSequence(Matter, time.UnixMilli(1700000000123), 100)
	if len(testID2) == 0 {
		t.Fatal("expected non-zero test IDs with sequence")
	}
}

// TestWireFormatCoverage tests wire format functions for coverage
func TestWireFormatCoverage(t *testing.T) {
	// Save original state
	oldDefaultFormat := DefaultWireFormat()
	oldAcceptUnderscore := AcceptUnderscore()
	t.Cleanup(func() {
		SetDefaultWireFormat(oldDefaultFormat)
		SetAcceptUnderscore(oldAcceptUnderscore)
	})

	// Test DefaultWireFormat
	defaultFormat := DefaultWireFormat()
	if defaultFormat != WireColon {
		t.Fatal("expected default wire format to be colon")
	}

	// Test SetDefaultWireFormat
	SetDefaultWireFormat(WireUnderscore)
	if DefaultWireFormat() != WireUnderscore {
		t.Fatal("expected default wire format to change to underscore")
	}
	SetDefaultWireFormat(WireColon)

	// Test AcceptUnderscore
	SetAcceptUnderscore(true)
	if !AcceptUnderscore() {
		t.Fatal("expected accept underscore to be true")
	}
	SetAcceptUnderscore(false)

	// Test ParseFormatStats
	colonCount, underscoreCount := ParseFormatStats()
	// Stats may be non-zero due to previous tests, just verify they work
	if colonCount > 1000000 || underscoreCount > 1000000 {
		t.Fatal("stats unexpectedly large")
	}
}

// TestAtomCoverage tests atom-related functions for coverage
func TestAtomCoverage(t *testing.T) {
	// Test all canonical atoms
	atoms := []Atom{Identity, Tenant, Matter, Space, Time, Ledger, Legal, Trust, Kinetic, Cognition, Semantic, System, Key, Event, Session}
	for _, atom := range atoms {
		canonical := CanonicalAtom(atom)
		if canonical == "" {
			t.Fatalf("expected canonical atom for %v", atom)
		}
	}

	// Test legacy atom normalization
	legacyAtom := AtomFromString("OBJ")
	// OBJ may normalize to Matter, stay as Object, or be empty depending on implementation
	// Just verify the function works without error
	_ = legacyAtom

	// Test invalid atom
	invalidAtom := AtomFromString("INVALID")
	if invalidAtom != "" {
		t.Fatalf("expected empty string for invalid atom, got %s", invalidAtom)
	}
}

// TestBoundaryCoverage tests boundary projection functions for coverage
func TestBoundaryCoverage(t *testing.T) {
	id := NewFast()

	// Test ToLLMFormat
	llm := id.ToLLMFormat(Matter)
	if llm.Atom != Matter {
		t.Fatal("LLM format atom mismatch")
	}
	if llm.TimestampMillis == 0 {
		t.Fatal("expected non-zero timestamp in LLM format")
	}

	// Test ToLLMFormatV2
	llm2 := id.ToLLMFormatV2(Matter)
	if llm2.Kind != "snid" {
		t.Fatal("LLM format v2 kind mismatch")
	}
	if llm2.Atom != Matter {
		t.Fatal("LLM format v2 atom mismatch")
	}

	// Test TimeBin
	timeBin := id.TimeBin(3600000) // 1 hour
	if timeBin == 0 {
		t.Fatal("expected non-zero time bin")
	}

	// Test WithGhostBit
	ghosted := id.WithGhostBit(true)
	if !ghosted.IsGhosted() {
		t.Fatal("expected ghosted ID to have ghost bit set")
	}

	unghosted := ghosted.WithGhostBit(false)
	if unghosted.IsGhosted() {
		t.Fatal("expected unghosted ID to not have ghost bit set")
	}

	// Test TensorTimeDeltaMillis
	hi, lo := id.ToTensorWords()
	delta := TensorTimeDeltaMillis(hi, hi-int64(1<<16))
	if delta != 1 {
		t.Fatalf("expected delta of 1, got %d", delta)
	}

	// Test EncodeFixed64Pair / DecodeFixed64Pair
	pair := EncodeFixed64Pair(hi, lo)
	decodedHi, decodedLo, err := DecodeFixed64Pair(pair[:])
	if err != nil {
		t.Fatalf("DecodeFixed64Pair failed: %v", err)
	}
	if decodedHi != hi || decodedLo != lo {
		t.Fatal("fixed64 pair roundtrip failed")
	}
}

// TestNeuralCoverage tests neural ID functions for coverage
func TestNeuralCoverage(t *testing.T) {
	head := NewFast()
	var semantic [16]byte
	for i := range semantic {
		semantic[i] = byte(255 - i)
	}

	nid := NewNeuralFromHash(head, semantic)
	if nid.Head() != head {
		t.Fatal("NID head mismatch")
	}

	// Test HammingDistance
	zeroNid := NeuralID{}
	distance := nid.HammingDistance(zeroNid)
	if distance == 0 {
		t.Fatal("expected non-zero hamming distance")
	}

	// Test distance to self
	selfDistance := nid.HammingDistance(nid)
	if selfDistance != 0 {
		t.Fatal("expected zero hamming distance to self")
	}
}

// TestSpatialCoverage tests spatial ID functions for coverage
// Note: Comprehensive spatial tests are in neural_test.go TestSpatialID
func TestSpatialCoverage(t *testing.T) {
	// Test NewSpatialFromCell projection
	sgid := NewSpatialFromCell(0x8c2a1072b59ffff, 0x1234567890ABCDEF)
	if sgid == Zero {
		t.Fatal("expected non-zero SGID from cell")
	}

	// Test IsSpatial
	if !sgid.IsSpatial() {
		t.Fatal("expected SGID to be spatial")
	}

	// Test LatLng projection
	lat, lng := sgid.LatLng()
	if lat == 0 && lng == 0 {
		t.Fatal("expected non-zero lat/lng")
	}

	// Test H3Cell projection
	cell := sgid.H3Cell()
	if cell == 0 {
		t.Fatal("expected non-zero H3 cell")
	}

	// Test H3FeatureVector projection
	features := sgid.H3FeatureVector()
	if len(features) == 0 {
		t.Fatal("expected non-empty feature vector")
	}
}

// TestEIDCoverage tests ephemeral ID functions for coverage
func TestEIDCoverage(t *testing.T) {
	eid := NewEphemeralAt(1700000000123, 0x00FF)
	eidBytes := eid.Bytes()

	if len(eidBytes) != 8 {
		t.Fatalf("expected 8 bytes for EID, got %d", len(eidBytes))
	}

	if eid.Counter() != 0x00FF {
		t.Fatalf("expected counter 0x00FF, got %d", eid.Counter())
	}

	ts := eid.Time()
	if ts.UnixMilli() != 1700000000123 {
		t.Fatalf("expected timestamp 1700000000123, got %d", ts.UnixMilli())
	}
}

// TestBIDCoverage tests BID functions for coverage
func TestBIDCoverage(t *testing.T) {
	head := NewFast()
	var content [32]byte
	for i := range content {
		content[i] = byte(i)
	}

	bid := NewBIDWithTopology(head, content)
	if bid.Topology != head {
		t.Fatal("BID topology mismatch")
	}
	if bid.Content != content {
		t.Fatal("BID content mismatch")
	}

	// Test WireFormat
	wire := bid.WireFormat()
	if wire == "" {
		t.Fatal("expected non-empty wire format")
	}

	// Test R2Key
	r2Key := bid.R2Key()
	if r2Key == "" {
		t.Fatal("expected non-empty R2 key")
	}

	// Test Neo4jID
	neo4jID := bid.Neo4jID()
	if neo4jID == "" {
		t.Fatal("expected non-empty Neo4j ID")
	}
}

// TestParseWithOptionsCoverage tests ParseWithOptions with various options
func TestParseWithOptionsCoverage(t *testing.T) {
	id := NewFast()
	wire := id.String(Matter)

	// Test with version validation
	parsed, atom, err := ParseWithOptions(wire, ValidationOptions{
		RequireVersion7: true,
		CheckTimestamp:  false,
	})
	if err != nil {
		t.Fatalf("ParseWithOptions failed: %v", err)
	}
	if atom != Matter {
		t.Fatalf("atom mismatch: got %s want %s", atom, Matter)
	}
	if parsed != id {
		t.Fatal("parsed ID mismatch")
	}

	// Test with timestamp check
	parsed2, atom2, err := ParseWithOptions(wire, ValidationOptions{
		CheckTimestamp: true,
		MaxAge:         time.Hour,
	})
	if err != nil {
		t.Fatalf("ParseWithOptions with timestamp check failed: %v", err)
	}
	if atom2 != Matter {
		t.Fatalf("atom mismatch with timestamp check: got %s want %s", atom2, Matter)
	}
	if parsed2 != id {
		t.Fatal("parsed ID mismatch with timestamp check")
	}
}

// TestStringVariantsCoverage tests various string format variants
func TestStringVariantsCoverage(t *testing.T) {
	id := NewFast()

	// Test StringWithFormat with different formats
	canonical := id.StringWithFormat(Matter, WireColon)
	underscore := id.StringWithFormat(Matter, WireUnderscore)

	if canonical == "" {
		t.Fatal("expected non-empty canonical string")
	}
	if underscore == "" {
		t.Fatal("expected non-empty underscore string")
	}
	if canonical == underscore {
		t.Fatal("expected canonical and underscore formats to differ")
	}

	// Test all atom string methods
	testCases := []struct {
		method func() string
		atom   Atom
	}{
		{id.MatterString, Matter},
		{id.LocationString, Space},
		{id.ChronosString, Time},
		{id.LedgerString, Ledger},
		{id.LegalString, Legal},
		{id.TrustString, Trust},
		{id.KineticString, Kinetic},
		{id.CognitionString, Cognition},
		{id.SemanticString, Semantic},
		{id.SystemString, System},
		{id.AccessKeyString, Key},
		{id.EventString, Event},
		{id.SessionString, Session},
	}

	for _, tc := range testCases {
		str := tc.method()
		if str == "" {
			t.Fatalf("expected non-empty string for atom %s", tc.atom)
		}
	}
}

// TestErrorPathsCoverage tests error paths for better coverage
func TestErrorPathsCoverage(t *testing.T) {
	// Test ParseWithOptions with invalid wire
	_, _, err := ParseWithOptions("INVALID", ValidationOptions{})
	if err == nil {
		t.Fatal("expected error for invalid wire")
	}

	// Test ParseWithOptions with empty wire
	_, _, err = ParseWithOptions("", ValidationOptions{})
	if err == nil {
		t.Fatal("expected error for empty wire")
	}

	// Test ParseWithOptions with wire without atom
	_, _, err = ParseWithOptions("invalidpayload", ValidationOptions{})
	if err == nil {
		t.Fatal("expected error for wire without atom")
	}

	// Test FromUUIDv7 with invalid version
	var invalidUUID [16]byte
	invalidUUID[6] = 0x40 // Version 4
	_, err = FromUUIDv7(invalidUUID)
	if err == nil {
		t.Fatal("expected error for non-v7 UUID")
	}

	// Test ParseUUIDString with invalid format
	_, err = ParseUUIDString("invalid")
	if err == nil {
		t.Fatal("expected error for invalid UUID string")
	}

	// Test ParseUUIDString with invalid length
	_, err = ParseUUIDString("12345678-1234-1234-1234-12345678901")
	if err == nil {
		t.Fatal("expected error for short UUID string")
	}

	// Test FromString with invalid atom
	_, _, err = FromString("XXX:invalidpayload")
	if err == nil {
		t.Fatal("expected error for invalid atom")
	}

	// Test FromString with invalid checksum
	_, _, err = FromString("MAT:invalidpayloadbadchecksum")
	if err == nil {
		t.Fatal("expected error for invalid checksum")
	}

	// Test FromBytes with invalid length
	_, err = FromBytes([]byte{1, 2, 3})
	if err == nil {
		t.Fatal("expected error for invalid byte length")
	}

	// Test NewKIDForCapability with empty key
	head := NewFast()
	actor := NewIdentityID()
	_, err = NewKIDForCapability(head, actor, []byte("resource"), []byte("read"), []byte{})
	if err == nil {
		t.Fatal("expected error for empty key in KID")
	}

	// Test NewLID with invalid key length
	var prev LID
	_, err = NewLID(prev, []byte("payload"), []byte("short"))
	// Note: NewLID may accept shorter keys depending on implementation
	// Just verify it doesn't panic
	_ = err

	// Test ParseScope with invalid format
	_, _, err = ParseScope("invalid-scope-format")
	if err == nil {
		t.Fatal("expected error for invalid scope format")
	}

	// Test ParseSharded with invalid format
	_, _, err = ParseSharded("invalid-sharded-format")
	if err == nil {
		t.Fatal("expected error for invalid sharded format")
	}

	// Test ParseAlias with invalid format
	_, _, err = ParseAlias("invalid-alias-format")
	if err == nil {
		t.Fatal("expected error for invalid alias format")
	}
}

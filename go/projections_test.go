package snid

import (
	"testing"
)

// TestProjections tests projection functions for coverage
func TestProjections(t *testing.T) {
	id := NewFast()

	// Test ToTensorWords
	hi, lo := id.ToTensorWords()
	if hi == 0 && lo == 0 {
		t.Fatal("expected non-zero tensor words")
	}

	// Test TensorTimeDeltaMillis
	delta := TensorTimeDeltaMillis(hi, hi-int64(1<<16))
	if delta != 1 {
		t.Fatalf("expected delta of 1, got %d", delta)
	}

	// Test negative delta
	negDelta := TensorTimeDeltaMillis(hi, hi+int64(1<<16))
	if negDelta != -1 {
		t.Fatalf("expected negative delta of -1, got %d", negDelta)
	}
}

// TestLLMFormat tests LLM format projections for coverage
func TestLLMFormat(t *testing.T) {
	id := NewFast()

	// Test ToLLMFormat with different atoms
	atoms := []Atom{Identity, Tenant, Matter, Space, Time}
	for _, atom := range atoms {
		llm := id.ToLLMFormat(atom)
		if llm.Atom != atom {
			t.Fatalf("atom mismatch in LLM format: got %s want %s", llm.Atom, atom)
		}
		if llm.TimestampMillis == 0 {
			t.Fatal("expected non-zero timestamp in LLM format")
		}
	}

	// Test ToLLMFormatV2
	llm2 := id.ToLLMFormatV2(Matter)
	if llm2.Kind != "snid" {
		t.Fatalf("expected kind 'snid', got %s", llm2.Kind)
	}
	if llm2.Atom != Matter {
		t.Fatalf("atom mismatch in LLM format v2: got %s want %s", llm2.Atom, Matter)
	}
	if llm2.TimestampMillis == 0 {
		t.Fatal("expected non-zero timestamp in LLM format v2")
	}
}

// TestTimeBin tests time binning functions for coverage
func TestTimeBin(t *testing.T) {
	id := NewFast()

	// Test different bin sizes
	bins := []int64{
		60000,     // 1 minute
		3600000,   // 1 hour
		86400000,  // 1 day
		604800000, // 1 week
	}
	for _, binSize := range bins {
		timeBin := id.TimeBin(binSize)
		if timeBin == 0 {
			t.Fatalf("expected non-zero time bin for size %d", binSize)
		}
	}

	// Test zero bin size (should handle gracefully)
	zeroBin := id.TimeBin(0)
	_ = zeroBin // Just verify it doesn't panic
}

// TestFixed64Pair tests fixed64 pair encoding for coverage
func TestFixed64Pair(t *testing.T) {
	hi := int64(0x0123456789ABCDEF)
	lo := int64(0x76543210)

	// Test EncodeFixed64Pair
	pair := EncodeFixed64Pair(hi, lo)
	if len(pair) != 16 {
		t.Fatalf("expected 16-byte pair, got %d", len(pair))
	}

	// Test DecodeFixed64Pair
	decodedHi, decodedLo, err := DecodeFixed64Pair(pair[:])
	if err != nil {
		t.Fatalf("DecodeFixed64Pair failed: %v", err)
	}
	if decodedHi != hi || decodedLo != lo {
		t.Fatal("fixed64 pair roundtrip failed")
	}

	// Test invalid pair length
	_, _, err = DecodeFixed64Pair([]byte{1, 2, 3})
	if err == nil {
		t.Fatal("expected error for invalid pair length")
	}
}

// TestGhostBit tests ghost bit functionality for coverage
func TestGhostBit(t *testing.T) {
	id := NewFast()

	// Test WithGhostBit
	ghosted := id.WithGhostBit(true)
	if !ghosted.IsGhosted() {
		t.Fatal("expected ghosted ID to have ghost bit set")
	}

	unghosted := ghosted.WithGhostBit(false)
	if unghosted.IsGhosted() {
		t.Fatal("expected unghosted ID to not have ghost bit set")
	}

	// Test ghost bit functionality
	// Ghost bit may or may not affect ID equality depending on implementation
	// Just verify the bit can be set and cleared
	if !ghosted.IsGhosted() {
		t.Fatal("expected ghosted ID to have ghost bit set")
	}
	if unghosted.IsGhosted() {
		t.Fatal("expected unghosted ID to not have ghost bit set")
	}
}

// TestNeuralProjections tests neural ID projections for coverage
func TestNeuralProjections(t *testing.T) {
	head := NewFast()
	var semantic [16]byte
	for i := range semantic {
		semantic[i] = byte(255 - i)
	}

	nid := NewNeuralFromHash(head, semantic)

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

// TestSpatialProjections tests spatial ID projections for coverage
// Note: Comprehensive spatial tests are in neural_test.go TestSpatialID
func TestSpatialProjections(t *testing.T) {
	sgid := NewSpatialFromCell(0x8c2a1072b59ffff, 0x1234567890ABCDEF)

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

	// Test feature vector ends at source cell
	if features[len(features)-1] != uint64(cell) {
		t.Fatal("feature vector should end at source cell")
	}
}

// TestCompositeTensor256 tests composite ID tensor256 projections for coverage
func TestCompositeTensor256(t *testing.T) {
	head := NewFast()
	var hash [16]byte
	for i := range hash {
		hash[i] = byte(i)
	}

	// Test WID.ToTensor256Words
	wid := NewWID(head, hash)
	w0, w1, w2, w3 := wid.ToTensor256Words()
	if w0 == 0 && w1 == 0 && w2 == 0 && w3 == 0 {
		t.Fatal("expected non-zero tensor256 words for WID")
	}

	// Test XID.ToTensor256Words
	var edge [16]byte
	for i := range edge {
		edge[i] = byte(16 - i)
	}
	xid := NewXID(head, edge)
	x0, x1, x2, x3 := xid.ToTensor256Words()
	if x0 == 0 && x1 == 0 && x2 == 0 && x3 == 0 {
		t.Fatal("expected non-zero tensor256 words for XID")
	}

	// Test KID.ToTensor256Words
	actor := NewIdentityID()
	kid, err := NewKIDForCapability(head, actor, []byte("resource"), []byte("read"), []byte("key"))
	if err != nil {
		t.Fatalf("NewKIDForCapability failed: %v", err)
	}
	k0, k1, k2, k3 := kid.ToTensor256Words()
	if k0 == 0 && k1 == 0 && k2 == 0 && k3 == 0 {
		t.Fatal("expected non-zero tensor256 words for KID")
	}
}

// TestEIDProjections tests ephemeral ID projections for coverage
func TestEIDProjections(t *testing.T) {
	eid := NewEphemeralAt(1700000000123, 0x00FF)

	// Test Bytes
	bytes := eid.Bytes()
	if len(bytes) != 8 {
		t.Fatalf("expected 8 bytes for EID, got %d", len(bytes))
	}

	// Test Counter
	counter := eid.Counter()
	if counter != 0x00FF {
		t.Fatalf("expected counter 0x00FF, got %d", counter)
	}

	// Test Time
	ts := eid.Time()
	if ts.UnixMilli() != 1700000000123 {
		t.Fatalf("expected timestamp 1700000000123, got %d", ts.UnixMilli())
	}
}

// TestBIDProjections tests BID projections for coverage
func TestBIDProjections(t *testing.T) {
	head := NewFast()
	var content [32]byte
	for i := range content {
		content[i] = byte(i)
	}

	bid := NewBIDWithTopology(head, content)

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

	// Test Topology accessor
	if bid.Topology != head {
		t.Fatal("topology mismatch")
	}

	// Test Content accessor
	if bid.Content != content {
		t.Fatal("content mismatch")
	}
}

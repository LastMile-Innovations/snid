package snid

import (
	"testing"
)

// TestNeuralID tests neural ID functions for coverage
func TestNeuralID(t *testing.T) {
	head := NewFast()
	var semantic [16]byte
	for i := range semantic {
		semantic[i] = byte(i)
	}

	// Test NewNeuralFromHash
	nid := NewNeuralFromHash(head, semantic)
	if nid.Head() != head {
		t.Fatal("neural ID head mismatch")
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

	// Test Head accessor
	if nid.Head() != head {
		t.Fatal("head accessor mismatch")
	}
}

// TestSpatialID tests spatial ID functions comprehensively
func TestSpatialID(t *testing.T) {
	testCases := []struct {
		name      string
		lat       float64
		lng       float64
		precision int
	}{
		{"San Francisco", 37.7749, -122.4194, 12},
		{"New York", 40.7128, -74.0060, 10},
		{"London", 51.5074, -0.1278, 11},
		{"Tokyo", 35.6762, 139.6503, 12},
	}

	for _, tc := range testCases {
		t.Run(tc.name+" NewSpatial", func(t *testing.T) {
			spatial := NewSpatial(tc.lat, tc.lng)
			if spatial == Zero {
				t.Fatal("expected non-zero spatial ID")
			}
			if !spatial.IsSpatial() {
				t.Fatal("expected spatial ID to be spatial")
			}
		})

		t.Run(tc.name+" NewSpatialPrecise", func(t *testing.T) {
			spatialPrecise := NewSpatialPrecise(tc.lat, tc.lng, tc.precision)
			if spatialPrecise == Zero {
				t.Fatal("expected non-zero spatial ID with precision")
			}
			if !spatialPrecise.IsSpatial() {
				t.Fatal("expected precise spatial ID to be spatial")
			}
		})
	}

	// Test NewSpatialFromCell
	t.Run("FromCell", func(t *testing.T) {
		sgid := NewSpatialFromCell(0x8c2a1072b59ffff, 0x1234567890ABCDEF)
		if sgid == Zero {
			t.Fatal("expected non-zero SGID from cell")
		}
		if !sgid.IsSpatial() {
			t.Fatal("expected SGID to be spatial")
		}

		// Test regular ID is not spatial
		regular := NewFast()
		if regular.IsSpatial() {
			t.Fatal("expected regular ID to not be spatial")
		}

		// Test LatLng
		lat, lng := sgid.LatLng()
		if lat == 0 && lng == 0 {
			t.Fatal("expected non-zero lat/lng")
		}

		// Test H3Cell
		cell := sgid.H3Cell()
		if cell == 0 {
			t.Fatal("expected non-zero H3 cell")
		}

		// Test H3FeatureVector
		features := sgid.H3FeatureVector()
		if len(features) == 0 {
			t.Fatal("expected non-empty feature vector")
		}

		// Test feature vector ends at source cell
		if features[len(features)-1] != uint64(cell) {
			t.Fatal("feature vector should end at source cell")
		}
	})
}

// TestEID tests ephemeral ID functions for coverage
func TestEID(t *testing.T) {
	// Test NewEphemeralAt
	eid := NewEphemeralAt(1700000000123, 0x00FF)
	eidBytes := eid.Bytes()

	if len(eidBytes) != 8 {
		t.Fatalf("expected 8 bytes for EID, got %d", len(eidBytes))
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

// TestBID tests BID functions for coverage
func TestBID(t *testing.T) {
	head := NewFast()
	var content [32]byte
	for i := range content {
		content[i] = byte(i)
	}

	bid := NewBIDWithTopology(head, content)

	// Test Topology
	if bid.Topology != head {
		t.Fatal("BID topology mismatch")
	}

	// Test Content
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

// TestCompositeID tests composite ID functions for coverage
func TestCompositeID(t *testing.T) {
	head := NewFast()
	var hash [16]byte
	for i := range hash {
		hash[i] = byte(i)
	}

	// Test WID
	wid := NewWID(head, hash)
	if wid.Head() != head {
		t.Fatal("WID head mismatch")
	}
	if wid.ScenarioHash() != hash {
		t.Fatal("WID scenario hash mismatch")
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
}

// TestIdentityID tests identity ID functions for coverage
func TestIdentityID(t *testing.T) {
	// Test NewIdentityID
	id := NewIdentityID()
	if id == Zero {
		t.Fatal("expected non-zero identity ID")
	}

	// Test TenantString
	tenantStr := id.TenantString()
	if tenantStr == "" {
		t.Fatal("expected non-empty tenant string")
	}

	// Test AccessKeyString
	keyStr := id.AccessKeyString()
	if keyStr == "" {
		t.Fatal("expected non-empty access key string")
	}
}

// TestLedgerID tests ledger ID functions for coverage
func TestLedgerID(t *testing.T) {
	// Test NewLedger
	ledger := NewLedger()
	if ledger == Zero {
		t.Fatal("expected non-zero ledger ID")
	}

	// Test LedgerString
	ledgerStr := ledger.LedgerString()
	if ledgerStr == "" {
		t.Fatal("expected non-empty ledger string")
	}
}

// TestAssetID tests asset ID functions for coverage
func TestAssetID(t *testing.T) {
	// Test NewAsset
	catalogID := NewCatalog("category", "brand", "specs")
	asset := NewAsset(catalogID, "tenant-123", "serial-456")
	if asset == Zero {
		t.Fatal("expected non-zero asset ID")
	}
}

// TestCatalogID tests catalog ID functions for coverage
func TestCatalogID(t *testing.T) {
	// Test NewCatalog
	catalog := NewCatalog("category", "brand", "specs")
	if catalog == Zero {
		t.Fatal("expected non-zero catalog ID")
	}
}

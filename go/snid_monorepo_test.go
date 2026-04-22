package snid

import (
	"encoding/hex"
	"testing"
	"time"
)

func TestCoreWireRoundTrip(t *testing.T) {
	id := FromParts(1700000000123, 42, 0x123456, 0x0ABCDEFFF)
	wire := id.String(Matter)

	parsed, atom, err := FromString(wire)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if atom != Matter {
		t.Fatalf("atom mismatch: got %s want %s", atom, Matter)
	}
	if parsed != id {
		t.Fatalf("bytes mismatch: got %x want %x", parsed, id)
	}
}

func TestBIDWireRoundTrip(t *testing.T) {
	head := FromParts(1700000000456, 7, 0x234567, 0x012345678)
	var hash [32]byte
	for i := range hash {
		hash[i] = byte(i)
	}
	bid := NewBIDWithTopology(head, hash)
	wire := bid.WireFormat()

	parsed, err := ParseBIDWire(wire)
	if err != nil {
		t.Fatalf("ParseBIDWire failed: %v", err)
	}
	if parsed.Topology != bid.Topology {
		t.Fatalf("topology mismatch: got %x want %x", parsed.Topology, bid.Topology)
	}
	if parsed.Content != bid.Content {
		t.Fatalf("content mismatch: got %s want %s", hex.EncodeToString(parsed.Content[:]), hex.EncodeToString(bid.Content[:]))
	}
}

func TestDeterministicExtendedConstructors(t *testing.T) {
	head := FromParts(1700000000999, 9, 0x654321, 0x123456789)
	sgid := NewSpatialFromCell(0x8c2a1072b59ffff, 0x1234567890ABCDEF)
	if !sgid.IsSpatial() {
		t.Fatalf("expected spatial id")
	}

	var semantic [16]byte
	for i := range semantic {
		semantic[i] = byte(255 - i)
	}
	nid := NewNeuralFromHash(head, semantic)
	if nid.Head() != head {
		t.Fatalf("nid head mismatch")
	}

	var prev LID
	copy(prev[:16], head[:])
	lid, err := NewLIDWithHead(head, prev, []byte("payload"), []byte("0123456789abcdef"))
	if err != nil {
		t.Fatalf("NewLIDWithHead failed: %v", err)
	}
	if lid.Head() != head {
		t.Fatalf("lid head mismatch")
	}

	eid := NewEphemeralAt(1700000000123, 0x00FF)
	if eid.Counter() != 0x00FF {
		t.Fatalf("eid counter mismatch: got %d", eid.Counter())
	}
}

func TestTensorAndLLMBoundaryHelpers(t *testing.T) {
	id := NewDeterministicIngestID(1700000000123, []byte("tensor-boundary"))
	hi, lo := id.ToTensorWords()
	if got := hi >> 16; got != 1700000000123 {
		t.Fatalf("unexpected timestamp in high word: got %d", got)
	}
	if lo == 0 {
		t.Fatalf("unexpected zero low word")
	}
	if got := TensorTimeDeltaMillis(hi, hi-int64(1<<16)); got != 1 {
		t.Fatalf("unexpected tensor delta: got %d want 1", got)
	}
	llm := id.ToLLMFormat(Matter)
	if llm.Atom != Matter || llm.TimestampMillis != 1700000000123 || llm.MachineOrShard != id.MachineOrShard() || llm.Sequence != id.Sequence() {
		t.Fatalf("unexpected llm format: %+v", llm)
	}
	llm2 := id.ToLLMFormatV2(Matter)
	if llm2.Kind != "snid" || llm2.TimestampMillis != 1700000000123 || llm2.Ghosted {
		t.Fatalf("unexpected llm format v2: %+v", llm2)
	}
	if got := id.TimeBin(3600000); got != 1699999200000 {
		t.Fatalf("unexpected time bin: got %d", got)
	}
	ghosted := id.WithGhostBit(true)
	if !ghosted.IsGhosted() {
		t.Fatalf("expected ghost bit to be set")
	}
	pair := EncodeFixed64Pair(hi, lo)
	decodedHi, decodedLo, err := DecodeFixed64Pair(pair[:])
	if err != nil {
		t.Fatalf("decode fixed64 pair failed: %v", err)
	}
	if decodedHi != hi || decodedLo != lo {
		t.Fatalf("fixed64 pair mismatch")
	}
}

func TestDeterministicIngestID(t *testing.T) {
	id1 := NewDeterministicIngestID(uint64(time.UnixMilli(1700000000123).UnixMilli()), []byte("same-content"))
	id2 := NewDeterministicIngestID(uint64(time.UnixMilli(1700000000123).UnixMilli()), []byte("same-content"))
	id3 := NewDeterministicIngestID(uint64(time.UnixMilli(1700000000123).UnixMilli()), []byte("other-content"))
	if id1 != id2 {
		t.Fatalf("expected deterministic ids to match")
	}
	if id1 == id3 {
		t.Fatalf("expected different content to produce different ids")
	}
	if got := id1.Time().UnixMilli(); got != 1700000000123 {
		t.Fatalf("unexpected timestamp: got %d", got)
	}
}

func TestCompositeTargetTypes(t *testing.T) {
	head := FromParts(1700000000123, 9, 0x123456, 0x123456789)
	actor := FromParts(1700000000456, 11, 0x654321, 0x987654321)

	var scenario [16]byte
	var edge [16]byte
	for i := range scenario {
		scenario[i] = byte(i)
		edge[i] = byte(16 - i)
	}

	wid := NewWID(head, scenario)
	if wid.Head() != head || wid.ScenarioHash() != scenario {
		t.Fatalf("wid projection mismatch")
	}
	if _, _, _, last := wid.ToTensor256Words(); last == 0 {
		t.Fatalf("wid tensor words missing tail")
	}

	xid := NewXID(head, edge)
	if xid.Head() != head || xid.EdgeHash() != edge {
		t.Fatalf("xid projection mismatch")
	}

	kid, err := NewKIDForCapability(head, actor, []byte("resource"), []byte("read"), []byte("0123456789abcdef"))
	if err != nil {
		t.Fatalf("NewKIDForCapability failed: %v", err)
	}
	if !kid.Verify(actor, []byte("resource"), []byte("read"), []byte("0123456789abcdef")) {
		t.Fatalf("kid verify failed")
	}
	if kid.Verify(actor, []byte("resource"), []byte("write"), []byte("0123456789abcdef")) {
		t.Fatalf("kid verify should fail for mismatched capability")
	}
}

func TestSpatialFeatureVector(t *testing.T) {
	sgid := NewSpatialFromCell(0x8c2a1072b59ffff, 0x1234567890ABCDEF)
	features := sgid.H3FeatureVector()
	if len(features) != 13 {
		t.Fatalf("unexpected feature vector length: got %d", len(features))
	}
	if features[len(features)-1] != uint64(sgid.H3Cell()) {
		t.Fatalf("feature vector should end at source cell")
	}
}

func TestUUIDv7Compatibility(t *testing.T) {
	// Test NewUUIDv7 generates valid UUIDv7-compatible IDs
	id := NewUUIDv7()
	uuidStr := id.UUIDString()

	// Verify UUID string format (8-4-4-4-12)
	if len(uuidStr) != 36 {
		t.Fatalf("unexpected UUID string length: got %d", len(uuidStr))
	}

	// Parse back from UUID string
	parsed, err := ParseUUIDString(uuidStr)
	if err != nil {
		t.Fatalf("ParseUUIDString failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("parsed UUID doesn't match original: got %x want %x", parsed, id)
	}

	// Verify version nibble is 0b0111 (version 7)
	version := (id[6] >> 4) & 0x0F
	if version != 7 {
		t.Fatalf("expected UUID version 7, got %d", version)
	}

	// Verify variant bits are 0b10
	variant := (id[8] >> 6) & 0b11
	if variant != 0b10 {
		t.Fatalf("expected UUID variant 0b10, got %b", variant)
	}

	// Test FromUUIDv7 rejects non-v7 UUIDs
	var nonV7 UUID
	copy(nonV7[:], id[:])
	nonV7[6] = (nonV7[6] & 0x0F) | (4 << 4) // Set version to 4
	_, err = FromUUIDv7(nonV7)
	if err == nil {
		t.Fatalf("expected error for non-v7 UUID")
	}

	var badVariant UUID
	copy(badVariant[:], id[:])
	badVariant[8] &^= 0xC0
	_, err = FromUUIDv7(badVariant)
	if err == nil {
		t.Fatalf("expected error for UUID with invalid variant")
	}

	// Test ToUUIDv7 conversion
	uuidObj := id.ToUUIDv7()
	if uuidObj.String() != uuidStr {
		t.Fatalf("ToUUIDv7 string mismatch")
	}
}

func TestUUIDv7Monotonicity(t *testing.T) {
	// Test that IDs generated in quick succession are monotonic
	ids := make([]ID, 100)
	for i := 0; i < 100; i++ {
		ids[i] = NewUUIDv7()
	}

	// Verify each ID is greater than the previous
	for i := 1; i < len(ids); i++ {
		prev := ids[i-1]
		curr := ids[i]
		// Compare as byte arrays
		for j := 0; j < 16; j++ {
			if curr[j] > prev[j] {
				break // curr is greater
			}
			if curr[j] < prev[j] {
				t.Fatalf("IDs not monotonic: ids[%d] < ids[%d-1] at byte %d", i, i, j)
			}
		}
	}

	// Test monotonicity within same millisecond
	// Generate IDs rapidly to force same millisecond
	fastIDs := make([]ID, 50)
	for i := 0; i < 50; i++ {
		fastIDs[i] = NewUUIDv7()
	}

	// Verify monotonic even when timestamps may be same
	for i := 1; i < len(fastIDs); i++ {
		for j := 0; j < 16; j++ {
			if fastIDs[i][j] > fastIDs[i-1][j] {
				break
			}
			if fastIDs[i][j] < fastIDs[i-1][j] {
				t.Fatalf("Fast IDs not monotonic: fastIDs[%d] < fastIDs[%d-1] at byte %d", i, i, j)
			}
		}
	}
}

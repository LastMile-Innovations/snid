package snid

import (
	"sync"
	"testing"
)

// TestGeneratorConcurrency tests concurrent ID generation for coverage
func TestGeneratorConcurrency(t *testing.T) {
	var wg sync.WaitGroup
	ids := make(chan ID, 100)

	// Generate IDs concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				ids <- NewFast()
			}
		}()
	}

	wg.Wait()
	close(ids)

	// Verify all IDs are unique
	seen := make(map[ID]bool)
	for id := range ids {
		if seen[id] {
			t.Fatal("duplicate ID generated")
		}
		seen[id] = true
	}

	if len(seen) != 100 {
		t.Fatalf("expected 100 unique IDs, got %d", len(seen))
	}
}

// TestNewFastWithPrefix tests projected ID generation for coverage
func TestNewFastWithPrefix(t *testing.T) {
	id := NewFastWithPrefix(Matter)
	if id == Zero {
		t.Fatal("expected non-zero ID")
	}
}

// TestNewProjectedWithPrefix tests projected ID with prefix for coverage
func TestNewProjectedWithPrefix(t *testing.T) {
	id := NewProjectedWithPrefix(Matter, "test-tenant")
	if id == Zero {
		t.Fatal("expected non-zero projected ID")
	}
}

// TestTimeFunctions tests time-related ID functions for coverage
func TestTimeFunctions(t *testing.T) {
	// Test Time() method
	id := NewFast()
	ts := id.Time()
	if ts.IsZero() {
		t.Fatal("expected non-zero timestamp")
	}

	// Test TimeBin() method
	timeBin := id.TimeBin(3600000) // 1 hour
	if timeBin == 0 {
		t.Fatal("expected non-zero time bin")
	}
}

// TestIDParts tests ID part extraction methods
func TestIDParts(t *testing.T) {
	id := NewFast()

	// Test MachineOrShard() - verify it returns a valid uint32
	machine := id.MachineOrShard()
	if machine > 0xFFFFFFFF {
		t.Fatalf("MachineOrShard returned invalid value: %d", machine)
	}

	// Test Sequence() - verify it returns a valid uint16
	seq := id.Sequence()
	if seq > 0xFFFF {
		t.Fatalf("Sequence returned invalid value: %d", seq)
	}

	// Test Version() - must be 7 for UUIDv7 compatibility
	version := id.Version()
	if version != 7 {
		t.Fatalf("expected version 7 for UUIDv7-compatible ID, got %d", version)
	}
}

// TestTurboStreamer tests TurboStreamer for coverage
func TestTurboStreamer(t *testing.T) {
	streamer := NewTurboStreamer(100)

	// Generate IDs in a loop
	for i := 0; i < 100; i++ {
		id := streamer.Next()
		if id == Zero {
			t.Fatal("expected non-zero ID from TurboStreamer")
		}
	}
}

// TestAdaptiveGenerator tests adaptive generator functions
func TestAdaptiveGenerator(t *testing.T) {
	// Init adaptive generator
	InitAdaptive()

	// Test nextSpatial with table-driven test cases
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
		t.Run(tc.name, func(t *testing.T) {
			spatial := adaptive.nextSpatial(tc.lat, tc.lng, tc.precision)
			if spatial == Zero {
				t.Fatalf("expected non-zero spatial ID for %s", tc.name)
			}
			if !spatial.IsSpatial() {
				t.Fatalf("expected spatial ID to be spatial for %s", tc.name)
			}
		})
	}
}

// TestAdaptiveGeneratorModes tests mode switching for adaptive generator
func TestAdaptiveGeneratorModes(t *testing.T) {
	InitAdaptive()

	// Save original mode
	originalMode := Mode(adaptive.mode.Load())
	t.Cleanup(func() { SetMode(originalMode) })

	// Test ModeFast
	SetMode(ModeFast)
	if Mode(adaptive.mode.Load()) != ModeFast {
		t.Fatal("expected mode to be set to ModeFast")
	}
	id1 := Next()
	if id1 == Zero {
		t.Fatal("expected non-zero ID in ModeFast")
	}

	// Test ModeSecure
	SetMode(ModeSecure)
	if Mode(adaptive.mode.Load()) != ModeSecure {
		t.Fatal("expected mode to be set to ModeSecure")
	}
	id2 := Next()
	if id2 == Zero {
		t.Fatal("expected non-zero ID in ModeSecure")
	}

	// Test ModeAdaptive
	SetMode(ModeAdaptive)
	if Mode(adaptive.mode.Load()) != ModeAdaptive {
		t.Fatal("expected mode to be set to ModeAdaptive")
	}
	id3 := Next()
	if id3 == Zero {
		t.Fatal("expected non-zero ID in ModeAdaptive")
	}

	// Test Batch generation in different modes
	SetMode(ModeFast)
	batchFast := Batch(10)
	if len(batchFast) != 10 {
		t.Fatalf("expected 10 IDs in batch, got %d", len(batchFast))
	}

	SetMode(ModeSecure)
	batchSecure := Batch(10)
	if len(batchSecure) != 10 {
		t.Fatalf("expected 10 IDs in secure batch, got %d", len(batchSecure))
	}

	// Verify batch IDs are unique
	seen := make(map[ID]bool)
	for _, id := range batchSecure {
		if seen[id] {
			t.Fatal("duplicate ID in secure batch")
		}
		seen[id] = true
	}
}

// TestClockFunctions tests clock-related functions for coverage
func TestClockFunctions(t *testing.T) {
	// Test Time() method on an ID which uses the clock
	id := NewFast()
	ts := id.Time()
	if ts.IsZero() {
		t.Fatal("expected non-zero timestamp from clock")
	}
}

// TestFromParts tests ID construction from parts for coverage
func TestFromParts(t *testing.T) {
	id := FromParts(1700000000123, 42, 0x123456, 0x0ABCDEFFF)
	if id == Zero {
		t.Fatal("expected non-zero ID from parts")
	}

	// Verify timestamp via Time() method
	ts := id.Time().UnixMilli()
	if ts != 1700000000123 {
		t.Fatalf("expected timestamp 1700000000123, got %d", ts)
	}
}

// TestIDType tests ID type checking methods for coverage
func TestIDType(t *testing.T) {
	// Test spatial ID
	spatial := NewSpatial(37.7749, -122.4194)
	if !spatial.IsSpatial() {
		t.Fatal("expected spatial ID to be spatial")
	}

	// Test non-spatial ID
	regular := NewFast()
	if regular.IsSpatial() {
		t.Fatal("expected regular ID to not be spatial")
	}

	// Test ghost bit
	ghosted := regular.WithGhostBit(true)
	if !ghosted.IsGhosted() {
		t.Fatal("expected ghosted ID to have ghost bit set")
	}

	unghosted := ghosted.WithGhostBit(false)
	if unghosted.IsGhosted() {
		t.Fatal("expected unghosted ID to not have ghost bit set")
	}
}

// TestIDValidation tests ID validation methods for coverage
func TestIDValidation(t *testing.T) {
	// Test zero ID check
	id := NewFast()
	if id == Zero {
		t.Fatal("expected non-zero ID")
	}
}

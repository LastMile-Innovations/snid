package neo4j

import (
	"testing"
	"time"

	snid "github.com/LastMile-Innovations/snid"
)

func TestMarshalUnmarshalBinaryProperty(t *testing.T) {
	id := snid.TestID(snid.Matter, time.UnixMilli(1700000000123), 42)
	raw := MarshalProperty(id)
	if len(raw) != 16 {
		t.Fatalf("raw length = %d, want 16", len(raw))
	}

	roundTrip, err := UnmarshalProperty(raw)
	if err != nil {
		t.Fatalf("UnmarshalProperty failed: %v", err)
	}
	if roundTrip != id {
		t.Fatalf("round trip mismatch: got %x want %x", roundTrip, id)
	}
}

func TestBindIDCreatesBinaryParam(t *testing.T) {
	id := snid.TestID(snid.Event, time.UnixMilli(1700000000123), 7)
	params := BindID(nil, "id", id)
	value, ok := params["id"].([]byte)
	if !ok {
		t.Fatalf("expected []byte param, got %T", params["id"])
	}
	if len(value) != 16 {
		t.Fatalf("param length = %d, want 16", len(value))
	}
}

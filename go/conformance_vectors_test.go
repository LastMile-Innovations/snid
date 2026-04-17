package snid

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type vectorFile struct {
	Core          []coreVector        `json:"core"`
	Compatibility compatibilityVector `json:"compatibility"`
	Negative      negativeVector      `json:"negative"`
}

type coreVector struct {
	Atom            string          `json:"atom"`
	BytesHex        string          `json:"bytes_hex"`
	Wire            string          `json:"wire"`
	UnderscoreWire  string          `json:"underscore_wire"`
	TimestampMillis int64           `json:"timestamp_millis"`
	TensorHi        int64           `json:"tensor_hi"`
	TensorLo        int64           `json:"tensor_lo"`
	LLMFormat       llmFormatVector `json:"llm_format"`
}

type llmFormatVector struct {
	Atom            string `json:"atom"`
	TimestampMillis int64  `json:"timestamp_millis"`
	MachineOrShard  uint32 `json:"machine_or_shard"`
	Sequence        uint16 `json:"sequence"`
}

type compatibilityVector struct {
	BytesHex string `json:"bytes_hex"`
	Wire     string `json:"wire"`
}

type negativeVector struct {
	InvalidAtomWire     string `json:"invalid_atom_wire"`
	InvalidBinaryHex    string `json:"invalid_binary_hex"`
	InvalidWireChecksum string `json:"invalid_wire_checksum"`
	InvalidAdapterHex   string `json:"invalid_adapter_hex"`
}

func TestConformanceVectors(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "conformance", "vectors.json"))
	if err != nil {
		t.Fatalf("read vectors: %v", err)
	}
	var vectors vectorFile
	if err := json.Unmarshal(raw, &vectors); err != nil {
		t.Fatalf("parse vectors: %v", err)
	}

	for _, c := range vectors.Core {
		rawBytes, err := hex.DecodeString(c.BytesHex)
		if err != nil {
			t.Fatalf("DecodeString(%s): %v", c.BytesHex, err)
		}
		id, err := FromBytes(rawBytes)
		if err != nil {
			t.Fatalf("FromBytes(%s): %v", c.BytesHex, err)
		}
		if got := id.String(Atom(c.Atom)); got != c.Wire {
			t.Fatalf("wire mismatch: got %s want %s", got, c.Wire)
		}
		if got := id.StringWithFormat(Atom(c.Atom), WireUnderscore); got != c.UnderscoreWire {
			t.Fatalf("underscore mismatch: got %s want %s", got, c.UnderscoreWire)
		}
		parsed, atom, err := FromString(c.Wire)
		if err != nil {
			t.Fatalf("parse wire %s: %v", c.Wire, err)
		}
		if parsed != id || string(atom) != c.Atom {
			t.Fatalf("parsed mismatch for %s", c.Wire)
		}
		hi, lo := id.ToTensorWords()
		if hi != c.TensorHi || lo != c.TensorLo {
			t.Fatalf("tensor mismatch for %s", c.Wire)
		}
		llm := id.ToLLMFormat(Atom(c.Atom))
		if string(llm.Atom) != c.LLMFormat.Atom || llm.TimestampMillis != c.LLMFormat.TimestampMillis || llm.MachineOrShard != c.LLMFormat.MachineOrShard || llm.Sequence != c.LLMFormat.Sequence {
			t.Fatalf("llm mismatch for %s", c.Wire)
		}
	}

	compatRaw, err := hex.DecodeString(vectors.Compatibility.BytesHex)
	if err != nil {
		t.Fatalf("compat DecodeString: %v", err)
	}
	compatID, err := FromBytes(compatRaw)
	if err != nil {
		t.Fatalf("compat FromBytes: %v", err)
	}
	if got := compatID.String(Matter); got != vectors.Compatibility.Wire {
		t.Fatalf("compat wire mismatch: got %s want %s", got, vectors.Compatibility.Wire)
	}

	if _, _, err := FromString(vectors.Negative.InvalidAtomWire); err == nil {
		t.Fatalf("expected invalid atom wire to fail")
	}
	if raw, err := hex.DecodeString(vectors.Negative.InvalidBinaryHex); err == nil {
		if _, err := FromBytes(raw); err == nil {
			t.Fatalf("expected invalid binary hex to fail")
		}
	} else {
		t.Fatalf("expected invalid binary hex to decode for length test: %v", err)
	}
	if _, _, err := FromString(vectors.Negative.InvalidWireChecksum); err == nil {
		t.Fatalf("expected invalid checksum wire to fail")
	}
	if raw, err := hex.DecodeString(vectors.Negative.InvalidAdapterHex); err == nil {
		if _, err := FromBytes(raw); err == nil {
			t.Fatalf("expected invalid adapter hex to fail")
		}
	} else {
		t.Fatalf("expected invalid adapter hex to decode for length test: %v", err)
	}
}

package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"os"
	"time"

	snid "github.com/neighbor/snid"
)

type vectorFile struct {
	Version       string            `json:"version"`
	GeneratedAt   string            `json:"generated_at"`
	Core          []coreVector      `json:"core"`
	Spatial       spatialCase       `json:"spatial"`
	Neural        neuralCase        `json:"neural"`
	Ledger        ledgerCase        `json:"ledger"`
	World         worldCase         `json:"world"`
	Edge          edgeCase          `json:"edge"`
	Capability    capabilityCase    `json:"capability"`
	Ephemeral     eidCase           `json:"ephemeral"`
	BID           bidCase           `json:"bid"`
	Compatibility compatibilityCase `json:"compatibility"`
	UUIDv7        uuidv7Case        `json:"uuidv7"`
	Negative      negativeCase      `json:"negative"`
}

type coreVector struct {
	Name            string        `json:"name"`
	Atom            string        `json:"atom"`
	BytesHex        string        `json:"bytes_hex"`
	Wire            string        `json:"wire"`
	UnderscoreWire  string        `json:"underscore_wire"`
	TimestampMillis int64         `json:"timestamp_millis"`
	TensorHi        int64         `json:"tensor_hi"`
	TensorLo        int64         `json:"tensor_lo"`
	LLMFormat       llmFormatCase `json:"llm_format"`
	LLMFormatV2     llmFormatV2   `json:"llm_format_v2"`
	TimeBinHour     int64         `json:"time_bin_hour"`
	GhostedBytesHex string        `json:"ghosted_bytes_hex"`
}

type llmFormatCase struct {
	Atom            string `json:"atom"`
	TimestampMillis int64  `json:"timestamp_millis"`
	MachineOrShard  uint32 `json:"machine_or_shard"`
	Sequence        uint16 `json:"sequence"`
}

type llmFormatV2 struct {
	Kind            string `json:"kind"`
	Atom            string `json:"atom"`
	TimestampMillis int64  `json:"timestamp_millis"`
	SpatialAnchor   uint64 `json:"spatial_anchor,omitempty"`
	MachineOrShard  uint32 `json:"machine_or_shard"`
	Sequence        uint16 `json:"sequence"`
	Ghosted         bool   `json:"ghosted"`
}

type spatialCase struct {
	Atom       string  `json:"atom"`
	BytesHex   string  `json:"bytes_hex"`
	Wire       string  `json:"wire"`
	H3CellHex  string  `json:"h3_cell_hex"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Resolution int     `json:"resolution"`
}

type neuralCase struct {
	HeadHex       string `json:"head_hex"`
	BytesHex      string `json:"bytes_hex"`
	SemanticHex   string `json:"semantic_hex"`
	HammingToZero int    `json:"hamming_to_zero"`
}

type ledgerCase struct {
	HeadHex    string `json:"head_hex"`
	BytesHex   string `json:"bytes_hex"`
	PrevHex    string `json:"prev_hex"`
	PayloadHex string `json:"payload_hex"`
	KeyHex     string `json:"key_hex"`
	Blake3Hex  string `json:"blake3_hex"`
}

type worldCase struct {
	HeadHex     string   `json:"head_hex"`
	BytesHex    string   `json:"bytes_hex"`
	ScenarioHex string   `json:"scenario_hex"`
	TensorWords [4]int64 `json:"tensor_words"`
}

type edgeCase struct {
	HeadHex     string   `json:"head_hex"`
	BytesHex    string   `json:"bytes_hex"`
	EdgeHex     string   `json:"edge_hex"`
	TensorWords [4]int64 `json:"tensor_words"`
}

type capabilityCase struct {
	HeadHex       string   `json:"head_hex"`
	ActorHex      string   `json:"actor_hex"`
	ResourceHex   string   `json:"resource_hex"`
	CapabilityHex string   `json:"capability_hex"`
	KeyHex        string   `json:"key_hex"`
	BytesHex      string   `json:"bytes_hex"`
	TensorWords   [4]int64 `json:"tensor_words"`
}

type eidCase struct {
	BytesHex        string `json:"bytes_hex"`
	TimestampMillis uint64 `json:"timestamp_millis"`
	Counter         uint16 `json:"counter"`
}

type bidCase struct {
	TopologyHex string `json:"topology_hex"`
	ContentHex  string `json:"content_hex"`
	Wire        string `json:"wire"`
	R2Key       string `json:"r2_key"`
	Neo4jID     string `json:"neo4j_id"`
}

type compatibilityCase struct {
	BytesHex string `json:"bytes_hex"`
	Wire     string `json:"wire"`
}

type uuidv7Case struct {
	BytesHex        string `json:"bytes_hex"`
	UUIDString      string `json:"uuid_string"`
	TimestampMillis int64  `json:"timestamp_millis"`
	Version         int    `json:"version"`
	Variant         int    `json:"variant"`
}

type negativeCase struct {
	InvalidAtomWire     string `json:"invalid_atom_wire"`
	InvalidBinaryHex    string `json:"invalid_binary_hex"`
	InvalidWireChecksum string `json:"invalid_wire_checksum"`
	InvalidAdapterHex   string `json:"invalid_adapter_hex"`
}

func main() {
	outPath := flag.String("out", "../conformance/vectors.json", "output path")
	flag.Parse()

	ts := time.UnixMilli(1700000000123).UTC()
	coreIDs := []struct {
		name string
		atom snid.Atom
		seq  uint32
	}{
		{name: "matter", atom: snid.Matter, seq: 1},
		{name: "event", atom: snid.Event, seq: 2},
		{name: "tenant", atom: snid.Tenant, seq: 3},
	}

	file := vectorFile{
		Version:     "0.2.0",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Core:        make([]coreVector, 0, len(coreIDs)),
	}

	for _, c := range coreIDs {
		id := snid.TestID(c.atom, ts, c.seq)
		hi, lo := id.ToTensorWords()
		llm := id.ToLLMFormat(c.atom)
		llm2 := id.ToLLMFormatV2(c.atom)
		ghosted := id.WithGhostBit(true)
		file.Core = append(file.Core, coreVector{
			Name:            c.name,
			Atom:            string(c.atom),
			BytesHex:        hex.EncodeToString(id[:]),
			Wire:            id.String(c.atom),
			UnderscoreWire:  id.StringWithFormat(c.atom, snid.WireUnderscore),
			TimestampMillis: id.Time().UnixMilli(),
			TensorHi:        hi,
			TensorLo:        lo,
			LLMFormat: llmFormatCase{
				Atom:            string(llm.Atom),
				TimestampMillis: llm.TimestampMillis,
				MachineOrShard:  llm.MachineOrShard,
				Sequence:        llm.Sequence,
			},
			LLMFormatV2: llmFormatV2{
				Kind:            llm2.Kind,
				Atom:            string(llm2.Atom),
				TimestampMillis: llm2.TimestampMillis,
				MachineOrShard:  llm2.MachineOrShard,
				Sequence:        llm2.Sequence,
				Ghosted:         llm2.Ghosted,
			},
			TimeBinHour:     id.TimeBin(3600000),
			GhostedBytesHex: hex.EncodeToString(ghosted[:]),
		})
	}

	sgid := snid.NewSpatialFromCell(0x8c2a1072b59ffff, 0x1234567890ABCDEF)
	lat, lng := sgid.LatLng()
	file.Spatial = spatialCase{
		Atom:       string(snid.Space),
		BytesHex:   hex.EncodeToString(sgid[:]),
		Wire:       sgid.String(snid.Space),
		H3CellHex:  "8c2a1072b59ffff",
		Latitude:   lat,
		Longitude:  lng,
		Resolution: 12,
	}

	head := snid.TestID(snid.Matter, ts, 11)
	var semantic [16]byte
	for i := range semantic {
		semantic[i] = byte(255 - i)
	}
	nid := snid.NewNeuralFromHash(head, semantic)
	file.Neural = neuralCase{
		HeadHex:       hex.EncodeToString(head[:]),
		BytesHex:      hex.EncodeToString(nid[:]),
		SemanticHex:   hex.EncodeToString(semantic[:]),
		HammingToZero: nid.HammingDistance(snid.NeuralID{}),
	}

	var prev snid.LID
	copy(prev[:16], head[:])
	payload := []byte("payload")
	key := []byte("0123456789abcdef")
	lid, err := snid.NewLIDWithHead(head, prev, payload, key)
	if err != nil {
		panic(err)
	}
	file.Ledger = ledgerCase{
		HeadHex:    hex.EncodeToString(head[:]),
		BytesHex:   hex.EncodeToString(lid[:]),
		PrevHex:    hex.EncodeToString(prev[:]),
		PayloadHex: hex.EncodeToString(payload),
		KeyHex:     hex.EncodeToString(key),
	}
	blake3LID, err := snid.LIDBLAKE3(head, prev, payload, key)
	if err != nil {
		panic(err)
	}
	file.Ledger.Blake3Hex = hex.EncodeToString(blake3LID[:])

	var scenario [16]byte
	for i := range scenario {
		scenario[i] = byte(i + 1)
	}
	wid := snid.NewWIDFromHash(head, scenario)
	w0, w1, w2, w3 := wid.ToTensor256Words()
	file.World = worldCase{
		HeadHex:     hex.EncodeToString(head[:]),
		BytesHex:    hex.EncodeToString(wid[:]),
		ScenarioHex: hex.EncodeToString(scenario[:]),
		TensorWords: [4]int64{w0, w1, w2, w3},
	}

	var edgeHash [16]byte
	for i := range edgeHash {
		edgeHash[i] = byte(16 - i)
	}
	xid := snid.NewXIDFromHash(head, edgeHash)
	x0, x1, x2, x3 := xid.ToTensor256Words()
	file.Edge = edgeCase{
		HeadHex:     hex.EncodeToString(head[:]),
		BytesHex:    hex.EncodeToString(xid[:]),
		EdgeHex:     hex.EncodeToString(edgeHash[:]),
		TensorWords: [4]int64{x0, x1, x2, x3},
	}

	actor := snid.TestID(snid.Identity, ts, 19)
	resource := []byte("neighbors/booking:read")
	capability := []byte("read")
	kid, err := snid.NewKIDWithHead(head, actor, resource, capability, key)
	if err != nil {
		panic(err)
	}
	k0, k1, k2, k3 := kid.ToTensor256Words()
	file.Capability = capabilityCase{
		HeadHex:       hex.EncodeToString(head[:]),
		ActorHex:      hex.EncodeToString(actor[:]),
		ResourceHex:   hex.EncodeToString(resource),
		CapabilityHex: hex.EncodeToString(capability),
		KeyHex:        hex.EncodeToString(key),
		BytesHex:      hex.EncodeToString(kid[:]),
		TensorWords:   [4]int64{k0, k1, k2, k3},
	}

	eid := snid.NewEphemeralAt(uint64(ts.UnixMilli()), 0x00FF)
	eidBytes := eid.Bytes()
	file.Ephemeral = eidCase{
		BytesHex:        hex.EncodeToString(eidBytes[:]),
		TimestampMillis: uint64(eid.Time().UnixMilli()),
		Counter:         eid.Counter(),
	}

	var content [32]byte
	for i := range content {
		content[i] = byte(i)
	}
	bid := snid.NewBIDWithTopology(head, content)
	file.BID = bidCase{
		TopologyHex: hex.EncodeToString(head[:]),
		ContentHex:  hex.EncodeToString(content[:]),
		Wire:        bid.WireFormat(),
		R2Key:       bid.R2Key(),
		Neo4jID:     bid.Neo4jID(),
	}
	file.Compatibility = compatibilityCase{
		BytesHex: file.Core[0].BytesHex,
		Wire:     file.Core[0].Wire,
	}

	// UUIDv7 compatibility vectors
	uuidv7ID := snid.TestID(snid.Matter, ts, 1)
	uuidv7UUID := uuidv7ID.UUID()
	// Extract variant from bits 64-65 (byte 8, bits 6-7)
	variant := (uuidv7ID[8] >> 6) & 0b11
	file.UUIDv7 = uuidv7Case{
		BytesHex:        hex.EncodeToString(uuidv7ID[:]),
		UUIDString:      uuidv7UUID.String(),
		TimestampMillis: uuidv7ID.Time().UnixMilli(),
		Version:         int(uuidv7ID.Version()),
		Variant:         int(variant),
	}

	file.Negative = negativeCase{
		InvalidAtomWire:     "BAD:" + file.Core[0].Wire[4:],
		InvalidBinaryHex:    "001122",
		InvalidWireChecksum: file.Core[0].Wire[:len(file.Core[0].Wire)-1] + invalidChecksumChar(file.Core[0].Wire[len(file.Core[0].Wire)-1]),
		InvalidAdapterHex:   "0011223344",
	}

	f, err := os.Create(*outPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(file); err != nil {
		panic(err)
	}
}

func invalidChecksumChar(current byte) string {
	if current != '1' {
		return "1"
	}
	return "2"
}

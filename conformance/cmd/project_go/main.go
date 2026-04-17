package main

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	snid "github.com/neighbor/snid/go"
)

type vectorFile struct {
	Core       []coreVector     `json:"core"`
	Spatial    spatialVector    `json:"spatial"`
	Neural     neuralVector     `json:"neural"`
	Ledger     ledgerVector     `json:"ledger"`
	World      worldVector      `json:"world"`
	Edge       edgeVector       `json:"edge"`
	Capability capabilityVector `json:"capability"`
	Ephemeral  eidVector        `json:"ephemeral"`
	BID        bidVector        `json:"bid"`
}

type coreVector struct {
	Atom        string `json:"atom"`
	BytesHex    string `json:"bytes_hex"`
	Wire        string `json:"wire"`
	TensorHi    int64  `json:"tensor_hi"`
	TensorLo    int64  `json:"tensor_lo"`
	LLMFormat   llmV1  `json:"llm_format"`
}

type llmV1 struct {
	Atom            string `json:"atom"`
	TimestampMillis int64  `json:"timestamp_millis"`
	MachineOrShard  uint32 `json:"machine_or_shard"`
	Sequence        uint16 `json:"sequence"`
}

type spatialVector struct {
	Atom      string `json:"atom"`
	BytesHex  string `json:"bytes_hex"`
	Wire      string `json:"wire"`
	H3CellHex string `json:"h3_cell_hex"`
}

type neuralVector struct {
	HeadHex       string `json:"head_hex"`
	BytesHex      string `json:"bytes_hex"`
	SemanticHex   string `json:"semantic_hex"`
	HammingToZero int    `json:"hamming_to_zero"`
}

type ledgerVector struct {
	HeadHex    string `json:"head_hex"`
	BytesHex   string `json:"bytes_hex"`
	PrevHex    string `json:"prev_hex"`
	PayloadHex string `json:"payload_hex"`
	KeyHex     string `json:"key_hex"`
}

type worldVector struct {
	HeadHex     string   `json:"head_hex"`
	BytesHex    string   `json:"bytes_hex"`
	ScenarioHex string   `json:"scenario_hex"`
	TensorWords [4]int64 `json:"tensor_words"`
}

type edgeVector struct {
	HeadHex     string   `json:"head_hex"`
	BytesHex    string   `json:"bytes_hex"`
	EdgeHex     string   `json:"edge_hex"`
	TensorWords [4]int64 `json:"tensor_words"`
}

type capabilityVector struct {
	HeadHex       string   `json:"head_hex"`
	ActorHex      string   `json:"actor_hex"`
	ResourceHex   string   `json:"resource_hex"`
	CapabilityHex string   `json:"capability_hex"`
	KeyHex        string   `json:"key_hex"`
	BytesHex      string   `json:"bytes_hex"`
	TensorWords   [4]int64 `json:"tensor_words"`
}

type eidVector struct {
	BytesHex        string `json:"bytes_hex"`
	TimestampMillis uint64 `json:"timestamp_millis"`
	Counter         uint16 `json:"counter"`
}

type bidVector struct {
	TopologyHex string `json:"topology_hex"`
	ContentHex  string `json:"content_hex"`
	Wire        string `json:"wire"`
	R2Key       string `json:"r2_key"`
	Neo4jID     string `json:"neo4j_id"`
}

type projectionFile struct {
	Core       []coreProjection  `json:"core"`
	Spatial    spatialProjection `json:"spatial"`
	Neural     neuralProjection  `json:"neural"`
	Ledger     ledgerProjection  `json:"ledger"`
	World      tensorProjection  `json:"world"`
	Edge       tensorProjection  `json:"edge"`
	Capability capabilityProj    `json:"capability"`
	Ephemeral  eidProjection     `json:"ephemeral"`
	BID        bidProjection     `json:"bid"`
}

type coreProjection struct {
	BytesHex  string `json:"bytes_hex"`
	Wire      string `json:"wire"`
	TensorHi  int64  `json:"tensor_hi"`
	TensorLo  int64  `json:"tensor_lo"`
	Atom      string `json:"atom"`
	Timestamp int64  `json:"timestamp_millis"`
	Machine   uint32 `json:"machine_or_shard"`
	Sequence  uint16 `json:"sequence"`
}

type spatialProjection struct {
	BytesHex  string `json:"bytes_hex"`
	Wire      string `json:"wire"`
	H3CellHex string `json:"h3_cell_hex"`
}

type neuralProjection struct {
	BytesHex      string `json:"bytes_hex"`
	HammingToZero int    `json:"hamming_to_zero"`
}

type ledgerProjection struct {
	BytesHex string `json:"bytes_hex"`
}

type tensorProjection struct {
	BytesHex     string   `json:"bytes_hex"`
	TensorWords  [4]int64 `json:"tensor_words"`
}

type capabilityProj struct {
	BytesHex    string   `json:"bytes_hex"`
	TensorWords [4]int64 `json:"tensor_words"`
	Verified    bool     `json:"verified"`
}

type eidProjection struct {
	BytesHex        string `json:"bytes_hex"`
	TimestampMillis uint64 `json:"timestamp_millis"`
	Counter         uint16 `json:"counter"`
}

type bidProjection struct {
	Wire    string `json:"wire"`
	R2Key   string `json:"r2_key"`
	Neo4jID string `json:"neo4j_id"`
}

func mustDecodeHex(s string) []byte {
	raw, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return raw
}

func mustID(hexString string) snid.ID {
	id, err := snid.FromBytes(mustDecodeHex(hexString))
	if err != nil {
		panic(err)
	}
	return id
}

func mustLID(hexString string) snid.LID {
	var lid snid.LID
	copy(lid[:], mustDecodeHex(hexString))
	return lid
}

func main() {
	raw, err := os.ReadFile(filepath.Join("..", "..", "vectors.json"))
	if err != nil {
		panic(err)
	}

	var vectors vectorFile
	if err := json.Unmarshal(raw, &vectors); err != nil {
		panic(err)
	}

	out := projectionFile{
		Core: make([]coreProjection, 0, len(vectors.Core)),
	}

	for _, c := range vectors.Core {
		id := mustID(c.BytesHex)
		llm := id.ToLLMFormat(snid.Atom(c.Atom))
		hi, lo := id.ToTensorWords()
		out.Core = append(out.Core, coreProjection{
			BytesHex:  hex.EncodeToString(id[:]),
			Wire:      id.String(snid.Atom(c.Atom)),
			TensorHi:  hi,
			TensorLo:  lo,
			Atom:      string(llm.Atom),
			Timestamp: llm.TimestampMillis,
			Machine:   llm.MachineOrShard,
			Sequence:  llm.Sequence,
		})
	}

	sgid := mustID(vectors.Spatial.BytesHex)
	out.Spatial = spatialProjection{
		BytesHex:  hex.EncodeToString(sgid[:]),
		Wire:      sgid.String(snid.Atom(vectors.Spatial.Atom)),
		H3CellHex: sgid.ExtractLocation().String(),
	}

	head := mustID(vectors.Neural.HeadHex)
	var semantic [16]byte
	copy(semantic[:], mustDecodeHex(vectors.Neural.SemanticHex))
	nid := snid.NewNeuralFromHash(head, semantic)
	out.Neural = neuralProjection{
		BytesHex:      hex.EncodeToString(nid[:]),
		HammingToZero: nid.HammingDistance(snid.NeuralID{}),
	}

	prev := mustLID(vectors.Ledger.PrevHex)
	payload := mustDecodeHex(vectors.Ledger.PayloadHex)
	key := mustDecodeHex(vectors.Ledger.KeyHex)
	lid := mustLID(vectors.Ledger.BytesHex)
	out.Ledger = ledgerProjection{
		BytesHex: hex.EncodeToString(lid[:]),
	}
	if !lid.Verify(prev, payload, key) {
		panic("ledger verify failed")
	}

	var scenario [16]byte
	copy(scenario[:], mustDecodeHex(vectors.World.ScenarioHex))
	wid := snid.NewWID(mustID(vectors.World.HeadHex), scenario)
	w0, w1, w2, w3 := wid.ToTensor256Words()
	out.World = tensorProjection{
		BytesHex:    hex.EncodeToString(wid[:]),
		TensorWords: [4]int64{w0, w1, w2, w3},
	}

	var edge [16]byte
	copy(edge[:], mustDecodeHex(vectors.Edge.EdgeHex))
	xid := snid.NewXID(mustID(vectors.Edge.HeadHex), edge)
	x0, x1, x2, x3 := xid.ToTensor256Words()
	out.Edge = tensorProjection{
		BytesHex:    hex.EncodeToString(xid[:]),
		TensorWords: [4]int64{x0, x1, x2, x3},
	}

	kid, err := snid.NewKIDForCapability(
		mustID(vectors.Capability.HeadHex),
		mustID(vectors.Capability.ActorHex),
		mustDecodeHex(vectors.Capability.ResourceHex),
		mustDecodeHex(vectors.Capability.CapabilityHex),
		mustDecodeHex(vectors.Capability.KeyHex),
	)
	if err != nil {
		panic(err)
	}
	k0, k1, k2, k3 := kid.ToTensor256Words()
	out.Capability = capabilityProj{
		BytesHex:    hex.EncodeToString(kid[:]),
		TensorWords: [4]int64{k0, k1, k2, k3},
		Verified: kid.Verify(
			mustID(vectors.Capability.ActorHex),
			mustDecodeHex(vectors.Capability.ResourceHex),
			mustDecodeHex(vectors.Capability.CapabilityHex),
			mustDecodeHex(vectors.Capability.KeyHex),
		),
	}

	eid := snid.NewEphemeralAt(vectors.Ephemeral.TimestampMillis, vectors.Ephemeral.Counter)
	eidBytes := eid.Bytes()
	out.Ephemeral = eidProjection{
		BytesHex:        hex.EncodeToString(eidBytes[:]),
		TimestampMillis: uint64(eid.Time().UnixMilli()),
		Counter:         eid.Counter(),
	}

	bid, err := snid.NewBIDFromHash(mustDecodeHex(vectors.BID.ContentHex))
	if err != nil {
		panic(err)
	}
	bid.Topology = mustID(vectors.BID.TopologyHex)
	out.BID = bidProjection{
		Wire:    bid.WireFormat(),
		R2Key:   bid.R2Key(),
		Neo4jID: bid.Neo4jID(),
	}

	encoded, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(encoded)
}

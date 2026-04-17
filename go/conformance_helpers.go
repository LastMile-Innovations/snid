package snid

import "encoding/binary"

// FromParts assembles a deterministic core SNID from explicit layout components.
func FromParts(ms uint64, seq uint16, machineID uint32, entropy uint64) ID {
	return assemble(ms, uint64(seq), machineID, 0, uint32(entropy))
}

// NewSpatialFromCell creates a deterministic SGID from an H3 cell and entropy tail.
func NewSpatialFromCell(cell uint64, entropy uint64) ID {
	var id ID
	binary.BigEndian.PutUint64(id[:8], cell)
	originalHighNibble := id[6] >> 4
	id[6] = 0x80 | (id[6] & 0x0F)
	binary.BigEndian.PutUint64(id[8:], entropy)
	id[8] = (id[8] & 0x3F) | 0x80
	id[14] = spatialMarkerByte
	id[15] = spatialTailNibble | (originalHighNibble & 0x0F)
	return id
}

// NewNeuralFromHash assembles a deterministic NID from a head and 128-bit semantic tail.
func NewNeuralFromHash(head ID, semanticHash [16]byte) NeuralID {
	var nid NeuralID
	copy(nid[:16], head[:])
	copy(nid[16:], semanticHash[:])
	return nid
}

// NewLIDWithHead assembles a deterministic LID using an explicit SNID head.
func NewLIDWithHead(head ID, prev LID, payload []byte, key []byte) (LID, error) {
	if len(key) == 0 {
		return LID{}, ErrInvalidLIDKey
	}
	var lid LID
	tail := computeLIDTail(head, prev, payload, key)
	copy(lid[:16], head[:])
	copy(lid[16:], tail[:])
	return lid, nil
}

// NewBIDWithTopology creates a BID from an explicit topology ID and content hash.
func NewBIDWithTopology(topology ID, contentHash [32]byte) BID {
	return BID{Topology: topology, Content: contentHash}
}

// NewEphemeralAt assembles a deterministic EID from explicit components.
func NewEphemeralAt(unixMillis uint64, counter uint16) EID {
	return EID((unixMillis << 16) | uint64(counter))
}

// NewWIDFromHash assembles a deterministic WID from an explicit head and scenario tail.
func NewWIDFromHash(head ID, scenarioHash [16]byte) WID {
	return NewWID(head, scenarioHash)
}

// NewXIDFromHash assembles a deterministic XID from an explicit head and edge tail.
func NewXIDFromHash(head ID, edgeHash [16]byte) XID {
	return NewXID(head, edgeHash)
}

// NewKIDWithHead assembles a deterministic KID using explicit inputs.
func NewKIDWithHead(head ID, actor ID, resource []byte, capability []byte, key []byte) (KID, error) {
	return NewKIDForCapability(head, actor, resource, capability, key)
}

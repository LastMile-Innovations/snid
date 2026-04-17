package snid

import "encoding/binary"

// generateAssetID creates the "Instance ID" (Version 10 / 0xA)
// [Cat 32] [Owner 32] [Serial 48] [Ver 4] [Var 4] [Tag 8]
func generateAssetID(cid ID, tenantID, serial string) ID {
	var id ID

	// 1. Copy Taxonomy (Catalog Prefix) - Bytes 0-3
	copy(id[0:4], cid[0:4])

	// 2. Owner Hash - Bytes 4-7
	binary.BigEndian.PutUint32(id[4:8], fnv1a(tenantID))

	// 3. Serial/Entropy - Bytes 8-13 (48 bits)
	var hash uint64
	if serial == "" {
		InitAdaptive()
		hash = adaptive.nextEntropy()
	} else {
		hash = fnv1a64Upper(serial)
	}
	// Write 48 bits
	id[8] = byte(hash >> 40)
	id[9] = byte(hash >> 32)
	id[10] = byte(hash >> 24)
	id[11] = byte(hash >> 16)
	id[12] = byte(hash >> 8)
	id[13] = byte(hash)

	// 4. Version A (Asset) in the UUID version nibble (byte 6 high bits).
	id[6] = (id[6] & 0x0F) | 0xA0

	return id
}

// generateCatalogID creates the "Taxonomy ID" (Version 9 / 0x9)
func generateCatalogID(category, brand, specs string) ID {
	var id ID

	// 1. Category Hash (Bytes 0-3)
	binary.BigEndian.PutUint32(id[0:4], fnv1a32Upper(category))

	// 2. Brand Hash (Bytes 4-7)
	binary.BigEndian.PutUint32(id[4:8], fnv1a32Upper(brand))

	// 3. Semantic Hash (Bytes 8-15)
	binary.BigEndian.PutUint64(id[8:], fnv1a64Lower(specs))

	// 4. Version 9 (Catalog) in Byte 6 high nibble
	// We need to overwrite Byte 6 carefully if we put Brand Hash there.
	// Byte 4,5,6,7.
	// Byte 6 contains part of Brand Hash.
	// We FORCE Version 9 into Byte 6 high nibble.
	id[6] = (id[6] & 0x0F) | 0x90

	return id
}

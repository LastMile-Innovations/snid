package snid

import (
	"encoding/binary"

	"github.com/uber/h3-go/v4"
)

// NewSpatialTyped generates a spatial ID while documenting intent via atom.
func NewSpatialTyped(atom Atom, lat, lng float64, res int) ID {
	_ = CanonicalAtom(atom)
	return NewSpatialPrecise(lat, lng, res)
}

// H3Cell returns the H3 cell for spatial IDs.
func (id ID) H3Cell() h3.Cell {
	return id.ExtractLocation()
}

// H3String returns the H3 cell string for spatial IDs.
func (id ID) H3String() string {
	cell := id.ExtractLocation()
	if cell == 0 {
		return ""
	}
	return cell.String()
}

// LatLng returns the approximate centerpoint of a spatial ID.
func (id ID) LatLng() (float64, float64) {
	cell := id.ExtractLocation()
	if cell == 0 {
		return 0, 0
	}
	ll, err := cell.LatLng()
	if err != nil {
		return 0, 0
	}
	return ll.Lat, ll.Lng
}

// SpatialParent remaps this ID to a parent H3 resolution while preserving tail entropy.
func (id ID) SpatialParent(res int) ID {
	cell := id.ExtractLocation()
	if cell == 0 {
		return Zero
	}
	parent, err := cell.Parent(res)
	if err != nil {
		return Zero
	}

	out := id
	binary.BigEndian.PutUint64(out[:8], uint64(parent))
	originalHighNibble := out[6] >> 4
	out[6] = 0x80 | (out[6] & 0x0F)
	out[14] = spatialMarkerByte
	out[15] = spatialTailNibble | (originalHighNibble & 0x0F)
	out[8] = (out[8] & 0x3F) | 0x80
	return out
}

// SpatialRange returns a lexicographic start/end range for IDs anchored to a cell.
func SpatialRange(cell h3.Cell) (ID, ID) {
	var start ID
	binary.BigEndian.PutUint64(start[:8], uint64(cell))
	originalHighNibble := start[6] >> 4
	start[6] = 0x80 | (start[6] & 0x0F)
	start[8] = 0x80 // variant bits
	start[14] = spatialMarkerByte
	start[15] = spatialTailNibble | (originalHighNibble & 0x0F)

	end := start
	for i := 8; i <= 13; i++ {
		end[i] = 0xFF
	}
	end = incrementID(end)
	return start, end
}

func incrementID(id ID) ID {
	for i := len(id) - 1; i >= 0; i-- {
		id[i]++
		if id[i] != 0 {
			return id
		}
	}
	return id
}

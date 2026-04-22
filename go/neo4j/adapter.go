package neo4j

import (
	"fmt"

	snid "github.com/LastMile-Innovations/snid"
)

// MarshalProperty converts a SNID into the canonical Neo4j binary property representation.
func MarshalProperty(id snid.ID) []byte {
	raw := make([]byte, 16)
	copy(raw, id[:])
	return raw
}

// MarshalProperty32 converts a 32-byte composite SNID family member into binary storage form.
func MarshalProperty32(raw [32]byte) []byte {
	out := make([]byte, 32)
	copy(out, raw[:])
	return out
}

// UnmarshalProperty converts a Neo4j property payload back into a SNID.
func UnmarshalProperty(value any) (snid.ID, error) {
	switch v := value.(type) {
	case []byte:
		return snid.FromBytes(v)
	case [16]byte:
		return snid.FromBytes(v[:])
	case string:
		if len(v) == 32 {
			return snidFromHex(v)
		}
		id, _, err := snid.FromString(v)
		return id, err
	default:
		return snid.Zero, fmt.Errorf("snid/neo4j: unsupported property type %T", value)
	}
}

// BindID writes a SNID into a parameter map using canonical binary storage.
func BindID(params map[string]any, key string, id snid.ID) map[string]any {
	if params == nil {
		params = make(map[string]any, 1)
	}
	params[key] = MarshalProperty(id)
	return params
}

// BindBinary writes an arbitrary canonical binary identifier into a parameter map.
func BindBinary(params map[string]any, key string, raw []byte) map[string]any {
	if params == nil {
		params = make(map[string]any, 1)
	}
	copied := make([]byte, len(raw))
	copy(copied, raw)
	params[key] = copied
	return params
}

// WireDebugValue returns the human-readable compatibility/debug rendering.
func WireDebugValue(id snid.ID, atom snid.Atom) string {
	return id.String(atom)
}

func snidFromHex(value string) (snid.ID, error) {
	var out snid.ID
	for i := 0; i < 16; i++ {
		hi, ok := fromHexNibble(value[i*2])
		if !ok {
			return snid.Zero, snid.ErrInvalidFormat
		}
		lo, ok := fromHexNibble(value[i*2+1])
		if !ok {
			return snid.Zero, snid.ErrInvalidFormat
		}
		out[i] = (hi << 4) | lo
	}
	return out, nil
}

func fromHexNibble(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	default:
		return 0, false
	}
}

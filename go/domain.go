package snid

import (
	"strings"
)

// =============================================================================
// DOMAIN CONSTRUCTORS (The API)
// =============================================================================

// NewItem creates an Asset ID from a Catalog ID.
func NewItem(cid ID, tenantID, serial string) ID {
	return NewAsset(cid, tenantID, serial)
}

// NewConcept creates a Catalog ID from product details.
func NewConcept(category, brand, specs string) ID {
	return NewCatalog(category, brand, specs)
}

// We need to re-expose CommonCategories if it was public.
var CommonCategories = map[string]uint32{
	"POWER_DRILL": 27111703,
	"MILK":        50131701,
	"COMPUTER":    43211500,
	"GENERAL":     99000000,
}

// SanitizeAlias is a helper for cleaning alias strings.
func SanitizeAlias(s string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		if r == ' ' {
			return '-'
		}
		return -1
	}, s)
}

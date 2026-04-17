package snid

// graph.go: Type-safe ID constructors and validators for the Identity Lattice.
// These helpers enforce type-correctness at compile time and provide
// utilities for binary database storage and routing.

import (
	"time"
)

// =============================================================================
// TYPE-SAFE ID CONSTRUCTORS
// =============================================================================

// NewUser generates a new User ID with IAM atom prefix.
// Use for: Registration, User creation.
func NewUser() ID { return NewFast() }

// NewPersona generates a new Persona ID with IAM atom prefix.
// Use for: Creating new personas/masks for a user.
func NewPersona() ID { return NewFast() }

// NewTenant generates a new Tenant ID with TEN atom prefix.
// Use for: Creating households, businesses, organizations.
func NewTenant() ID { return NewFast() }

// NewGroup generates a new Group ID with TEN atom prefix.
// Use for: Creating communities, guilds, collectives.
func NewGroup() ID { return NewFast() }

// =============================================================================
// TYPE-SAFE STRING SERIALIZATION
// =============================================================================

// UserString returns the ID as "IAM:..." format.
func (id ID) UserString() string { return id.String(Identity) }

// PersonaString returns the ID as "IAM:..." format.
func (id ID) PersonaString() string { return id.String(Persona) }

// TenantString returns the ID as "TEN:..." format.
func (id ID) TenantString() string { return id.String(Tenant) }

// GroupString returns the ID as "TEN:..." format.
func (id ID) GroupString() string { return id.String(Group) }

// =============================================================================
// ATOM TYPE DETECTION (O(1) from prefix)
// =============================================================================

// AtomFromString extracts the atom type from a string ID.
// Returns empty string if format is invalid.
// Cost: O(1) - just reads first 3-4 bytes.
func AtomFromString(s string) Atom {
	if len(s) < 4 {
		return ""
	}
	delim := s[3]
	if delim != ':' && delim != '_' {
		return ""
	}
	if delim == '_' && !AcceptUnderscore() {
		return ""
	}
	atom := Atom(s[:3])
	if !IsValidAtom(atom) {
		return ""
	}
	return CanonicalAtom(atom)
}

// IsUserID checks if the string is a User ID (IAM: prefix).
func IsUserID(s string) bool { return AtomFromString(s) == Identity }

// IsPersonaID checks if the string is a Persona ID (IAM: prefix, accepts ACT: legacy).
func IsPersonaID(s string) bool { return AtomFromString(s) == Persona }

// IsTenantID checks if the string is a Tenant ID (TEN: prefix).
func IsTenantID(s string) bool { return AtomFromString(s) == Tenant }

// IsGroupID checks if the string is a Group ID (TEN: prefix, accepts GRP: legacy).
func IsGroupID(s string) bool { return AtomFromString(s) == Group }

// =============================================================================
// NEO4J BINARY HELPERS
// =============================================================================

// Neo4jBytes returns the ID as a byte slice for Neo4j APOC binary storage.
// Use with: apoc.convert.toHexString(id) for compact storage in properties.
func (id ID) Neo4jBytes() []byte {
	return id[:]
}

// Neo4jHex returns the ID as a hex string (32 chars) for Neo4j storage.
// More compact than Base58+checksum (26 chars) in some scenarios.
func (id ID) Neo4jHex() string {
	const hextable = "0123456789abcdef"
	dst := make([]byte, 32)
	for i, v := range id {
		dst[i*2] = hextable[v>>4]
		dst[i*2+1] = hextable[v&0x0f]
	}
	return string(dst)
}

// =============================================================================
// TIME-BASED UTILITIES (K-Sortable Benefits)
// =============================================================================

// CreatedBefore returns true if this ID was created before the other.
// Uses embedded timestamp - no database lookup required.
// Cost: O(1), ~5ns.
func (id ID) CreatedBefore(other ID) bool {
	return id.Before(other)
}

// CreatedAfter returns true if this ID was created after the other.
func (id ID) CreatedAfter(other ID) bool {
	return id.After(other)
}

// IsFresh returns true if the ID was created within the last duration.
// Useful for: Rate limiting, session validation, recently-created checks.
func (id ID) IsFresh(maxAgeMs int64) bool {
	return id.Time().UnixMilli() > (time.Now().UnixMilli() - maxAgeMs)
}

// =============================================================================
// BATCH GENERATION (High-Throughput Paths)
// =============================================================================

// GenerateUserBatch creates n User IDs efficiently using the global streamer.
func GenerateUserBatch(n int) []string {
	result := make([]string, n)
	for i := range n {
		result[i] = NewFast().String(Identity)
	}
	return result
}

// GenerateTenantBatch creates n Tenant IDs efficiently.
func GenerateTenantBatch(n int) []string {
	result := make([]string, n)
	for i := range n {
		result[i] = NewFast().String(Tenant)
	}
	return result
}

// GeneratePersonaBatch creates n Persona IDs efficiently.
func GeneratePersonaBatch(n int) []string {
	result := make([]string, n)
	for i := range n {
		result[i] = NewFast().String(Persona)
	}
	return result
}

// =============================================================================
// ROUTING HELPERS (Geographic Sharding)
// =============================================================================

// ShardKey returns a consistent hash key for sharding.
// Uses first 8 bytes (timestamp + random) for distribution.
// Benefit: Time-locality for hot data, even distribution across shards.
func (id ID) ShardKey(numShards int) int {
	if numShards <= 0 {
		return 0
	}
	// Use bytes 0-7 as uint64, mod by shard count
	h := uint64(id[0])<<56 | uint64(id[1])<<48 | uint64(id[2])<<40 | uint64(id[3])<<32 |
		uint64(id[4])<<24 | uint64(id[5])<<16 | uint64(id[6])<<8 | uint64(id[7])
	return int(h % uint64(numShards))
}

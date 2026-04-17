package snid

// Atom defines the 3-character prefix for Spacetime Neighbor IDs.
type Atom string

const (
	// The 6 Atoms of Reality (canonical v8.4).
	Identity Atom = "IAM"
	Tenant   Atom = "TEN"
	Matter   Atom = "MAT"
	Space    Atom = "LOC" // SGID space anchor
	Time     Atom = "CHR" // Chronos
	Ledger   Atom = "LED" // Iron Ledger

	// Higher-order engines.
	Legal     Atom = "LEG"
	Trust     Atom = "TRU"
	Kinetic   Atom = "KIN"
	Cognition Atom = "COG"
	Semantic  Atom = "SEM"
	Stream    Atom = "STM" // Stream Assets

	// System infrastructure.
	System  Atom = "SYS"
	Vault   Atom = "VLT" // Secure PII Vault Reference
	Key     Atom = "KEY" // Access key public identifier
	Event   Atom = "EVT"
	Session Atom = "SES"
)

const (
	// Legacy compatibility aliases (compile-time).
	Location    Atom = Space
	Transaction Atom = Ledger
	Object      Atom = Matter
	Schedule    Atom = Time
	Network     Atom = Trust
	Operations  Atom = Event
	Persona     Atom = Identity
	Group       Atom = Tenant
	Bio         Atom = Identity
	Atmosphere  Atom = Space
)

const (
	// Legacy wire prefixes accepted at parse time and normalized to canonical.
	LegacyObject      Atom = "OBJ"
	LegacyTransaction Atom = "TXN"
	LegacySchedule    Atom = "SCH"
	LegacyNetwork     Atom = "NET"
	LegacyOperations  Atom = "OPS"
	LegacyPersona     Atom = "ACT"
	LegacyGroup       Atom = "GRP"
	LegacyBio         Atom = "BIO"
	LegacyAtmosphere  Atom = "ATM"
)

// CanonicalAtom maps legacy atom tags to canonical v8.4 atoms.
func CanonicalAtom(atom Atom) Atom {
	switch atom {
	case LegacyObject:
		return Matter
	case LegacyTransaction:
		return Ledger
	case LegacySchedule:
		return Time
	case LegacyNetwork:
		return Trust
	case LegacyOperations:
		return Event
	case LegacyPersona:
		return Identity
	case LegacyGroup:
		return Tenant
	case LegacyBio:
		return Identity
	case LegacyAtmosphere:
		return Space
	default:
		return atom
	}
}

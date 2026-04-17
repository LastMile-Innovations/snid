package snid

// constructors.go - Type-safe constructors for SNID atoms.
//
// Canonical v8.4 ontology uses: IAM, TEN, MAT, LOC, CHR, LED,
// LEG, TRU, KIN, COG, SEM, SYS, EVT, SES.
//
// Compatibility note:
// Legacy helper names are preserved (`NewObject`, `NewTransaction`, etc.) but now
// emit canonical prefixes. Legacy wire prefixes remain parseable via CanonicalAtom.

// NewFastWithPrefix documents intent at callsites where the domain atom matters.
// Atoms are not encoded in the 16-byte ID itself; they are applied at serialization.
func NewFastWithPrefix(atom Atom) ID {
	_ = CanonicalAtom(atom)
	return NewFast()
}

// NewProjectedWithPrefix generates a tenant-sharded ID.
// REQ-SNI-054: Fractal Tenancy sharding key logic
func NewProjectedWithPrefix(atom Atom, tenantID string) ID {
	_ = CanonicalAtom(atom)
	return NewProjected(tenantID, 0)
}

// =============================================================================
// IDENTITY / TENANCY
// =============================================================================

// NewIdentityID generates a canonical IAM ID.
func NewIdentityID() ID { return NewFastWithPrefix(Identity) }

// NewIdentityIDForTenant generates a canonical IAM ID sharded for a tenant.
func NewIdentityIDForTenant(tenantID string) ID { return NewProjectedWithPrefix(Identity, tenantID) }

// NewTenantID generates a canonical TEN ID.
func NewTenantID() ID { return NewFastWithPrefix(Tenant) }

// =============================================================================
// MATTER / SPACE / TIME / VALUE
// =============================================================================

// NewMatter generates a canonical MAT ID.
func NewMatter() ID { return NewFastWithPrefix(Matter) }

// NewMatterForTenant generates a canonical MAT ID sharded for a tenant.
func NewMatterForTenant(tenantID string) ID { return NewProjectedWithPrefix(Matter, tenantID) }

// MatterString returns the ID as "MAT:...".
func (id ID) MatterString() string { return id.String(Matter) }

// IsMatterID checks if the string is a Matter ID (MAT or OBJ legacy).
func IsMatterID(s string) bool { return AtomFromString(s) == Matter }

// NewSpaceID generates a canonical LOC ID.
func NewSpaceID() ID { return NewFastWithPrefix(Space) }

// NewLocation is a compatibility wrapper for space IDs.
func NewLocation() ID { return NewSpaceID() }

// LocationString returns the ID as "LOC:...".
func (id ID) LocationString() string { return id.String(Location) }

// IsLocationID checks if the string is a space ID (LOC).
func IsLocationID(s string) bool { return AtomFromString(s) == Location }

// NewChronos generates a canonical CHR ID.
func NewChronos() ID { return NewFastWithPrefix(Time) }

// NewChronosForTenant generates a canonical CHR ID sharded for a tenant.
func NewChronosForTenant(tenantID string) ID { return NewProjectedWithPrefix(Time, tenantID) }

// ChronosString returns the ID as "CHR:...".
func (id ID) ChronosString() string { return id.String(Time) }

// IsChronosID checks if the string is a chronos ID (CHR or SCH legacy).
func IsChronosID(s string) bool { return AtomFromString(s) == Time }

// NewLedger generates a canonical LED ID.
func NewLedger() ID { return NewFastWithPrefix(Ledger) }

// NewLedgerForTenant generates a canonical LED ID sharded for a tenant.
func NewLedgerForTenant(tenantID string) ID { return NewProjectedWithPrefix(Ledger, tenantID) }

// LedgerString returns the ID as "LED:...".
func (id ID) LedgerString() string { return id.String(Ledger) }

// IsLedgerID checks if the string is a ledger ID (LED or TXN legacy).
func IsLedgerID(s string) bool { return AtomFromString(s) == Ledger }

// =============================================================================
// HIGHER-ORDER ENGINES
// =============================================================================

// NewLegal generates a canonical LEG ID.
func NewLegal() ID { return NewFastWithPrefix(Legal) }

// LegalString returns the ID as "LEG:...".
func (id ID) LegalString() string { return id.String(Legal) }

// IsLegalID checks if the string is a legal ID.
func IsLegalID(s string) bool { return AtomFromString(s) == Legal }

// NewTrust generates a canonical TRU ID.
func NewTrust() ID { return NewFastWithPrefix(Trust) }

// NewTrustForTenant generates a canonical TRU ID sharded for a tenant.
func NewTrustForTenant(tenantID string) ID { return NewProjectedWithPrefix(Trust, tenantID) }

// TrustString returns the ID as "TRU:...".
func (id ID) TrustString() string { return id.String(Trust) }

// IsTrustID checks if the string is a trust ID (TRU or NET legacy).
func IsTrustID(s string) bool { return AtomFromString(s) == Trust }

// NewKinetic generates a canonical KIN ID.
func NewKinetic() ID { return NewFastWithPrefix(Kinetic) }

// KineticString returns the ID as "KIN:...".
func (id ID) KineticString() string { return id.String(Kinetic) }

// IsKineticID checks if the string is a kinetic ID.
func IsKineticID(s string) bool { return AtomFromString(s) == Kinetic }

// NewCognition generates a canonical COG ID.
func NewCognition() ID { return NewFastWithPrefix(Cognition) }

// NewCognitionForTenant generates a canonical COG ID sharded for a tenant.
func NewCognitionForTenant(tenantID string) ID { return NewProjectedWithPrefix(Cognition, tenantID) }

// CognitionString returns the ID as "COG:...".
func (id ID) CognitionString() string { return id.String(Cognition) }

// IsCognitionID checks if the string is a cognition ID.
func IsCognitionID(s string) bool { return AtomFromString(s) == Cognition }

// NewSemantic generates a canonical SEM ID.
func NewSemantic() ID { return NewFastWithPrefix(Semantic) }

// SemanticString returns the ID as "SEM:...".
func (id ID) SemanticString() string { return id.String(Semantic) }

// IsSemanticID checks if the string is a semantic ID.
func IsSemanticID(s string) bool { return AtomFromString(s) == Semantic }

// =============================================================================
// STREAM / MEDIA
// =============================================================================

// NewStream generates a canonical STM ID.
func NewStream() ID { return NewFastWithPrefix(Stream) }

// StreamString returns the ID as "STM:...".
func (id ID) StreamString() string { return id.String(Stream) }

// IsStreamID checks if the string is a stream ID.
func IsStreamID(s string) bool { return AtomFromString(s) == Stream }

// =============================================================================
// INFRASTRUCTURE
// =============================================================================

// NewSystem generates a canonical SYS ID.
func NewSystem() ID { return NewFastWithPrefix(System) }

// SystemString returns the ID as "SYS:...".
func (id ID) SystemString() string { return id.String(System) }

// IsSystemID checks if the string is a system ID.
func IsSystemID(s string) bool { return AtomFromString(s) == System }

// NewVaultID generates a canonical VLT ID.
func NewVaultID() ID { return NewFastWithPrefix(Vault) }

// VaultString returns the ID as "VLT:...".
func (id ID) VaultString() string { return id.String(Vault) }

// IsVaultID checks if the string is a vault ID.
func IsVaultID(s string) bool { return AtomFromString(s) == Vault }

// NewAccessKeyID generates a canonical KEY ID.
func NewAccessKeyID() ID { return NewFastWithPrefix(Key) }

// NewAccessKeyIDForTenant generates a tenant-projected KEY ID.
func NewAccessKeyIDForTenant(tenantID string) ID { return NewProjectedWithPrefix(Key, tenantID) }

// AccessKeyString returns the ID as "KEY:...".
func (id ID) AccessKeyString() string { return id.String(Key) }

// IsAccessKeyID checks if the string is an access key ID.
func IsAccessKeyID(s string) bool { return AtomFromString(s) == Key }

// NewEvent generates a canonical EVT ID.
func NewEvent() ID { return NewFastWithPrefix(Event) }

// NewEventForTenant generates a canonical EVT ID sharded for a tenant.
func NewEventForTenant(tenantID string) ID { return NewProjectedWithPrefix(Event, tenantID) }

// EventString returns the ID as "EVT:...".
func (id ID) EventString() string { return id.String(Event) }

// IsEventID checks if the string is an event ID (EVT or OPS legacy).
func IsEventID(s string) bool { return AtomFromString(s) == Event }

// NewSession generates a canonical SES ID.
func NewSession() ID { return NewFastWithPrefix(Session) }

// SessionString returns the ID as "SES:...".
func (id ID) SessionString() string { return id.String(Session) }

// IsSessionID checks if the string is a session ID.
func IsSessionID(s string) bool { return AtomFromString(s) == Session }

// =============================================================================
// LEGACY WRAPPERS (v8.4 compatibility)
// =============================================================================

// Deprecated: use NewMatter.
// NewObject generates a new object ID. Canonical prefix is MAT.
func NewObject() ID { return NewMatter() }

// ObjectString returns the ID as canonical matter prefix.
func (id ID) ObjectString() string { return id.String(Object) }

// IsObjectID checks if the string is an object/matter ID (MAT or OBJ legacy).
func IsObjectID(s string) bool { return AtomFromString(s) == Object }

// Deprecated: use NewLedger.
// NewTransaction generates a new ledger ID. Canonical prefix is LED.
func NewTransaction() ID { return NewLedger() }

// TransactionString returns the ID as canonical ledger prefix.
func (id ID) TransactionString() string { return id.String(Transaction) }

// IsTransactionID checks if the string is a ledger/transaction ID (LED or TXN legacy).
func IsTransactionID(s string) bool { return AtomFromString(s) == Transaction }

// Deprecated: use NewEvent.
// NewOperation generates a new event ID. Canonical prefix is EVT.
func NewOperation() ID { return NewEvent() }

// OperationString returns the ID as canonical event prefix.
func (id ID) OperationString() string { return id.String(Operations) }

// IsOperationID checks if the string is an operation/event ID (EVT or OPS legacy).
func IsOperationID(s string) bool { return AtomFromString(s) == Operations }

// Deprecated: use NewChronos.
// NewSchedule generates a new chronos ID. Canonical prefix is CHR.
func NewSchedule() ID { return NewChronos() }

// ScheduleString returns the ID as canonical chronos prefix.
func (id ID) ScheduleString() string { return id.String(Schedule) }

// IsScheduleID checks if the string is a schedule/chronos ID (CHR or SCH legacy).
func IsScheduleID(s string) bool { return AtomFromString(s) == Schedule }

// Deprecated: use NewTrust.
// NewNetwork generates a new trust ID. Canonical prefix is TRU.
func NewNetwork() ID { return NewTrust() }

// NetworkString returns the ID as canonical trust prefix.
func (id ID) NetworkString() string { return id.String(Network) }

// IsNetworkID checks if the string is a network/trust ID (TRU or NET legacy).
func IsNetworkID(s string) bool { return AtomFromString(s) == Network }

// Deprecated: use NewIdentityID.
// NewBio generates a new identity ID. Canonical prefix is IAM.
func NewBio() ID { return NewIdentityID() }

// BioString returns the ID as canonical identity prefix.
func (id ID) BioString() string { return id.String(Bio) }

// IsBioID checks if the string is a bio/identity ID (IAM or BIO legacy).
func IsBioID(s string) bool { return AtomFromString(s) == Bio }

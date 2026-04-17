package snid

// CAS node constructors keep callsites explicit while preserving
// SNID canonical atoms. CAS data nodes are matter; commit lineage is event.

// NewCASRawID generates a SNID for CAS_RawNode primary keys.
func NewCASRawID() ID { return NewMatter() }

// CASRawString returns a wire ID for CAS_RawNode.
func (id ID) CASRawString() string { return id.String(Matter) }

// NewCASFileID generates a SNID for CAS_FileNode primary keys.
func NewCASFileID() ID { return NewMatter() }

// CASFileString returns a wire ID for CAS_FileNode.
func (id ID) CASFileString() string { return id.String(Matter) }

// NewCASDirectoryID generates a SNID for CAS_DirectoryNode primary keys.
func NewCASDirectoryID() ID { return NewMatter() }

// CASDirectoryString returns a wire ID for CAS_DirectoryNode.
func (id ID) CASDirectoryString() string { return id.String(Matter) }

// NewCASManifestID generates a SNID for CAS_ManifestNode primary keys.
func NewCASManifestID() ID { return NewMatter() }

// CASManifestString returns a wire ID for CAS_ManifestNode.
func (id ID) CASManifestString() string { return id.String(Matter) }

// NewCASCommitID generates a SNID for CAS_CommitNode primary keys.
func NewCASCommitID() ID { return NewEvent() }

// CASCommitString returns a wire ID for CAS_CommitNode.
func (id ID) CASCommitString() string { return id.String(Event) }

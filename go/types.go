package snid

import (
	"crypto/aes"
	"crypto/hmac"
	"encoding/binary"
	"errors"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/mr-tron/base58"
)

var (
	ErrInvalidSignature = errors.New("snid: invalid signature")
	ErrExpired          = errors.New("snid: grant has expired")
)

func validAESKeyLength(secret []byte) bool {
	switch len(secret) {
	case 16, 24, 32:
		return true
	default:
		return false
	}
}

// --- SECURITY: GRANT ID ---

// GrantID provides tamper-proof capability access.
// If ExpiresAt is Zero, the grant is permanent (Integrity-only).
type GrantID struct {
	ID        ID
	Atom      Atom
	Signature [16]byte
	ExpiresAt time.Time
}

// NewGrant creates a signed capability grant with optional expiration.
func NewGrant(atom Atom, ttl time.Duration, secret []byte) (g GrantID) {
	doSecret(func() {
		atom = CanonicalAtom(atom)
		id := NewFast()
		var exp time.Time
		if ttl > 0 {
			exp = time.Now().Add(ttl)
		}
		g = GrantID{
			ID:        id,
			Atom:      atom,
			ExpiresAt: exp,
			Signature: signGrant(id, atom, exp, secret),
		}
	})
	return
}

// Verify checks signature validity and expiration for the grant.
func (g GrantID) Verify(secret []byte) (ok bool) {
	if !validAESKeyLength(secret) {
		return false
	}
	doSecret(func() {
		if !g.ExpiresAt.IsZero() && time.Now().After(g.ExpiresAt) {
			ok = false
			return
		}
		atom := CanonicalAtom(g.Atom)
		if atom != "" {
			expected := signGrant(g.ID, atom, g.ExpiresAt, secret)
			if hmac.Equal(g.Signature[:], expected[:]) {
				ok = true
				return
			}
		}
		// Backward compatibility for pre-v8.4 grants that did not bind atom.
		expected := signGrantLegacy(g.ID, g.ExpiresAt, secret)
		ok = hmac.Equal(g.Signature[:], expected[:])
	})
	return
}

// String renders a grant token including optional expiry and signature.
func (g GrantID) String(atom Atom) string {
	useAtom := CanonicalAtom(atom)
	if g.Atom != "" {
		useAtom = CanonicalAtom(g.Atom)
	}
	buf := make([]byte, 0, 64)
	buf = g.ID.AppendTo(buf, useAtom)
	if !g.ExpiresAt.IsZero() {
		buf = append(buf, '@')
		buf = strconv.AppendInt(buf, g.ExpiresAt.Unix(), 10)
	}
	buf = append(buf, '.')
	return string(append(buf, encode16Base58(g.Signature)...))
}

// ParseGrant parses and verifies a serialized grant token.
func ParseGrant(s string, secret []byte) (GrantID, Atom, error) {
	dotIdx := strings.LastIndex(s, ".")
	if dotIdx == -1 {
		return GrantID{}, "", ErrInvalidFormat
	}

	sigPart := s[dotIdx+1:]
	mainPart := s[:dotIdx]

	var exp time.Time
	atIdx := strings.LastIndex(mainPart, "@")
	var idPart string

	if atIdx != -1 {
		idPart = mainPart[:atIdx]
		ts, err := strconv.ParseInt(mainPart[atIdx+1:], 10, 64)
		if err != nil {
			return GrantID{}, "", ErrInvalidFormat
		}
		exp = time.Unix(ts, 0).UTC()
	} else {
		idPart = mainPart
	}

	var id ID
	atom, err := id.Parse(idPart)
	if err != nil {
		return GrantID{}, "", err
	}

	sig, err := decode16Base58Bytes(sigPart)
	if err != nil {
		return GrantID{}, "", err
	}

	g := GrantID{ID: id, Atom: atom, ExpiresAt: exp, Signature: sig}
	if !g.Verify(secret) {
		return GrantID{}, "", ErrInvalidSignature
	}
	return g, atom, nil
}

// --- ROUTING: SCOPE ID ---

// ScopeID embeds a 32-bit logical group hash (e.g. Tenant, Region).
type ScopeID struct {
	ID    ID
	Scope string
}

// NewScope creates a scope-embedded ID for logical routing domains.
func NewScope(atom Atom, scope string) ScopeID {
	return ScopeID{ID: NewProjected(scope, 0), Scope: scope}
}

// NewScopeWithHash bypasses runtime string hashing.
// Use this for hot-path routing where the Scope (Tenant/Region) is known.
func NewScopeWithHash(atom Atom, scope string, hash uint32) ScopeID {
	id := NewFast()
	binary.BigEndian.PutUint32(id[10:14], hash)
	return ScopeID{ID: id, Scope: scope}
}

// HashScope pre-calculates the 32-bit logical group fingerprint.
func HashScope(s string) uint32 {
	return fnv1a(s)
}

// String renders a scope ID using `atom:scope.payload` format.
func (sid ScopeID) String(atom Atom) string {
	if sid.Scope == "" {
		return sid.ID.String(atom)
	}
	buf := make([]byte, 0, len(atom)+len(sid.Scope)+26)
	buf = append(buf, string(atom)...)
	buf = append(buf, delimiterForFormat(DefaultWireFormat()))
	buf = append(buf, sid.Scope...)
	buf = append(buf, '.')
	return bytesToString(sid.ID.appendPayload(buf))
}

// ParseScope parses scope-aware IDs and falls back to plain SNID parsing.
func ParseScope(s string) (ScopeID, Atom, error) {
	delimIdx := strings.IndexAny(s, ":_")
	if delimIdx != -1 && s[delimIdx] == '_' && !AcceptUnderscore() {
		delimIdx = -1
	}
	dotIdx := strings.LastIndex(s, ".")
	if delimIdx == -1 || dotIdx == -1 || delimIdx >= dotIdx {
		var id ID
		atom, err := id.Parse(s)
		return ScopeID{ID: id}, atom, err
	}
	var id ID
	atom := Atom(s[:delimIdx])
	if !IsValidAtom(atom) {
		return ScopeID{}, "", ErrInvalidAtom
	}
	atom = CanonicalAtom(atom)
	countParsedFormat(formatFromDelimiter(s[delimIdx]))
	if err := id.ParseCompact(s[dotIdx+1:]); err != nil {
		return ScopeID{}, "", err
	}
	return ScopeID{ID: id, Scope: s[delimIdx+1 : dotIdx]}, atom, nil
}

// --- ROUTING: SHARD ID ---

type ShardID struct {
	ID       ID
	ShardKey uint16
}

// NewSharded creates an ID bound to a fixed shard key.
func NewSharded(atom Atom, shard uint16) ShardID {
	return ShardID{ID: NewProjected("", shard), ShardKey: shard}
}

// Shard maps the shard key into a bounded shard index.
func (sid ShardID) Shard(total int) int {
	if total <= 0 {
		return 0
	}
	return int(sid.ShardKey) % total
}

// String renders a sharded ID with `#<shard>` suffix.
// Max length: atom(8) + delim(1) + base58+chk(23) + '#'(1) + shard(5) = 38 bytes.
// Stack-allocated [48]byte fits all cases; bytesToString follows the live
// backing array (safe if append never grows past cap, which it won't here).
func (sid ShardID) String(atom Atom) string {
	var buf [48]byte
	dst := sid.ID.AppendTo(buf[:0], atom)
	dst = append(dst, '#')
	dst = strconv.AppendUint(dst, uint64(sid.ShardKey), 10)
	return bytesToString(dst)
}

// ParseSharded parses `atom:payload#shard` formatted identifiers.
func ParseSharded(s string) (ShardID, Atom, error) {
	idx := strings.LastIndex(s, "#")
	if idx == -1 {
		return ShardID{}, "", ErrInvalidFormat
	}
	var id ID
	atom, err := id.Parse(s[:idx])
	if err != nil {
		return ShardID{}, "", err
	}
	sk, err := strconv.ParseUint(s[idx+1:], 10, 16)
	if err != nil {
		return ShardID{}, "", err
	}
	return ShardID{ID: id, ShardKey: uint16(sk)}, atom, nil
}

// --- UX: ALIAS ID ---

type AliasID struct {
	ID    ID
	Alias string
}

// NewWithAlias creates an ID carrying a sanitized human alias.
func NewWithAlias(atom Atom, alias string) AliasID {
	return AliasID{ID: NewFast(), Alias: sanitizeAlias(alias)}
}

// String renders an alias ID using `atom:alias/payload` format.
func (aid AliasID) String(atom Atom) string {
	buf := make([]byte, 0, len(atom)+len(aid.Alias)+26)
	buf = append(buf, string(atom)...)
	buf = append(buf, delimiterForFormat(DefaultWireFormat()))
	buf = append(buf, aid.Alias...)
	buf = append(buf, '/')
	return bytesToString(aid.ID.appendPayload(buf))
}

// ParseAlias parses alias-form IDs and extracts alias metadata.
func ParseAlias(s string) (AliasID, Atom, error) {
	colonIdx := strings.IndexAny(s, ":_")
	if colonIdx != -1 && s[colonIdx] == '_' && !AcceptUnderscore() {
		colonIdx = -1
	}
	slashIdx := strings.LastIndex(s, "/")
	if colonIdx == -1 || slashIdx == -1 || colonIdx >= slashIdx {
		var id ID
		atom, err := id.Parse(s)
		return AliasID{ID: id}, atom, err
	}
	var id ID
	atom := Atom(s[:colonIdx])
	if !IsValidAtom(atom) {
		return AliasID{}, "", ErrInvalidAtom
	}
	atom = CanonicalAtom(atom)
	countParsedFormat(formatFromDelimiter(s[colonIdx]))
	if err := id.ParseCompact(s[slashIdx+1:]); err != nil {
		return AliasID{}, "", err
	}
	return AliasID{ID: id, Alias: s[colonIdx+1 : slashIdx]}, atom, nil
}

// --- UX: SHORT ID ---

type ShortID [8]byte

// NewShort creates a compact 64-bit short identifier.
func NewShort(atom Atom) ShortID {
	var id ShortID

	s := nextShard()

	s.mu.Lock()
	// Xoshiro256** step
	res := rotl(s.s1*5, 7) * 9
	t := s.s1 << 17
	s.s2 ^= s.s0
	s.s3 ^= s.s1
	s.s1 ^= s.s2
	s.s0 ^= s.s3
	s.s2 ^= t
	s.s3 = rotl(s.s3, 45)
	s.mu.Unlock()

	binary.BigEndian.PutUint64(id[:], res)
	return id
}

// String renders a short ID as `atom:<base58>`.
func (sid ShortID) String(atom Atom) string {
	return string(atom) + string(delimiterForFormat(DefaultWireFormat())) + base58.Encode(sid[:])
}

// --- OBSERVABILITY: TRACE ID ---

// TraceID is a 16-byte ID compatible with W3C/OpenTelemetry.
type TraceID ID

// NewTrace creates a new TraceID using the fast path.
func NewTrace() TraceID { return TraceID(NewFast()) }

// hexTable for zero-alloc hex encoding
const hexChars = "0123456789abcdef"
const hexCharsUpper = "0123456789ABCDEF"

// String returns the hex representation of the TraceID.
// OPTIMIZATION: Inline hex encoding to minimize allocations.
// Performance: ~15ns (vs ~35ns for hex.EncodeToString)
func (t TraceID) String() string {
	var buf [32]byte
	for i, b := range t {
		buf[i*2] = hexChars[b>>4]
		buf[i*2+1] = hexChars[b&0x0F]
	}
	// unsafe.String causes buf to escape to heap via escape analysis,
	// avoiding the extra memcopy that string(buf[:]) would perform.
	return unsafe.String(&buf[0], 32)
}

// TraceParent returns a W3C compatible traceparent header.
// Format: 00-<trace_id>-<span_id>-01
//
// OPTIMIZATION: Fixed-size buffer with inline hex encoding.
// Performance: ~18ns (vs ~80ns for string concatenation + hex.EncodeToString)
func (t TraceID) TraceParent(spanID [8]byte) string {
	var buf [55]byte // Fixed: "00-" (3) + trace (32) + "-" (1) + span (16) + "-01" (3)

	// Version prefix
	buf[0], buf[1], buf[2] = '0', '0', '-'

	// TraceID (32 hex chars)
	for i, b := range t {
		buf[3+i*2] = hexChars[b>>4]
		buf[3+i*2+1] = hexChars[b&0x0F]
	}

	// Separator
	buf[35] = '-'

	// SpanID (16 hex chars)
	for i, b := range spanID {
		buf[36+i*2] = hexChars[b>>4]
		buf[36+i*2+1] = hexChars[b&0x0F]
	}

	// Flags
	buf[52], buf[53], buf[54] = '-', '0', '1'

	// unsafe.String causes buf to escape to heap via escape analysis,
	// avoiding the extra memcopy that string(buf[:]) would perform.
	return unsafe.String(&buf[0], 55)
}

// --- INTERNAL HELPERS ---

// signGrant signs grant payload and binds atom to prevent prefix spoofing.
func signGrant(id ID, atom Atom, exp time.Time, secret []byte) [16]byte {
	// For standard grants, we init the cipher on the fly.
	// This costs ~800ns. Use NewGrantKey/NewGrantTurbo for high-performance paths.
	block, err := aes.NewCipher(secret)
	if err != nil {
		// Fallback for invalid key lengths to avoid panic, though this shouldn't happen
		// if input is validated. Returning zero-sig fails verification safely.
		return [16]byte{}
	}

	return signAESCBCWithAtom(block, id, atom, exp)
}

// signGrantLegacy verifies/creates signatures for pre-v8.4 grants.
func signGrantLegacy(id ID, exp time.Time, secret []byte) [16]byte {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return [16]byte{}
	}
	return signAESCBC(block, id, exp)
}

// sanitizeAlias implements the corresponding operation.
func sanitizeAlias(s string) string {
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

// encode16Base58 implements the corresponding operation.
func encode16Base58(src [16]byte) string {
	// Base58 encoding of 16 bytes -> ~22 chars
	return base58.Encode(src[:])
}

// decode16Base58Bytes implements the corresponding operation.
func decode16Base58Bytes(s string) ([16]byte, error) {
	b, err := base58.Decode(s)
	if err != nil || len(b) != 16 {
		return [16]byte{}, ErrInvalidSignature
	}
	var sig [16]byte
	copy(sig[:], b)
	return sig, nil
}

// TestID builds deterministic IDs for tests from timestamp and sequence.
func TestID(atom Atom, ts time.Time, seq uint32) ID {
	return assemble(uint64(ts.UnixMilli()), uint64(seq), 0, 0, 0)
}

// TestIDSequence returns deterministic test IDs with incremental sequence numbers.
func TestIDSequence(atom Atom, ts time.Time, count int) []ID {
	ids := make([]ID, count)
	for i := range count {
		ids[i] = TestID(atom, ts, uint32(i))
	}
	return ids
}

type ValidationOptions struct {
	RequireVersion7 bool
	CheckTimestamp  bool
	MaxAge          time.Duration
}

// ParseWithOptions parses SNID and enforces additional validation constraints.
func ParseWithOptions(s string, opts ValidationOptions) (ID, Atom, error) {
	var id ID
	atom, err := id.Parse(s)
	if err != nil {
		return Zero, "", err
	}
	if opts.RequireVersion7 && id.Version() != 7 {
		return Zero, "", errors.New("snid: not v7")
	}
	if opts.CheckTimestamp && id.Time().After(time.Now().Add(time.Minute)) {
		return Zero, "", errors.New("snid: future ts")
	}
	if opts.MaxAge > 0 && time.Since(id.Time()) > opts.MaxAge {
		return Zero, "", ErrTooOld
	}
	return id, atom, nil
}

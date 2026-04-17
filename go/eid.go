package snid

import (
	"encoding/binary"
	"sync/atomic"
	"time"
	"unsafe"
)

// EID is a compact 64-bit ephemeral ID for ultra-short-lived events.
// Layout: [48-bit millisecond timestamp][16-bit session-scoped counter].
type EID uint64

var eidCounter atomic.Uint32

// NewEphemeral creates an 8-byte ephemeral ID for session-scoped telemetry.
func NewEphemeral(session uint16) EID {
	ms := unixMilliCoarse() & ((1 << 48) - 1)
	ctr := uint16(eidCounter.Add(1))
	low := ctr ^ session
	return EID((ms << 16) | uint64(low))
}

// Bytes returns the big-endian wire representation.
func (e EID) Bytes() [8]byte {
	var out [8]byte
	binary.BigEndian.PutUint64(out[:], uint64(e))
	return out
}

// Time extracts the millisecond timestamp component.
func (e EID) Time() time.Time {
	return time.UnixMilli(int64(uint64(e) >> 16))
}

// Counter returns the low 16-bit counter/session field.
func (e EID) Counter() uint16 {
	return uint16(e)
}

// String renders a debug-only identifier string with zero allocations.
// Matches the %016X format: "TICK_" followed by 16 uppercase hex digits.
func (e EID) String() string {
	var buf [21]byte
	buf[0], buf[1], buf[2], buf[3], buf[4] = 'T', 'I', 'C', 'K', '_'
	v := uint64(e)
	for i := 20; i >= 5; i-- {
		buf[i] = hexCharsUpper[v&0x0F]
		v >>= 4
	}
	return unsafe.String(&buf[0], 21)
}

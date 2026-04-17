package snid

import (
	"os"
	"strings"
	"sync/atomic"
)

// WireFormat controls atom/payload delimiter rendering and parsing.
type WireFormat uint8

const (
	WireColon WireFormat = iota
	WireUnderscore
)

var (
	defaultWireFormat atomic.Uint32
	acceptUnderscore  atomic.Bool

	parseColonCount      atomic.Uint64
	parseUnderscoreCount atomic.Uint64
)

func init() {
	defaultWireFormat.Store(uint32(parseWireFormat(os.Getenv("SNID_WIRE_OUTPUT_FORMAT"))))
	acceptUnderscore.Store(parseBoolDefaultTrue(os.Getenv("SNID_ACCEPT_UNDERSCORE")))
}

func parseWireFormat(v string) WireFormat {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", "colon", ":", "default":
		return WireColon
	case "underscore", "_":
		return WireUnderscore
	default:
		return WireColon
	}
}

func parseBoolDefaultTrue(v string) bool {
	if v == "" {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func delimiterForFormat(f WireFormat) byte {
	if f == WireUnderscore {
		return '_'
	}
	return ':'
}

func formatFromDelimiter(delim byte) WireFormat {
	if delim == '_' {
		return WireUnderscore
	}
	return WireColon
}

// DefaultWireFormat returns the process-wide default formatter mode.
func DefaultWireFormat() WireFormat {
	return WireFormat(defaultWireFormat.Load())
}

// SetDefaultWireFormat sets the process-wide default formatter mode.
func SetDefaultWireFormat(f WireFormat) {
	defaultWireFormat.Store(uint32(f))
}

// AcceptUnderscore reports whether underscore-delimited wire IDs are accepted.
func AcceptUnderscore() bool {
	return acceptUnderscore.Load()
}

// SetAcceptUnderscore toggles underscore-delimited parser acceptance.
func SetAcceptUnderscore(v bool) {
	acceptUnderscore.Store(v)
}

// ParseFormatStats returns parser counters by delimiter format.
func ParseFormatStats() (colon uint64, underscore uint64) {
	return parseColonCount.Load(), parseUnderscoreCount.Load()
}

func countParsedFormat(f WireFormat) {
	if f == WireUnderscore {
		parseUnderscoreCount.Add(1)
		return
	}
	parseColonCount.Add(1)
}

package snid

import (
	"os"
	"strings"
)

type LIDMode string

const (
	LIDModeOff     LIDMode = "off"
	LIDModeShadow  LIDMode = "shadow"
	LIDModeEnforce LIDMode = "enforce"
)

// CurrentLIDMode reads rollout mode from SNID_ENABLE_LID.
func CurrentLIDMode() LIDMode {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("SNID_ENABLE_LID"))) {
	case "enforce":
		return LIDModeEnforce
	case "shadow":
		return LIDModeShadow
	default:
		return LIDModeOff
	}
}

// EIDInternalEnabled reads SNID_ENABLE_EID_INTERNAL rollout switch.
func EIDInternalEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("SNID_ENABLE_EID_INTERNAL"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

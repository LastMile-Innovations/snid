//go:build runtimesecret

package snid

import "runtime/secret"

// doSecret executes f within a runtime/secret boundary, ensuring that
// temporary stack, registers, and heap allocations are erased after use.
func doSecret(f func()) {
	secret.Do(f)
}

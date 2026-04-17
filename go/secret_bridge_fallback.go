//go:build !runtimesecret

package snid

// doSecret directly executes f when runtime/secret is not enabled.
func doSecret(f func()) {
	f()
}

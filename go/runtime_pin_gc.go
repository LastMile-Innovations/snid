//go:build gc

package snid

import _ "unsafe"

const runtimePinEnabled = true

//go:linkname procPin runtime.procPin
func procPin() int

//go:linkname procUnpin runtime.procUnpin
func procUnpin()

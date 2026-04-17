//go:build !gc

package snid

const runtimePinEnabled = false

func procPin() int { return -1 }

func procUnpin() {}

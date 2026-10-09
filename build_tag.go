//go:build tinygo && !kb02_inputonly

package main

// Fail closed: never build upstream keyboard/Vial initialization accidentally.
var _ = buildFirmwareWithKb02InputonlyTag

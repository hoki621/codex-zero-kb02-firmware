//go:build reset_to_default && !kb02_inputonly

package keyboard

import (
	"machine"
)

func init() {
	machine.Flash.EraseBlocks(0, 1)
}

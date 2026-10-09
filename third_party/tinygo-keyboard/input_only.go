//go:build tinygo && kb02_inputonly

package keyboard

// Input-only build: the upstream matrix scanner/debounce is unchanged.
// No USB keyboard, Vial, flash configuration, macros or output dispatcher.
type Device struct{ kb []*MatrixKeyboard }
type Keycode uint16
type State uint8

const (
	None State = iota
	NoneToPress
	Press
	PressToRelease
)

type Callback func(layer, index int, state State)

func New() *Device { return &Device{} }

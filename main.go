//go:build tinygo && kb02_inputonly

package main

import (
	"machine"
	"machine/usb"
	"machine/usb/hid/mouse"
	"time"

	keyboard "github.com/sago35/tinygo-keyboard"
	pio "github.com/tinygo-org/pio/rp2-pio"
	"github.com/tinygo-org/pio/rp2-pio/piolib"
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/encoders"
	"tinygo.org/x/drivers/ssd1306"
)

// Pin mapping follows sago35/keyboards zero-kb02 at 4b18114 (MIT).
// See recovery/sago35-keyboards-LICENSE.txt; no workshop source is copied.
var (
	columnPins = [4]machine.Pin{machine.GPIO5, machine.GPIO6, machine.GPIO7, machine.GPIO8}
	rowPins    = [3]machine.Pin{machine.GPIO9, machine.GPIO10, machine.GPIO11}
)

const (
	scanPeriod           = time.Millisecond
	joystickReportPeriod = 10 * time.Millisecond
	maxSerialReads       = 32
)

func main() {
	usb.Product = "zero-kb02-agent-panel"
	usb.Serial = "zero-kb02-v2"

	matrix := keyboard.New().AddMatrixKeyboard(columnPins[:], rowPins[:], nil, keyboard.InvertDiode(matrixInvertDiode))
	encoder := encoders.NewQuadratureViaInterrupt(machine.GPIO3, machine.GPIO4)
	if err := encoder.Configure(encoders.QuadratureConfig{Precision: encoderPrecision}); err != nil {
		panic(err)
	}
	machine.InitADC()
	joystickX := machine.ADC{Pin: machine.GPIO29}
	joystickY := machine.ADC{Pin: machine.GPIO28}
	joystickX.Configure(machine.ADCConfig{})
	joystickY.Configure(machine.ADCConfig{})
	pointer := mouse.Port()

	display := newOLED(machine.I2C0)
	leds := newLEDStrip()
	protocol := newSession()
	var parser lineParser
	var input keyInput
	var outgoing transmitQueue
	connected := false
	lastValid := time.Now()
	reportRevision := ^uint32(0)
	lastPanel := panelState{selected: -2}
	var lastJoystickReport time.Time
	var lastEncoderPosition int32

	for {
		now := time.Now()
		dtr := machine.Serial.DTR()
		// DTR marks a new Host connection; old input must not cross that boundary.
		if dtr != connected {
			connected = dtr
			parser.reset()
			protocol.resetUSB()
			if connected {
				lastValid = now
			}
		}

		if connected {
			// Bound receive work so a busy Host cannot starve the input scan.
			for reads := 0; reads < maxSerialReads && machine.Serial.Buffered() > 0; reads++ {
				b, err := machine.Serial.ReadByte()
				if err != nil {
					break
				}
				cmd, complete := parser.feed(b)
				if !complete {
					continue
				}
				reply, accepted, _ := protocol.handle(cmd)
				if accepted {
					lastValid = now
				}
				writeCDC(reply)
			}
			if protocol.handshake && heartbeatExpired(now.Sub(lastValid)) {
				protocol.expire()
			}
		}

		states := matrix.Get()
		// Discard queued edges and fractional rotation on every session change.
		if protocol.revision != reportRevision {
			outgoing.reset()
			input.reset()
			encoder.SetPosition(0)
			lastEncoderPosition = 0
			reportRevision = protocol.revision
		}
		for key, state := range states {
			changed := state == keyboard.NoneToPress || state == keyboard.PressToRelease
			pressed := state == keyboard.NoneToPress || state == keyboard.Press
			if !outgoing.push(input.update(key, protocol.generation, protocol.canEmit(), changed, pressed)) {
				protocol.resetUSB()
				outgoing.reset()
			}
		}
		position := int32(encoder.Position())
		delta := int(position-lastEncoderPosition) * encoderDirection
		lastEncoderPosition = position
		if protocol.canEmit() && !outgoing.pushEncoder(protocol.generation, delta) {
			protocol.resetUSB()
			outgoing.reset()
		}

		if connected {
			writeCDC(outgoing.pop())
		}

		if now.Sub(lastJoystickReport) >= joystickReportPeriod {
			lastJoystickReport = now
			rawX, rawY := joystickX.Get(), joystickY.Get()
			if dx, dy := joystickPointerDelta(rawX, rawY, zeroKB02Joystick); dx != 0 || dy != 0 {
				pointer.Move(dx, dy)
			}
		}

		// Commit only successful output so a failed transfer is retried next scan.
		if protocol.panel != lastPanel {
			frame := ledFrame(protocol.panel)
			var ledErr error
			if leds != nil {
				ledErr = leds.WriteRaw(frame[:])
			}
			displayErr := renderPanel(display, protocol.panel)
			if displayErr != nil {
				ledErr = displayErr
			}
			lastPanel = committedPanel(lastPanel, protocol.panel, ledErr)
		}

		time.Sleep(scanPeriod)
	}
}

func newLEDStrip() *piolib.WS2812B {
	stateMachine, err := pio.PIO0.ClaimStateMachine()
	if err != nil {
		return nil
	}
	leds, err := piolib.NewWS2812B(stateMachine, machine.GPIO1)
	if err != nil {
		return nil
	}
	return leds
}

func writeCDC(message string) {
	if message != "" {
		// TinyGo buffers CDC writes; this does not acknowledge delivery to the Host.
		_, _ = machine.Serial.Write([]byte(message))
	}
}

func newOLED(i2c *machine.I2C) *ssd1306.Device {
	if err := i2c.Configure(machine.I2CConfig{Frequency: 400 * machine.KHz, SDA: machine.GPIO12, SCL: machine.GPIO13}); err != nil {
		panic(err)
	}
	device := ssd1306.NewI2C(i2c)
	device.Configure(ssd1306.Config{Width: 128, Height: 64, Address: 0x3c, Rotation: drivers.Rotation180})
	return device
}

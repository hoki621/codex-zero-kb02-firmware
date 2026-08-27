//go:build tinygo

package main

import (
	"machine"
	"machine/usb"
	"machine/usb/hid/mouse"
	"time"

	pio "github.com/tinygo-org/pio/rp2-pio"
	"github.com/tinygo-org/pio/rp2-pio/piolib"
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/ssd1306"
)

var (
	columnPins = [4]machine.Pin{machine.GPIO5, machine.GPIO6, machine.GPIO7, machine.GPIO8}
	rowPins    = [3]machine.Pin{machine.GPIO9, machine.GPIO10, machine.GPIO11}
)

func main() {
	usb.Product = "zero-kb02-agent-panel"
	usb.Serial = "zero-kb02-v1"

	configureInputs()
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
	var keys [12]debouncer
	var keyReported [12]bool
	var escape escapeInput
	var popup popupInput
	var newChat newChatInput
	var encoderButton debouncer
	var encoderButtonReported bool
	var joystickButton debouncer
	var encoder encoderDecoder
	joystick := joystickDecoder{calibration: zeroKB02Joystick}
	connected := false
	lastValid := time.Now()
	reportGeneration := uint64(0)
	lastPanel := panelState{selected: -2}
	lastJoystickReport := time.Time{}

	for {
		now := time.Now()
		dtr := machine.Serial.DTR()
		if dtr && !connected {
			connected = true
			parser.reset()
			protocol.resetUSB()
			lastValid = now
		}
		if !dtr && connected {
			connected = false
			parser.reset()
			protocol.resetUSB()
		}

		if connected {
			for reads := 0; reads < 32 && machine.Serial.Buffered() > 0; reads++ {
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

		if !protocol.canEmit() || protocol.generation != reportGeneration {
			for i := range keyReported {
				keyReported[i] = false
			}
			encoderButtonReported = false
			encoder.reset(machine.GPIO3.Get(), machine.GPIO4.Get())
			reportGeneration = protocol.generation
		}

		rawKeys := scanMatrix()
		for key, raw := range rawKeys {
			changed, pressed := keys[key].update(raw)
			if isEscapeKey(key) {
				writeCDC(escape.update(protocol.generation, protocol.canEmit(), changed, pressed))
				continue
			}
			if isPopupKey(key) {
				writeCDC(popup.update(protocol.generation, protocol.canEmit(), changed, pressed))
				continue
			}
			if isNewChatKey(key) {
				writeCDC(newChat.update(protocol.generation, protocol.canEmit(), changed, pressed))
				continue
			}
			if !changed {
				continue
			}
			if event, approvalKey := approvalEvent(key, protocol.generation, pressed); approvalKey {
				if pressed && protocol.canEmit() {
					writeCDC(event)
					keyReported[key] = true
				} else if !pressed && keyReported[key] {
					writeCDC(event)
					keyReported[key] = false
				}
				continue
			}
			slot, agentKey := slotForKey(key)
			if !agentKey {
				continue
			}
			if pressed && protocol.canEmit() {
				writeCDC(formatKey(protocol.generation, slot, true))
				keyReported[key] = true
			} else if !pressed && keyReported[key] {
				writeCDC(formatKey(protocol.generation, slot, false))
				keyReported[key] = false
			}
		}

		if event := encoder.update(machine.GPIO3.Get(), machine.GPIO4.Get()); event != "" && protocol.canEmit() {
			writeCDC(formatEncoder(protocol.generation, event))
		}
		if changed, pressed := encoderButton.update(!machine.GPIO2.Get()); changed {
			if message, ok := pushEvent(encoderPush, protocol.generation, pressed); pressed && ok && protocol.canEmit() {
				writeCDC(message)
				encoderButtonReported = true
			} else if !pressed && ok && encoderButtonReported {
				writeCDC(message)
				encoderButtonReported = false
			}
		}
		if changed, pressed := joystickButton.update(!machine.GPIO0.Get()); changed {
			_, _ = pushEvent(joystickPush, protocol.generation, pressed)
		}

		if now.Sub(lastJoystickReport) >= 10*time.Millisecond {
			lastJoystickReport = now
			rawX, rawY := joystickX.Get(), joystickY.Get()
			if dx, dy := joystickPointerDelta(rawX, rawY, zeroKB02Joystick); dx != 0 || dy != 0 {
				pointer.Move(dx, dy)
			}
			if direction := joystick.update(rawX, rawY); direction != "" && protocol.canEmit() {
				writeCDC(formatJoystick(protocol.generation, direction))
			}
		}

		if protocol.panel != lastPanel {
			frame := ledFrame(protocol.panel)
			if leds != nil {
				_ = leds.WriteRaw(frame[:])
			}
			lastPanel = committedPanel(lastPanel, protocol.panel, display.render(protocol.panel))
		}

		time.Sleep(time.Millisecond)
	}
}

func configureInputs() {
	for _, pin := range columnPins {
		pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	}
	for _, pin := range rowPins {
		pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	}
	for _, pin := range []machine.Pin{machine.GPIO0, machine.GPIO2, machine.GPIO3, machine.GPIO4} {
		pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}
}

func scanMatrix() [12]bool {
	var pressed [12]bool
	for column, pin := range columnPins {
		pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		pin.High()
		for row, rowPin := range rowPins {
			pressed[row*len(columnPins)+column] = rowPin.Get()
		}
		pin.Low()
		pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	}
	return pressed
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
		_, _ = machine.Serial.Write([]byte(message))
	}
}

type oled struct {
	display *ssd1306.Device
	buffer  [1024]byte
}

func newOLED(i2c *machine.I2C) *oled {
	_ = i2c.Configure(machine.I2CConfig{
		Frequency: 400 * machine.KHz,
		SDA:       machine.GPIO12,
		SCL:       machine.GPIO13,
	})
	device := ssd1306.NewI2C(i2c)
	device.Configure(ssd1306.Config{
		Width:    128,
		Height:   64,
		Address:  0x3c,
		Rotation: drivers.Rotation180,
	})
	return &oled{display: device}
}

func (d *oled) render(panel panelState) error {
	clear(d.buffer[:])
	if !panel.online {
		for i := 0; i < 64; i++ {
			d.setPixel(i+32, i)
			d.setPixel(95-i, i)
		}
	} else {
		for slot := range slotCount {
			x := (slot % 2) * 64
			y := (slot / 2) * 21
			if panel.selected == int8(slot) {
				d.rectangle(x, y, 64, 21)
			} else {
				d.setPixel(x+2, y+2)
				d.setPixel(x+61, y+2)
			}
			d.drawGlyph(x+27, y+3, panel.states[slot])
		}
	}
	if err := d.display.SetBuffer(d.buffer[:]); err != nil {
		return err
	}
	return d.display.Display()
}

func (d *oled) setPixel(x, y int) {
	if x < 0 || x >= 128 || y < 0 || y >= 64 {
		return
	}
	d.buffer[x+(y/8)*128] |= 1 << uint(y%8)
}

func (d *oled) rectangle(x, y, width, height int) {
	for offset := 0; offset < width; offset++ {
		d.setPixel(x+offset, y)
		d.setPixel(x+offset, y+height-1)
	}
	for offset := 0; offset < height; offset++ {
		d.setPixel(x, y+offset)
		d.setPixel(x+width-1, y+offset)
	}
}

func (d *oled) drawGlyph(x, y int, character byte) {
	rows := glyph(character)
	for row, bits := range rows {
		for column := 0; column < 5; column++ {
			if bits&(1<<uint(4-column)) == 0 {
				continue
			}
			for scaleY := 0; scaleY < 2; scaleY++ {
				for scaleX := 0; scaleX < 2; scaleX++ {
					d.setPixel(x+column*2+scaleX, y+row*2+scaleY)
				}
			}
		}
	}
}

func glyph(character byte) [7]byte {
	switch character {
	case 'W':
		return [7]byte{0x11, 0x11, 0x11, 0x15, 0x15, 0x15, 0x0a}
	case 'I':
		return [7]byte{0x1f, 0x04, 0x04, 0x04, 0x04, 0x04, 0x1f}
	case 'B':
		return [7]byte{0x1e, 0x11, 0x11, 0x1e, 0x11, 0x11, 0x1e}
	case 'D':
		return [7]byte{0x1e, 0x11, 0x11, 0x11, 0x11, 0x11, 0x1e}
	case 'U':
		return [7]byte{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x0e}
	case 'E':
		return [7]byte{0x1f, 0x10, 0x10, 0x1e, 0x10, 0x10, 0x1f}
	default:
		return [7]byte{}
	}
}

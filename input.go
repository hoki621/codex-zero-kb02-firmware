package main

const (
	debounceSamples     = 5
	joystickPointerStep = 3
)

type debouncer struct {
	stable    bool
	candidate bool
	count     uint8
}

func (d *debouncer) update(raw bool) (changed, pressed bool) {
	if raw == d.stable {
		d.candidate = raw
		d.count = 0
		return false, d.stable
	}
	if raw != d.candidate {
		d.candidate = raw
		d.count = 1
	} else if d.count < debounceSamples {
		d.count++
	}
	if d.count < debounceSamples {
		return false, d.stable
	}
	d.stable = raw
	d.count = 0
	return true, d.stable
}

func slotForKey(key int) (int, bool) {
	switch key {
	case 1:
		return 0, true
	case 2:
		return 1, true
	case 4:
		return 2, true
	case 5:
		return 3, true
	case 6:
		return 4, true
	case 7:
		return 5, true
	default:
		return 0, false
	}
}

func isEscapeKey(key int) bool {
	return key == 0
}

func isPopupKey(key int) bool {
	return key == 3
}

type escapeInput struct {
	generation uint64
	reported   bool
}

func (e *escapeInput) update(generation uint64, online, changed, pressed bool) string {
	if !online || generation != e.generation {
		e.generation = generation
		e.reported = false
	}
	if !online || !changed {
		return ""
	}
	if pressed {
		if e.reported {
			return ""
		}
		e.reported = true
		return formatEscape(generation, true)
	}
	if !e.reported {
		return ""
	}
	e.reported = false
	return formatEscape(generation, false)
}

type popupInput struct {
	generation uint64
	reported   bool
}

func (p *popupInput) update(generation uint64, online, changed, pressed bool) string {
	if !online || generation != p.generation {
		p.generation = generation
		p.reported = false
	}
	if !online || !changed {
		return ""
	}
	if pressed {
		if p.reported {
			return ""
		}
		p.reported = true
		return formatPopup(generation, true)
	}
	if !p.reported {
		return ""
	}
	p.reported = false
	return formatPopup(generation, false)
}

type pushInput uint8

const (
	encoderPush pushInput = iota
	joystickPush
)

func pushEvent(input pushInput, generation uint64, pressed bool) (string, bool) {
	if input != encoderPush {
		return "", false
	}
	edge := "UP"
	if pressed {
		edge = "DOWN"
	}
	return formatEncoder(generation, edge), true
}

type encoderDecoder struct {
	initialized bool
	state       uint8
	steps       int8
}

var encoderTransitions = [16]int8{0, -1, 1, 0, 1, 0, 0, -1, -1, 0, 0, 1, 0, 1, -1, 0}

func (d *encoderDecoder) reset(a, b bool) {
	d.initialized = true
	d.state = encoderState(a, b)
	d.steps = 0
}

func (d *encoderDecoder) update(a, b bool) string {
	next := encoderState(a, b)
	if !d.initialized {
		d.initialized = true
		d.state = next
		return ""
	}
	previous := d.state
	d.state = next
	delta := encoderTransitions[(previous<<2)|next]
	if delta == 0 && previous != next {
		d.steps = 0
		return ""
	}
	d.steps += delta
	if d.steps >= 4 {
		d.steps = 0
		return "CW"
	}
	if d.steps <= -4 {
		d.steps = 0
		return "CCW"
	}
	return ""
}

func encoderState(a, b bool) uint8 {
	state := uint8(0)
	if a {
		state |= 2
	}
	if b {
		state |= 1
	}
	return state
}

type joystickCalibration struct {
	centerX, centerY uint16
	enter, release   uint16
	invertX, invertY bool
}

var zeroKB02Joystick = joystickCalibration{
	centerX: 31584,
	centerY: 32496,
	enter:   8192,
	release: 4096,
	invertX: false,
	invertY: true,
}

type joystickDecoder struct {
	calibration joystickCalibration
	active      bool
}

func (d *joystickDecoder) update(rawX, rawY uint16) string {
	x, y := joystickAxes(rawX, rawY, d.calibration)
	absX, absY := absolute(x), absolute(y)
	if d.active {
		if absX <= int32(d.calibration.release) && absY <= int32(d.calibration.release) {
			d.active = false
		}
		return ""
	}
	direction := joystickDirection(x, y, d.calibration.enter)
	if direction == "" {
		return ""
	}
	d.active = true
	return direction
}

func joystickPointerDelta(rawX, rawY uint16, calibration joystickCalibration) (int, int) {
	x, y := joystickAxes(rawX, rawY, calibration)
	switch joystickDirection(x, y, calibration.enter) {
	case "LEFT":
		return -joystickPointerStep, 0
	case "RIGHT":
		return joystickPointerStep, 0
	case "UP":
		return 0, -joystickPointerStep
	case "DOWN":
		return 0, joystickPointerStep
	default:
		return 0, 0
	}
}

func joystickAxes(rawX, rawY uint16, calibration joystickCalibration) (int32, int32) {
	x := int32(rawX) - int32(calibration.centerX)
	y := int32(rawY) - int32(calibration.centerY)
	if calibration.invertX {
		x = -x
	}
	if calibration.invertY {
		y = -y
	}
	return x, y
}

func joystickDirection(x, y int32, deadZone uint16) string {
	absX, absY := absolute(x), absolute(y)
	if absX < int32(deadZone) && absY < int32(deadZone) {
		return ""
	}
	if absX >= absY {
		if x < 0 {
			return "LEFT"
		}
		return "RIGHT"
	}
	if y < 0 {
		return "UP"
	}
	return "DOWN"
}

func absolute(value int32) int32 {
	if value < 0 {
		return -value
	}
	return value
}

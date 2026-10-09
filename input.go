package main

const (
	joystickPointerStep = 3
	// Upstream matrix debounce is 8: nine None samples exclude a latent press.
	matrixSettleSamples = 9
	matrixInvertDiode   = false
	encoderPrecision    = 4
	encoderDirection    = 1 // Change to -1 only after a physical direction check.
)

type keyInput struct {
	blocked      [12]bool
	reported     [12]bool
	quietSamples [12]uint8
}

func (k *keyInput) reset() {
	*k = keyInput{}
	for i := range k.blocked {
		k.blocked[i] = true
	}
}

func (k *keyInput) update(key int, generation uint64, online, changed, pressed bool) string {
	if !online {
		k.blocked[key] = true
		k.reported[key] = false
		k.quietSamples[key] = 0
		return ""
	}
	if k.blocked[key] {
		if changed && !pressed {
			k.blocked[key] = false
			return ""
		}
		if pressed {
			k.quietSamples[key] = 0
		} else if k.quietSamples[key] < matrixSettleSamples {
			k.quietSamples[key]++
		}
		if k.quietSamples[key] == matrixSettleSamples {
			k.blocked[key] = false
		}
		return ""
	}
	if !changed {
		return ""
	}
	if pressed {
		if k.reported[key] {
			return ""
		}
		k.reported[key] = true
	} else {
		if !k.reported[key] {
			return ""
		}
		k.reported[key] = false
	}
	return formatKey(generation, key+1, pressed)
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

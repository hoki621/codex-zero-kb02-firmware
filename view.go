package main

const maxLEDChannel = 16

type rgb struct {
	r, g, b uint8
}

func committedPanel(last, next panelState, err error) panelState {
	if err != nil {
		return last
	}
	return next
}

var slotLEDIndices = [slotCount]int{3, 6, 1, 4, 7, 10}

func ledFrame(panel panelState) [12]uint32 {
	var frame [12]uint32
	for slot := range slotCount {
		color := rgb{r: 4}
		if panel.online {
			color = stateColor(panel.states[slot])
			if panel.selected == int8(slot) {
				color = brighten(color, 4)
			}
		}
		raw := uint32(color.g)<<24 | uint32(color.r)<<16 | uint32(color.b)<<8
		frame[slotLEDIndices[slot]] = raw
	}
	return frame
}

func stateColor(state byte) rgb {
	switch state {
	case 'W':
		return rgb{b: 12}
	case 'I':
		return rgb{g: 10, b: 6}
	case 'B':
		return rgb{r: 12, g: 4}
	case 'D':
		return rgb{g: 12}
	case 'U':
		return rgb{r: 8, b: 8}
	default:
		return rgb{}
	}
}

func brighten(color rgb, amount uint8) rgb {
	color.r = clampLED(color.r + amount)
	color.g = clampLED(color.g + amount)
	color.b = clampLED(color.b + amount)
	return color
}

func clampLED(value uint8) uint8 {
	if value > maxLEDChannel {
		return maxLEDChannel
	}
	return value
}

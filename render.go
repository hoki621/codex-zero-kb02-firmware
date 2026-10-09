package main

import (
	"image/color"

	"tinygo.org/x/drivers"
	"tinygo.org/x/tinydraw"
	"tinygo.org/x/tinyfont"
)

func renderPanel(display drivers.Displayer, panel panelState) error {
	white := color.RGBA{255, 255, 255, 255}
	if err := tinydraw.FilledRectangle(display, 0, 0, 128, 64, color.RGBA{}); err != nil {
		return err
	}
	if !panel.online {
		tinydraw.Line(display, 32, 0, 95, 63, white)
		tinydraw.Line(display, 32, 63, 95, 0, white)
	} else {
		for slot := range slotCount {
			x, y := int16(slot%2)*64, int16(slot/2)*21
			if panel.selected == int8(slot) {
				if err := tinydraw.Rectangle(display, x, y, 64, 21, white); err != nil {
					return err
				}
			} else {
				display.SetPixel(x+2, y+2, white)
				display.SetPixel(x+61, y+2, white)
			}
			tinyfont.WriteLine(display, &tinyfont.TomThumb, x+30, y+13, string(panel.states[slot]), white)
		}
	}
	return display.Display()
}

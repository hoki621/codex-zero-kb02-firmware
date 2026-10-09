package main

import (
	"errors"
	"image/color"
	"testing"
)

type testDisplay struct {
	pixels      [64][128]bool
	outOfBounds bool
	err         error
}

func (d *testDisplay) Size() (int16, int16) { return 128, 64 }
func (d *testDisplay) Display() error       { return d.err }
func (d *testDisplay) SetPixel(x, y int16, c color.RGBA) {
	if x < 0 || x >= 128 || y < 0 || y >= 64 {
		d.outOfBounds = true
		return
	}
	d.pixels[y][x] = c.R != 0 || c.G != 0 || c.B != 0
}

func TestLibraryLayoutSixStatesSelectionAndOffline(t *testing.T) {
	var d testDisplay
	p := panelState{online: true, selected: 0, states: [6]byte{'W', 'I', 'B', 'D', 'U', 'E'}}
	for slot := range slotCount {
		p.selected = int8(slot)
		if err := renderPanel(&d, p); err != nil {
			t.Fatal(err)
		}
		for cell := range slotCount {
			x, y := cell%2*64, cell/2*21
			if d.pixels[y][x] != (cell == slot) {
				t.Fatal("selection or stale border", slot, cell)
			}
			lit := 0
			for row := y + 5; row < y+15; row++ {
				for col := x + 28; col < x+36; col++ {
					if d.pixels[row][col] {
						lit++
					}
				}
			}
			if lit == 0 {
				t.Fatal("missing state glyph", cell)
			}
		}
	}
	if err := renderPanel(&d, offlinePanel()); err != nil {
		t.Fatal(err)
	}
	if !d.pixels[0][32] || !d.pixels[63][95] || d.pixels[42][64] || d.outOfBounds {
		t.Fatal("offline/bounds")
	}
	d.err = errors.New("I2C failure")
	if renderPanel(&d, p) != d.err {
		t.Fatal("lost transfer error")
	}
}

func TestEncoderQueueBoundsStepsAcrossLibrarySamples(t *testing.T) {
	var q transmitQueue
	if !q.pushEncoder(1, 20) || q.pushEncoder(1, 13) {
		t.Fatal("pending delta overflow")
	}
	if q.pop() != "ENC 1 20\n" || !q.pushEncoder(1, -32) {
		t.Fatal("draining did not release capacity")
	}
	q.reset()
	if !q.pushEncoder(2, 32) || q.pushEncoder(2, 1) || q.pushEncoder(2, -33) {
		t.Fatal("encoder bound")
	}
}

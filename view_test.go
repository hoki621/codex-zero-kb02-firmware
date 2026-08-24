package main

import "testing"

func TestLEDFrameUsesSixAgentKeyLEDsAtLowBrightness(t *testing.T) {
	panel := panelState{online: true, selected: 2, states: [6]byte{'W', 'I', 'B', 'D', 'U', 'E'}}
	frame := ledFrame(panel)
	active := map[int]bool{}
	for _, index := range slotLEDIndices {
		active[index] = true
	}
	for index, raw := range frame {
		if !active[index] && raw != 0 {
			t.Fatalf("reserved LED %d is on", index)
		}
		channels := []uint8{uint8(raw >> 24), uint8(raw >> 16), uint8(raw >> 8)}
		for _, channel := range channels {
			if channel > maxLEDChannel {
				t.Fatalf("LED %d channel %d exceeds limit", index, channel)
			}
		}
	}
}

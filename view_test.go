package main

import (
	"errors"
	"testing"
)

func TestCommittedPanelRetriesDisplayFailure(t *testing.T) {
	last := panelState{selected: -1}
	next := panelState{online: true, selected: 2}
	if got := committedPanel(last, next, errors.New("display failed")); got != last {
		t.Fatalf("failed display advanced panel: %+v", got)
	}
	if got := committedPanel(last, next, nil); got != next {
		t.Fatalf("successful display did not advance panel: %+v", got)
	}
}

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

func TestOfflineLEDsAreOff(t *testing.T) {
	if got := ledFrame(offlinePanel()); got != [12]uint32{} {
		t.Fatal(got)
	}
}

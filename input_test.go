package main

import "testing"

func TestDebouncerEmitsOneEdgePerPress(t *testing.T) {
	var d debouncer
	sequence := []bool{false, true, false, true, true, true, true, true, true, false, true, false, false, false, false, false}
	var edges []bool
	for _, raw := range sequence {
		if changed, pressed := d.update(raw); changed {
			edges = append(edges, pressed)
		}
	}
	if len(edges) != 2 || !edges[0] || edges[1] {
		t.Fatalf("edges=%v", edges)
	}
}

func TestEncoderProducesOneEventPerDetent(t *testing.T) {
	var counterclockwise encoderDecoder
	var events []string
	// Physical CCW on this board, recorded as GPIO3 then GPIO4.
	for _, pins := range [][2]bool{{false, false}, {true, false}, {true, true}, {false, true}, {false, false}} {
		if event := counterclockwise.update(pins[0], pins[1]); event != "" {
			events = append(events, event)
		}
	}
	if len(events) != 1 || events[0] != "CCW" {
		t.Fatalf("counterclockwise events=%v", events)
	}

	var clockwise encoderDecoder
	events = nil
	// Physical CW on this board, recorded as GPIO3 then GPIO4.
	for _, pins := range [][2]bool{{false, false}, {false, true}, {true, true}, {true, false}, {false, false}} {
		if event := clockwise.update(pins[0], pins[1]); event != "" {
			events = append(events, event)
		}
	}
	if len(events) != 1 || events[0] != "CW" {
		t.Fatalf("clockwise events=%v", events)
	}
}

func TestEncoderResetDropsPartialOfflineDetent(t *testing.T) {
	var decoder encoderDecoder
	for _, pins := range [][2]bool{{false, false}, {true, false}, {true, true}, {false, true}} {
		if event := decoder.update(pins[0], pins[1]); event != "" {
			t.Fatalf("partial detent event=%q", event)
		}
	}

	decoder.reset(false, true)
	if event := decoder.update(false, false); event != "" {
		t.Fatalf("pre-session edge replayed as %q", event)
	}

	events := 0
	for _, pins := range [][2]bool{{true, false}, {true, true}, {false, true}, {false, false}} {
		if event := decoder.update(pins[0], pins[1]); event != "" {
			events++
			if event != "CCW" {
				t.Fatalf("fresh detent event=%q", event)
			}
		}
	}
	if events != 1 {
		t.Fatalf("fresh detent events=%d", events)
	}
}

func TestJoystickRequiresNeutralBeforeRepeating(t *testing.T) {
	decoder := joystickDecoder{calibration: zeroKB02Joystick}
	for range 100 {
		if event := decoder.update(31584, 32496); event != "" {
			t.Fatalf("idle event=%q", event)
		}
	}
	if event := decoder.update(55900, 34600); event != "RIGHT" {
		t.Fatalf("right event=%q", event)
	}
	for range 20 {
		if event := decoder.update(55900, 34600); event != "" {
			t.Fatalf("repeated event=%q", event)
		}
	}
	decoder.update(31584, 32496)
	if event := decoder.update(26700, 57900); event != "UP" {
		t.Fatalf("up event=%q", event)
	}
}

func TestOnlySixAgentKeysProduceEvents(t *testing.T) {
	want := map[int]int{1: 0, 2: 1, 4: 2, 5: 3, 6: 4, 7: 5}
	for key := range 12 {
		slot, ok := slotForKey(key)
		wantSlot, agentKey := want[key]
		if ok != agentKey || (ok && slot != wantSlot) {
			t.Fatalf("key=%d slot=%d ok=%v", key+1, slot, ok)
		}
		if ok {
			if down := formatKey(9, slot, true); down != "KEY 9 "+string(rune('0'+slot))+" DOWN\n" {
				t.Fatalf("key=%d down=%q", key+1, down)
			}
			if up := formatKey(9, slot, false); up != "KEY 9 "+string(rune('0'+slot))+" UP\n" {
				t.Fatalf("key=%d up=%q", key+1, up)
			}
		}
	}
}

func TestJoystickPushIsDebouncedAndSafelyIgnored(t *testing.T) {
	var button debouncer
	changes := 0
	for _, raw := range []bool{true, true, true, true, true, false, false, false, false, false} {
		if changed, pressed := button.update(raw); changed {
			changes++
			if message, ok := pushEvent(joystickPush, 9, pressed); ok || message != "" {
				t.Fatalf("joystick push emitted %q", message)
			}
		}
	}
	if changes != 2 {
		t.Fatalf("changes=%d", changes)
	}
}

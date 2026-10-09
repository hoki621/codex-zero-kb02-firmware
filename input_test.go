package main

import "testing"

func TestJoystickPointerDeltaUsesCalibratedDirections(t *testing.T) {
	tests := []struct {
		name       string
		rawX, rawY uint16
		dx, dy     int
	}{
		{name: "center", rawX: 31584, rawY: 32496},
		{name: "inside dead zone", rawX: 39775, rawY: 32496},
		{name: "right", rawX: 55900, rawY: 34600, dx: joystickPointerStep},
		{name: "left", rawX: 8000, rawY: 31400, dx: -joystickPointerStep},
		{name: "up", rawX: 26700, rawY: 57900, dy: -joystickPointerStep},
		{name: "down", rawX: 32900, rawY: 8250, dy: joystickPointerStep},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dx, dy := joystickPointerDelta(test.rawX, test.rawY, zeroKB02Joystick)
			if dx != test.dx || dy != test.dy {
				t.Fatalf("delta=(%d,%d), want=(%d,%d)", dx, dy, test.dx, test.dy)
			}
		})
	}
}

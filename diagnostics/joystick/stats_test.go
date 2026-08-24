package main

import "testing"

func TestStats(t *testing.T) {
	var got stats
	got.add(100, 500)
	got.add(300, 100)

	if got.count != 2 || got.minX != 100 || got.maxX != 300 || got.meanX() != 200 ||
		got.minY != 100 || got.maxY != 500 || got.meanY() != 300 {
		t.Fatalf("unexpected stats: %+v", got)
	}
}

func TestWaitForDTR(t *testing.T) {
	states := []bool{false, false, true}
	checks := 0
	pauses := 0
	waitForDTR(func() bool {
		state := states[checks]
		checks++
		return state
	}, func() {
		pauses++
	})

	if checks != 3 || pauses != 2 {
		t.Fatalf("checks=%d pauses=%d", checks, pauses)
	}
}

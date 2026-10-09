package main

import (
	"strconv"
	"strings"
	"testing"
)

func TestMajor2HeldKeysRequireReleaseAcrossContexts(t *testing.T) {
	for key := range 12 {
		var input keyInput
		input.reset()
		// Eight undecided upstream samples can hide a held press; ninth becomes DOWN.
		for range matrixSettleSamples - 1 {
			if event := input.update(key, 7, true, false, false); event != "" {
				t.Fatal(event)
			}
		}
		if event := input.update(key, 7, true, true, true); event != "" {
			t.Fatal(event)
		}
		for range 20 {
			if event := input.update(key, 7, true, false, true); event != "" {
				t.Fatal(event)
			}
		}
		// Release after reset does not send an orphan UP.
		if event := input.update(key, 7, true, true, false); event != "" {
			t.Fatal(event)
		}
		want := "KEY 7 " + strconv.Itoa(key+1)
		if event := input.update(key, 7, true, true, true); event != want+" DOWN\n" {
			t.Fatal(event)
		}
		if event := input.update(key, 7, true, false, true); event != "" {
			t.Fatal(event)
		}
		if event := input.update(key, 7, true, true, false); event != want+" UP\n" {
			t.Fatal(event)
		}
		// New generation or offline recovery again requires a release.
		input.reset()
		if event := input.update(key, 8, true, true, true); event != "" {
			t.Fatal(event)
		}
		input.update(key, 8, false, false, true)
		if event := input.update(key, 9, true, false, true); event != "" {
			t.Fatal(event)
		}
	}
}

func TestMajor2UnheldKeysBecomeReadyAfterLibrarySettles(t *testing.T) {
	var input keyInput
	input.reset()
	for key := range 12 {
		for range matrixSettleSamples {
			input.update(key, 1, true, false, false)
		}
		if event := input.update(key, 1, true, true, true); event != "KEY 1 "+strconv.Itoa(key+1)+" DOWN\n" {
			t.Fatal(event)
		}
	}
}

func TestMajor2StateVectorsAndContextChanges(t *testing.T) {
	for _, line := range []string{"STATE 0 0 WIBDUE", "STATE 7 5 WIBDUE", "STATE 7 - WIBDUX", "STATE 18446744073709551616 - IIIIII", "OFFLINE 07", "PING -1", "PING 4294967296"} {
		if _, ok := parseCommand(line); ok {
			t.Fatalf("accepted %q", line)
		}
	}
	s := newSession()
	apply := func(line string) {
		t.Helper()
		cmd, ok := parseCommand(line)
		if !ok {
			t.Fatal(line)
		}
		if _, accepted, _ := s.handle(cmd); !accepted {
			t.Fatal(line)
		}
	}
	apply("HELLO HOST 2")
	if s.canEmit() {
		t.Fatal("HELLO enabled input")
	}
	apply("STATE 18446744073709551615 - IIIIII")
	revision := s.revision
	apply("STATE 18446744073709551615 2 WIBDUE")
	if s.revision != revision {
		t.Fatal("display-only change reset input")
	}
	apply("OFFLINE 9")
	if s.canEmit() || s.generation != 9 || s.revision == revision {
		t.Fatal("offline context not invalidated")
	}
	apply("STATE 9 - EEEEEE")
	if !s.canEmit() {
		t.Fatal("STATE failed")
	}
	s.expire()
	cmd, _ := parseCommand("STATE 10 - IIIIII")
	if _, accepted, _ := s.handle(cmd); accepted {
		t.Fatal("timeout did not require HELLO")
	}
}

func TestMajor2QueueIsBoundedAndDiscarded(t *testing.T) {
	var q transmitQueue
	for i := range 32 {
		if !q.push(formatKey(1, i%12+1, true)) {
			t.Fatal(i)
		}
	}
	if q.push("ENC 1 1\n") {
		t.Fatal("unbounded queue")
	}
	q.reset()
	if q.pop() != "" {
		t.Fatal("stale queue")
	}
	for _, delta := range []int{-32, -1, 1, 32} {
		if !q.push(formatEncoder(2, delta)) {
			t.Fatal(delta)
		}
	}
	for _, delta := range []int{-33, 0, 33} {
		if formatEncoder(2, delta) != "" {
			t.Fatal(delta)
		}
	}
	for _, delta := range []int{-32, -1, 1, 32} {
		if q.pop() != "ENC 2 "+strconv.Itoa(delta)+"\n" {
			t.Fatal("delta/order")
		}
	}
}

func TestMajor2LineLengthIncludesCRLFAndRecovers(t *testing.T) {
	var p lineParser
	valid := "PING 0\r\n"
	input := strings.Repeat("X", 127) + "\r\n" + valid + "PING 1\x00\n" + "PING 4294967295\n"
	var sequences []uint32
	for i := range len(input) {
		if cmd, ok := p.feed(input[i]); ok {
			sequences = append(sequences, cmd.sequence)
		}
	}
	if len(sequences) != 2 || sequences[0] != 0 || sequences[1] != 4294967295 {
		t.Fatal(sequences)
	}
}

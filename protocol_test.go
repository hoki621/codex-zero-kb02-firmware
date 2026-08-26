package main

import (
	"strings"
	"testing"
	"time"
)

func TestLineParserRecoversAfterOversizedAndMalformedLines(t *testing.T) {
	var parser lineParser
	input := strings.Repeat("X", 128) + "\nSTATE 1 0 WIBEU\nSTATE 2 0 WIBDUE\r\n"
	var commands []command
	for i := range len(input) {
		if cmd, ok := parser.feed(input[i]); ok {
			commands = append(commands, cmd)
		}
	}
	if len(commands) != 1 || commands[0].generation != 2 || commands[0].states != [6]byte{'W', 'I', 'B', 'D', 'U', 'E'} {
		t.Fatalf("unexpected commands: %+v", commands)
	}
}

func TestHeartbeatExpiresAtTwelveSeconds(t *testing.T) {
	if heartbeatExpired(12*time.Second - time.Nanosecond) {
		t.Fatal("expired before 12 seconds")
	}
	if !heartbeatExpired(12 * time.Second) {
		t.Fatal("did not expire at 12 seconds")
	}
}

func TestProtocolHandshakeStatePingAndTimeout(t *testing.T) {
	state := newSession()

	mismatch, _ := parseCommand("HELLO HOST 2")
	reply, accepted, _ := state.handle(mismatch)
	if !accepted || reply != "HELLO ZERO-KB02 1\n" || state.canEmit() {
		t.Fatalf("major mismatch state: reply=%q state=%+v", reply, state)
	}

	hello, _ := parseCommand("HELLO HOST 1")
	state.handle(hello)
	update, _ := parseCommand("STATE 41827 2 WIBDUE")
	_, accepted, changed := state.handle(update)
	if !accepted || !changed || !state.canEmit() || state.generation != 41827 || state.panel.selected != 2 {
		t.Fatalf("online state: %+v", state)
	}
	if _, _, changed = state.handle(update); changed {
		t.Fatal("identical STATE changed the panel")
	}

	ping, _ := parseCommand("PING 4294967295")
	reply, accepted, _ = state.handle(ping)
	if !accepted || reply != "PONG 4294967295\n" {
		t.Fatalf("ping reply=%q accepted=%v", reply, accepted)
	}
	if !state.expire() || state.canEmit() || state.panel.online {
		t.Fatalf("timeout state: %+v", state)
	}
}

func TestEscapeDoesNotCrossOfflineOrGenerationReset(t *testing.T) {
	state := newSession()
	var escape escapeInput
	if event := escape.update(state.generation, state.canEmit(), true, true); event != "" {
		t.Fatalf("pre-session event=%q", event)
	}

	hello, _ := parseCommand("HELLO HOST 1")
	state.handle(hello)
	online, _ := parseCommand("STATE 9 0 IIIIII")
	state.handle(online)
	if event := escape.update(state.generation, state.canEmit(), true, true); event != "ESC 9 DOWN\n" {
		t.Fatalf("online event=%q", event)
	}

	offline, _ := parseCommand("OFFLINE 10")
	state.handle(offline)
	escape.update(state.generation, state.canEmit(), false, true)
	state.handle(hello)
	next, _ := parseCommand("STATE 11 0 IIIIII")
	state.handle(next)
	if event := escape.update(state.generation, state.canEmit(), true, false); event != "" {
		t.Fatalf("release replayed after reconnect as %q", event)
	}
	if event := escape.update(state.generation, state.canEmit(), true, true); event != "ESC 11 DOWN\n" {
		t.Fatalf("new generation event=%q", event)
	}
	remapped, _ := parseCommand("STATE 12 0 IIIIII")
	state.handle(remapped)
	escape.update(state.generation, state.canEmit(), false, true)
	if event := escape.update(state.generation, state.canEmit(), true, false); event != "" {
		t.Fatalf("release replayed after generation change as %q", event)
	}
}

func TestPopupDoesNotCrossOfflineOrGenerationReset(t *testing.T) {
	state := newSession()
	var popup popupInput
	if event := popup.update(state.generation, state.canEmit(), true, true); event != "" {
		t.Fatalf("pre-session event=%q", event)
	}

	hello, _ := parseCommand("HELLO HOST 1")
	state.handle(hello)
	online, _ := parseCommand("STATE 9 0 IIIIII")
	state.handle(online)
	if event := popup.update(state.generation, state.canEmit(), true, true); event != "POPUP 9 DOWN\n" {
		t.Fatalf("online event=%q", event)
	}

	offline, _ := parseCommand("OFFLINE 10")
	state.handle(offline)
	popup.update(state.generation, state.canEmit(), false, true)
	state.handle(hello)
	next, _ := parseCommand("STATE 11 0 IIIIII")
	state.handle(next)
	if event := popup.update(state.generation, state.canEmit(), true, false); event != "" {
		t.Fatalf("release replayed after reconnect as %q", event)
	}
	if event := popup.update(state.generation, state.canEmit(), true, true); event != "POPUP 11 DOWN\n" {
		t.Fatalf("new generation event=%q", event)
	}
	remapped, _ := parseCommand("STATE 12 0 IIIIII")
	state.handle(remapped)
	popup.update(state.generation, state.canEmit(), false, true)
	if event := popup.update(state.generation, state.canEmit(), true, false); event != "" {
		t.Fatalf("release replayed after generation change as %q", event)
	}
}

func TestMalformedCommandsAreRejected(t *testing.T) {
	for _, line := range []string{
		" STATE 1 0 WIBDUE",
		"STATE +1 0 WIBDUE",
		"STATE 01 0 WIBDUE",
		"STATE 1 6 WIBDUE",
		"STATE 1 0 WIBEU",
		"PING 4294967296",
		"RUN herdr focus 2",
	} {
		if _, ok := parseCommand(line); ok {
			t.Errorf("accepted malformed line %q", line)
		}
	}
}

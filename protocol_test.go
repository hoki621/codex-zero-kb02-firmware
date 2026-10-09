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

	mismatch, _ := parseCommand("HELLO HOST 1")
	reply, accepted, _ := state.handle(mismatch)
	if accepted || reply != "HELLO ZERO-KB02 2\n" || state.canEmit() {
		t.Fatalf("major mismatch state: reply=%q state=%+v", reply, state)
	}

	hello, _ := parseCommand("HELLO HOST 2")
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

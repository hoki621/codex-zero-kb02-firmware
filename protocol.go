package main

import (
	"strconv"
	"strings"
	"time"
)

const (
	maxLineBytes     = 128
	slotCount        = 6
	heartbeatTimeout = 12 * time.Second
)

type commandKind uint8

const (
	commandHello commandKind = iota + 1
	commandState
	commandPing
	commandOffline
)

type command struct {
	kind       commandKind
	major      uint64
	generation uint64
	sequence   uint32
	selected   int8
	states     [slotCount]byte
}

type lineParser struct {
	buffer   [maxLineBytes - 1]byte
	length   int
	dropping bool
}

func (p *lineParser) reset() {
	p.length = 0
	p.dropping = false
}

func (p *lineParser) feed(b byte) (command, bool) {
	if p.dropping {
		if b == '\n' {
			p.reset()
		}
		return command{}, false
	}
	if b == '\n' {
		line := p.buffer[:p.length]
		p.length = 0
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		return parseCommand(string(line))
	}
	if p.length == len(p.buffer) {
		p.length = 0
		p.dropping = true
		return command{}, false
	}
	p.buffer[p.length] = b
	p.length++
	return command{}, false
}

func parseCommand(line string) (command, bool) {
	if line == "" || line[0] == ' ' || line[len(line)-1] == ' ' || strings.Contains(line, "  ") {
		return command{}, false
	}
	for i := range len(line) {
		if line[i] < 0x20 || line[i] > 0x7e {
			return command{}, false
		}
	}

	parts := strings.Split(line, " ")
	switch parts[0] {
	case "HELLO":
		if len(parts) != 3 || parts[1] != "HOST" {
			return command{}, false
		}
		major, ok := parseDecimal(parts[2], 0, ^uint64(0))
		return command{kind: commandHello, major: major}, ok
	case "STATE":
		if len(parts) != 4 {
			return command{}, false
		}
		generation, ok := parseDecimal(parts[1], 1, ^uint64(0))
		if !ok {
			return command{}, false
		}
		selected := int8(-1)
		if parts[2] != "-" {
			value, valid := parseDecimal(parts[2], 0, slotCount-1)
			if !valid {
				return command{}, false
			}
			selected = int8(value)
		}
		if len(parts[3]) != slotCount {
			return command{}, false
		}
		var states [slotCount]byte
		for i := range states {
			if !validState(parts[3][i]) {
				return command{}, false
			}
			states[i] = parts[3][i]
		}
		return command{kind: commandState, generation: generation, selected: selected, states: states}, true
	case "PING":
		if len(parts) != 2 {
			return command{}, false
		}
		sequence, ok := parseDecimal(parts[1], 0, uint64(^uint32(0)))
		return command{kind: commandPing, sequence: uint32(sequence)}, ok
	case "OFFLINE":
		if len(parts) != 2 {
			return command{}, false
		}
		generation, ok := parseDecimal(parts[1], 1, ^uint64(0))
		return command{kind: commandOffline, generation: generation}, ok
	default:
		return command{}, false
	}
}

func parseDecimal(value string, minimum, maximum uint64) (uint64, bool) {
	if value == "" || (len(value) > 1 && value[0] == '0') {
		return 0, false
	}
	for i := range len(value) {
		if value[i] < '0' || value[i] > '9' {
			return 0, false
		}
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed < minimum || parsed > maximum {
		return 0, false
	}
	return parsed, true
}

func validState(state byte) bool {
	switch state {
	case 'W', 'I', 'B', 'D', 'U', 'E':
		return true
	default:
		return false
	}
}

type panelState struct {
	online   bool
	selected int8
	states   [slotCount]byte
}

type session struct {
	handshake  bool
	panel      panelState
	generation uint64
}

func newSession() session {
	return session{panel: offlinePanel()}
}

func offlinePanel() panelState {
	return panelState{selected: -1}
}

func (s *session) resetUSB() bool {
	changed := s.panel != offlinePanel()
	s.handshake = false
	s.generation = 0
	s.panel = offlinePanel()
	return changed
}

func (s *session) handle(cmd command) (reply string, accepted, changed bool) {
	if cmd.kind == commandHello {
		before := s.panel
		s.handshake = cmd.major == 1
		s.generation = 0
		s.panel = offlinePanel()
		return "HELLO ZERO-KB02 1\n", true, before != s.panel
	}
	if !s.handshake {
		return "", false, false
	}

	switch cmd.kind {
	case commandState:
		before := s.panel
		s.generation = cmd.generation
		s.panel = panelState{online: true, selected: cmd.selected, states: cmd.states}
		return "", true, before != s.panel
	case commandPing:
		return "PONG " + strconv.FormatUint(uint64(cmd.sequence), 10) + "\n", true, false
	case commandOffline:
		before := s.panel
		s.generation = 0
		s.panel = offlinePanel()
		return "", true, before != s.panel
	default:
		return "", false, false
	}
}

func (s *session) expire() bool {
	if !s.handshake || !s.panel.online {
		return false
	}
	s.generation = 0
	s.panel = offlinePanel()
	return true
}

func (s session) canEmit() bool {
	return s.handshake && s.panel.online && s.generation != 0
}

func heartbeatExpired(elapsed time.Duration) bool {
	return elapsed >= heartbeatTimeout
}

func formatKey(generation uint64, slot int, down bool) string {
	edge := "UP"
	if down {
		edge = "DOWN"
	}
	return "KEY " + strconv.FormatUint(generation, 10) + " " + strconv.Itoa(slot) + " " + edge + "\n"
}

func formatEscape(generation uint64, down bool) string {
	edge := "UP"
	if down {
		edge = "DOWN"
	}
	return "ESC " + strconv.FormatUint(generation, 10) + " " + edge + "\n"
}

func formatPopup(generation uint64, down bool) string {
	edge := "UP"
	if down {
		edge = "DOWN"
	}
	return "POPUP " + strconv.FormatUint(generation, 10) + " " + edge + "\n"
}

func formatEncoder(generation uint64, event string) string {
	return "ENC " + strconv.FormatUint(generation, 10) + " " + event + "\n"
}

func formatJoystick(generation uint64, direction string) string {
	return "JOY " + strconv.FormatUint(generation, 10) + " " + direction + "\n"
}

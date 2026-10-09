package main

// One input stream, at most 32 unsent edges; overflow requires a new handshake.
type transmitQueue struct {
	lines        [32]string
	steps        [32]uint8
	pendingSteps int
	head, count  int
}

func (q *transmitQueue) reset() { *q = transmitQueue{} }

func (q *transmitQueue) push(line string) bool {
	return q.pushSteps(line, 0)
}

func (q *transmitQueue) pushEncoder(generation uint64, delta int) bool {
	if delta == 0 {
		return true
	}
	if delta < -32 || delta > 32 {
		return false
	}
	steps := delta
	if steps < 0 {
		steps = -steps
	}
	return q.pushSteps(formatEncoder(generation, delta), steps)
}

func (q *transmitQueue) pushSteps(line string, steps int) bool {
	if line == "" {
		return true
	}
	if q.count == len(q.lines) || q.pendingSteps+steps > 32 {
		return false
	}
	q.lines[(q.head+q.count)%len(q.lines)] = line
	q.steps[(q.head+q.count)%len(q.lines)] = uint8(steps)
	q.pendingSteps += steps
	q.count++
	return true
}

func (q *transmitQueue) pop() string {
	if q.count == 0 {
		return ""
	}
	line := q.lines[q.head]
	q.pendingSteps -= int(q.steps[q.head])
	q.steps[q.head] = 0
	q.lines[q.head] = ""
	q.head = (q.head + 1) % len(q.lines)
	q.count--
	return line
}

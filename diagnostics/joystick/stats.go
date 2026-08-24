package main

func waitForDTR(dtr func() bool, pause func()) {
	for !dtr() {
		pause()
	}
}

type stats struct {
	count      uint32
	sumX, sumY uint64
	minX, minY uint16
	maxX, maxY uint16
}

func (s *stats) add(x, y uint16) {
	if s.count == 0 || x < s.minX {
		s.minX = x
	}
	if s.count == 0 || x > s.maxX {
		s.maxX = x
	}
	if s.count == 0 || y < s.minY {
		s.minY = y
	}
	if s.count == 0 || y > s.maxY {
		s.maxY = y
	}
	s.count++
	s.sumX += uint64(x)
	s.sumY += uint64(y)
}

func (s stats) meanX() uint16 { return uint16(s.sumX / uint64(s.count)) }
func (s stats) meanY() uint16 { return uint16(s.sumY / uint64(s.count)) }

//go:build tinygo

package main

import (
	"machine"
	"machine/usb"
	"time"
)

const (
	samplePeriod     = 50 * time.Millisecond
	centerSamples    = 100
	directionSamples = 40
)

type phase struct {
	name    string
	samples int
}

func capture(p phase, x, y machine.ADC) {
	instruction := "move-and-hold"
	if p.name == "center" {
		instruction = "release-and-center"
	}
	println("READY", p.name, instruction)
	for remaining := 3; remaining > 0; remaining-- {
		println("CAPTURE_IN", p.name, remaining)
		time.Sleep(time.Second)
	}

	var result stats
	println("BEGIN", p.name, p.samples)
	for sample := 0; sample < p.samples; sample++ {
		rawX := x.Get()
		rawY := y.Get()
		result.add(rawX, rawY)
		println("SAMPLE", p.name, sample, rawX, rawY)
		time.Sleep(samplePeriod)
	}
	println("SUMMARY", p.name, result.count,
		result.minX, result.maxX, result.meanX(),
		result.minY, result.maxY, result.meanY())
	println("END", p.name, "release")
}

func main() {
	usb.Product = "zero-kb02-joystick-diag"
	usb.Serial = "zero-kb02-diag"

	machine.InitADC()
	x := machine.ADC{Pin: machine.GPIO29}
	y := machine.ADC{Pin: machine.GPIO28}
	x.Configure(machine.ADCConfig{})
	y.Configure(machine.ADCConfig{})

	waitForDTR(machine.Serial.DTR, func() {
		time.Sleep(10 * time.Millisecond)
	})
	for retry := 1; retry <= 3; retry++ {
		println("JOYSTICK_DIAGNOSTIC", "v2", "DTR", retry)
		time.Sleep(200 * time.Millisecond)
	}
	println("FORMAT", "SUMMARY phase count minX maxX meanX minY maxY meanY")

	for _, p := range []phase{
		{name: "center", samples: centerSamples},
		{name: "up", samples: directionSamples},
		{name: "down", samples: directionSamples},
		{name: "left", samples: directionSamples},
		{name: "right", samples: directionSamples},
	} {
		capture(p, x, y)
	}

	println("DONE")
	for {
		time.Sleep(time.Hour)
	}
}

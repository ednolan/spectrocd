package main

import (
	"math"
)

type Note int

func Frequency(note Note) float64 {
	const a4Pitch float64 = 440
	const a4 Note = 69 // MIDI note number of A4
	var halfStep float64 = math.Pow(2.0, 1.0/12.0)
	return a4Pitch * math.Pow(halfStep, float64(int(note-a4)))
}

func ClosestNote(frequency float64, highest Note) Note {
	// linear search closest note
	var closest Note = 0
	if frequency < Frequency(0) {
		return -1
	}
	δBest := math.Abs(frequency - Frequency(closest))
	for note := Note(1); note <= highest; note += 1 {
		δNew := math.Abs(frequency - Frequency(note))
		if δNew < δBest {
			closest = note
			δBest = δNew
		} else {
			return closest
		}
	}
	return -1
}

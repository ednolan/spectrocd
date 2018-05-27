package main

import (
	"github.com/cocoonlife/goalsa"
)

// initCaptureCDQ initializes an ALSA capture device at CD quality (stereo, 16 bit, 44100hz)
func InitCaptureCDQ(deviceName string, periodFrames int) (*alsa.CaptureDevice, error) {
	alsaBufferParams := alsa.BufferParams{
		BufferFrames: 0,
		PeriodFrames: periodFrames,
		Periods:      1,
	}
	return alsa.NewCaptureDevice(deviceName, 2, alsa.FormatS16LE, 44100, alsaBufferParams)
}

func ReadInterleavedSamples(device *alsa.CaptureDevice, readBuffer []int16) error {
	// read samples
	_, err := device.Read(readBuffer)
	return err
}

func InterleavedStereoToMono(stereo []int16, mono []int16) {
	if len(stereo) != len(mono)*2 {
		panic("stereo to mono conversion with mismatched buffer sizes")
	}
	for loop := 0; loop < len(stereo); loop += 2 {
		mono[loop/2] = int16((int32(stereo[loop]) + int32(stereo[loop+1])) / 2)
	}
}

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

func readInterleavedSamples(device *alsa.CaptureDevice, readBuffer []int16) error {
	// read samples
	_, err := device.Read(readBuffer)
	return err
}

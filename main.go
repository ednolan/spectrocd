package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func FFTBinNotes(bins int, highest Note) []Note {
	fftBinNotes := make([]Note, bins)
	for fftBin := 0; fftBin < bins; fftBin += 1 {
		var windowSize = bins * 2
		binFrequency := float64(fftBin * 44100 / windowSize)
		fftBinNotes[fftBin] = closestNote(binFrequency, highest)
	}
	return fftBinNotes
}

func SetLEDs(window []int16, settings []uint8, fftBinLEDs []int) {
	fftData := make([]float64, len(window)/2)
	ledSettingFloats := make([]float64, len(settings))
	ExecuteFFT(window, fftData)
	for fftBin, _ := range fftData {
		if fftBinLEDs[fftBin] > 0 {
			ledSettingFloats[fftBinLEDs[fftBin]] += fftData[fftBin]
		}
	}
	for led, settingFloat := range ledSettingFloats {
		settings[led] = uint8(settingFloat / 100000)
	}
}

func main() {
	// use first input device
	const deviceName string = "hw:1,0"

	// 64 LED array
	const nLEDs = 64

	// lowest note
	const lowest Note = 33 // low A, 55hz
	const highest Note = lowest + nLEDs - 1

	// 1s long buffer
	const bufferFrames int = 44100
	// mono audio buffer
	const monoBufferSize int = bufferFrames
	// read in periodFrames frames every loop iteration
	const periodFrames = 1000 // ~4.5ms
	const readBufferSize int = periodFrames * 2
	// set window size
	const windowFrames int = 8192 // ~186ms
	const bins = windowFrames / 2

	// initialize device
	device, err := InitCaptureCDQ(deviceName, periodFrames)
	if err != nil {
		panic(err)
	}
	defer device.Close()

	// set up sigint
	gotInterrupt := make(chan os.Signal, 1)
	signal.Notify(gotInterrupt, syscall.SIGINT)

	// allocate buffers
	readBuffer := make([]int16, readBufferSize)
	monoBuffer := make([]int16, monoBufferSize)
	monoBufferEnd := 0
	windowEnd := 0
	ledSettings := make([]uint8, nLEDs)
	fftBinNotes := FFTBinNotes(bins, highest)
	fftBinLEDs := make([]int, len(fftBinNotes))
	for loop, _ := range fftBinLEDs {
		fmt.Println(int(fftBinNotes[loop]))
		fftBinLEDs[loop] = func() int {
			switch {
			case lowest <= fftBinNotes[loop] && fftBinNotes[loop] <= highest:
				return int(fftBinNotes[loop] - lowest)
			default:
				return -1
			}
		}()
	}

	for {
		select {
		case <-gotInterrupt:
			fmt.Println("interrupt")
			os.Exit(0)
		default:
			err = readInterleavedSamples(device, readBuffer)
			if err != nil {
				panic(err)
			}

			// add samples to buffer
			// convert interleaved stereo samples to mono
			for loop := 0; loop < periodFrames; loop += 2 {
				monoSample := int16((int32(readBuffer[loop]) + int32(readBuffer[loop+1])) / 2)
				monoBuffer[monoBufferEnd+loop] = monoSample
			}
			monoBufferEnd += periodFrames
			windowEnd = monoBufferEnd

			// if end of buffer reached, copy window to beginning and reset bounds
			if monoBufferEnd+periodFrames > monoBufferSize {
				copy(monoBuffer[:windowFrames], monoBuffer[windowEnd-windowFrames:windowEnd])
				windowEnd = windowFrames
				monoBufferEnd = windowFrames
			}

			// if window is loaded, update LED settings
			if windowEnd >= windowFrames {
				SetLEDs(monoBuffer[windowEnd-windowFrames:windowEnd], ledSettings, fftBinLEDs)
			}

			for _, setting := range ledSettings {
				fmt.Printf("%3d", setting)
			}
			fmt.Printf("\n")
		}
	}
}

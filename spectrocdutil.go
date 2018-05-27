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

func run(deviceName string, nLEDs int, lowest Note, bufferFrames int, periodFrames int,
	windowFrames int) {
	var highest = Note(int(lowest) + nLEDs - 1)
	// mono audio buffer
	var monoBufferSize int = bufferFrames
	var readBufferSize int = periodFrames * 2
	var bins = windowFrames / 2
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
			err = ReadInterleavedSamples(device, readBuffer)
			if err != nil {
				panic(err)
			}

			// add samples to buffer
			// convert interleaved stereo samples to mono
			InterleavedStereoToMono(readBuffer, monoBuffer[monoBufferEnd:monoBufferEnd+periodFrames])
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

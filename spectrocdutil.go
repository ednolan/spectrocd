package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
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

func GetLEDLumas(window []int16, lumas []uint8, fftBinLEDs []int) {
	fftData := make([]float64, len(window)/2)
	lumaFloats := make([]float64, len(lumas))
	ExecuteFFT(window, fftData)
	for fftBin, _ := range fftData {
		if fftBinLEDs[fftBin] > 0 {
			lumaFloats[fftBinLEDs[fftBin]] += fftData[fftBin]
		}
	}
	for led, lumaFloat := range lumaFloats {
		lumas[led] = uint8(lumaFloat / 100000)
	}
}

func run(deviceName string, nLEDs int, lowest Note, bufferFrames int, periodFrames int,
	windowFrames int, ledLumas []uint8) {
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
	fftBinNotes := FFTBinNotes(bins, highest)
	fftBinLEDs := make([]int, len(fftBinNotes))
	for loop, _ := range fftBinLEDs {
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
		start := time.Now()
		select {
		case <-gotInterrupt:
			fmt.Println("interrupt")
			os.Exit(0)
		default:
			fmt.Println(time.Since(start))

			err = ReadInterleavedSamples(device, readBuffer)
			if err != nil {
				panic(err)
			}

			fmt.Println(time.Since(start))

			// add samples to buffer
			// convert interleaved stereo samples to mono
			InterleavedStereoToMono(readBuffer, monoBuffer[monoBufferEnd:monoBufferEnd+periodFrames])
			monoBufferEnd += periodFrames
			windowEnd = monoBufferEnd

			fmt.Println(time.Since(start))

			// if end of buffer reached, copy window to beginning and reset bounds
			if monoBufferEnd+periodFrames > monoBufferSize {
				copy(monoBuffer[:windowFrames], monoBuffer[windowEnd-windowFrames:windowEnd])
				windowEnd = windowFrames
				monoBufferEnd = windowFrames
			}

			fmt.Println(time.Since(start))

			// if window is loaded, update LED settings
			if windowEnd >= windowFrames {
				GetLEDLumas(monoBuffer[windowEnd-windowFrames:windowEnd], ledLumas, fftBinLEDs)
			}

			fmt.Println(time.Since(start))

			for _, luma := range ledLumas {
				fmt.Printf("%3d", luma)
			}
			fmt.Printf("\n")

			fmt.Println(time.Since(start))
		}
		loopsPerSec, _ := time.ParseDuration("1s")
		loopsPerSec /= time.Since(start)
		fmt.Printf("loop took %s, %d loops per sec\n", time.Since(start), loopsPerSec)
	}
}

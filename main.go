package main

import (
	"fmt"
	"github.com/cocoonlife/goalsa"
	"github.com/runningwild/go-fftw/fftw"
	"math"
	"math/cmplx"
	"os"
	"os/signal"
	"syscall"
	"time"
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

func ExecuteFFT(inputBuffer []int16, outputData []float64) {
	startTime := time.Now()
	fftwBuffer := fftw.NewArray(len(inputBuffer))
	transformPlan := fftw.NewPlan(fftwBuffer, fftwBuffer, fftw.Forward, fftw.Estimate)
	defer transformPlan.Destroy()
	for index, val := range inputBuffer {
		fftwBuffer.Elems[index] = complex(float64(val), 0)
	}
	transformPlan.Execute()
	for loop := 0; loop < len(inputBuffer)/2; loop += 1 {
		outputData[loop] = cmplx.Abs(fftwBuffer.Elems[loop])
	}
	fmt.Printf("FFT execution took %s\n", time.Since(startTime))
}

func FFTBinToLEDMapping(windowSize int, nLEDs int) []int {
	const lowestNoteFrequency float64 = 55.0
	ledFrequencies := make([]float64, nLEDs)
	ledFrequencies[0] = lowestNoteFrequency
	for note := 1; note < nLEDs; note += 1 {
		// multiplying by 12th root of 2 gives the note a half step higher in pitch
		ledFrequencies[note] = ledFrequencies[note-1] * math.Pow(2, 1.0/12.0)
	}
	fftBinNoteIndices := make([]int, windowSize/2)
	for fftBin := 0; fftBin < windowSize/2; fftBin += 1 {
		binFrequency := float64(fftBin * 44100 / windowSize)
		if binFrequency > ledFrequencies[0] && binFrequency < ledFrequencies[nLEDs-1] {
			// find closest note
			closestNote := 0
			difference := math.Abs(binFrequency - ledFrequencies[closestNote])
			for noteIndex, noteFrequency := range ledFrequencies {
				newDifference := math.Abs(binFrequency - noteFrequency)
				if newDifference < difference {
					closestNote = noteIndex
					difference = newDifference
				}
			}
			fftBinNoteIndices[fftBin] = closestNote
		} else {
			fftBinNoteIndices[fftBin] = -1
		}
	}
	return fftBinNoteIndices
}

func SetLEDs(window []int16, settings []uint8, fftBinNoteIndices []int) {
	fftData := make([]float64, len(window)/2)
	ledSettingFloats := make([]float64, len(settings))
	ExecuteFFT(window, fftData)
	for fftBin, _ := range fftData {
		if fftBinNoteIndices[fftBin] > 0 {
			ledSettingFloats[fftBinNoteIndices[fftBin]] += fftData[fftBin]
		}
	}
	for led, settingFloat := range ledSettingFloats {
		settings[led] = uint8(settingFloat / 100000)
	}
}

func main() {
	// use first input device
	const deviceName string = "hw:0,0"

	// 64 LED array
	const nLEDs = 64

	// 1s long buffer
	const bufferFrames int = 44100
	// mono audio buffer
	const monoBufferSize int = bufferFrames
	// read in periodFrames frames every loop iteration
	const periodFrames = 1000 // ~4.5ms
	const readBufferSize int = periodFrames * 2
	// set window size
	const windowFrames int = 8192 // ~186ms

	// initialize device
	device, err := InitCaptureCDQ(deviceName, periodFrames)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// set up sigint
	gotInterrupt := make(chan os.Signal, 1)
	signal.Notify(gotInterrupt, syscall.SIGINT)

	// allocate buffers
	readBuffer := make([]int16, readBufferSize)
	monoBuffer := make([]int16, monoBufferSize)
	monoBufferEnd := 0
	windowEnd := 0
	ledSettings := make([]uint8, nLEDs)
	fftBinNoteIndices := FFTBinToLEDMapping(windowFrames, nLEDs)

	for {
		select {
		case <-gotInterrupt:
			fmt.Println("interrupt")
			device.Close()
			os.Exit(0)
		default:
			// read samples
			_, err := device.Read(readBuffer)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// add samples to buffer
			// convert interleaved stereo samples to mono
			for loop := 0; loop < periodFrames; loop += 1 {
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
				SetLEDs(monoBuffer[windowEnd-windowFrames:windowEnd], ledSettings, fftBinNoteIndices)
			}

			for _, setting := range ledSettings {
				fmt.Printf("%3d ", setting)
			}
			fmt.Printf("\n")
		}
	}
}

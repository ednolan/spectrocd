package main

import (
	"github.com/runningwild/go-fftw/fftw"
	"math/cmplx"
	//"time"
)

func ExecuteFFT(inputBuffer []int16, outputData []float64) {
	if len(inputBuffer) != len(outputData)*2 {
		panic("FFT with mismatched buffer sizes")
	}
	//startTime := time.Now()
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
	//fmt.Printf("FFT execution took %s\n", time.Since(startTime))
}

package main

import (
	"fmt"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("spectrocd")               // name of config file (without extension)
	viper.AddConfigPath("$HOME/.config/spectrocd") // call multiple times to add many search paths
	err := viper.ReadInConfig()                    // Find and read the config file
	if err != nil {                                // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	var deviceName = viper.GetString("deviceName")

	// 64 LED array
	var nLEDs = viper.GetInt("nLEDs")

	// lowest note
	var lowest = Note(viper.GetInt("lowest"))

	var bufferFrames = viper.GetInt("bufferFrames")
	// read in periodFrames frames every loop iteration
	var periodFrames = viper.GetInt("periodFrames")
	// set window size
	var windowFrames int = viper.GetInt("windowFrames")

	ledLumas := make([]uint8, nLEDs)

	InitUnicornHat()

	go DisplayLoop(nLEDs, ledLumas)

	run(deviceName, nLEDs, lowest, bufferFrames, periodFrames, windowFrames, ledLumas)

}

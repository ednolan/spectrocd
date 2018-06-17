package main
import (
	"fmt"
	"github.com/jgarff/rpi_ws281x/golang/ws2811"
	//"time"
)

func InitUnicornHat() {
	ws2811.Init(10, 64, 64)
}

func DisplayLoop(n int, ledLumas []uint8) {
	for {
		for i := 0; i < n; i++ {
			ws2811.SetLed(i, uint32(ledLumas[i] / 4)<<12)
		}
		err := ws2811.Render()
		if err != nil {
			panic(fmt.Errorf("Render error: %s \n", err))
		}
		err2 := ws2811.Wait()
		if err2 != nil {
			panic(fmt.Errorf("Wait error: %s \n", err))
		}
	}
}

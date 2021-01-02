package main
import (
	"fmt"
	"github.com/ednolan/rpi_ws281x/golang/ws2811"
	"github.com/lucasb-eyer/go-colorful"
	"time"
)

func InitUnicornHat() {
	ws2811.Init(10, 64, 64)
}

func FlipRows(ledSettings []uint32) []uint32 {
	// TODO
	return ledSettings
}

func GetLEDSettings(ledLumas []uint8) []uint32 {
	var multiplier float64 = 5;
	base_hue := int64(float64(time.Now().Unix()) * multiplier) % 360
	hue := base_hue
	ledSettings := make([]uint32, 64)
	for i := 0; i < 64; i++ {
		value := float64(ledLumas[i]) / 255.0
		c := colorful.Hsv(float64(hue), 1.0, value).Clamped()
		setting := uint32(c.R * 255) << 16 | uint32(c.G * 255) << 8 | uint32(c.B * 255)
		ledSettings[i] = setting
		hue += 11
		hue %= 360
	}
	ledSettings = FlipRows(ledSettings)
	return ledSettings
}

func DisplayLoop(ledLumas []uint8) {
	for {
		ledSettings := GetLEDSettings(ledLumas)
		for i := 0; i < 64; i++ {
			ws2811.SetLed(i, ledSettings[i])
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

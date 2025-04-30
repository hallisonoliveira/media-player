package main

import (
	"log"
	"time"

	device "github.com/d2r2/go-hd44780"
	"github.com/d2r2/go-i2c"
)

func main() {
	i2c, err := i2c.NewI2C(0x27, 1) //I2C address might be different
	if err != nil {
		log.Fatal(err)
	}
	defer i2c.Close()

	lcd, err := device.NewLcd(i2c, device.LCD_16x2)
	if err != nil {
		log.Fatal(err)
	}

	err = lcd.BacklightOn()
	if err != nil {
		log.Fatal(err)
	}

	err = lcd.ShowMessage("Funciona!!", device.SHOW_LINE_1)
	if err != nil {
		log.Fatal(err)
	}

	err = lcd.ShowMessage("1234567890", device.SHOW_LINE_2)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(10 * time.Second)

	err = lcd.BacklightOff()
	if err != nil {
		log.Fatal(err)
	}
}

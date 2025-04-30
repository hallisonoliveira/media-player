package display

import (
	"errors"
	"strings"

	device "github.com/d2r2/go-hd44780"
	"github.com/d2r2/go-i2c"
	"github.com/d2r2/go-logger"
)

type Alignment int

const (
	CENTER Alignment = 0
	LEFT             = 1 << iota
	RIGHT
)

type Display struct {
	lcd *device.Lcd
	i2c *i2c.I2C
}

const displayWidth = 16

func NewDisplay() (*Display, error) {
	logger.ChangePackageLogLevel("i2c", logger.ErrorLevel)

	i2cBus, err := i2c.NewI2C(0x27, 1)
	if err != nil {
		return nil, err
	}

	lcd, err := device.NewLcd(i2cBus, device.LCD_16x2)
	if err != nil {
		i2cBus.Close()
		return nil, err
	}

	return &Display{lcd: lcd, i2c: i2cBus}, nil
}

func (d *Display) TurnBacklightOn() error {
	return d.lcd.BacklightOn()
}

func (d *Display) TurnBacklightOff() error {
	return d.lcd.BacklightOff()
}

func (d *Display) ShowText(message string, line int, align Alignment) error {
	var deviceLine device.ShowOptions
	switch line {
	case 1:
		deviceLine = device.SHOW_LINE_1
	case 2:
		deviceLine = device.SHOW_LINE_2
	default:
		return errors.New("invalid line number")
	}

	if len(message) > displayWidth {
		message = message[:displayWidth]
	}

	switch align {
	case CENTER:
		padding := (displayWidth - len(message)) / 2
		message = spaces(padding) + message
	case RIGHT:
		padding := displayWidth - len(message)
		message = spaces(padding) + message
	case LEFT:
		// already left
	default:
		// left side as fallback
	}

	return d.lcd.ShowMessage(message, deviceLine)
}

func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(" ", n)
}

func (display *Display) Clear() error {
	err := display.lcd.Clear()
	if err != nil {
		return err
	}
	return nil
}

func (d *Display) Close() error {
	return d.i2c.Close()
}

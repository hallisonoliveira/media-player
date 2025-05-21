package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

type inputEvent struct {
	Sec   uint64
	Usec  uint64
	Type  uint16
	Code  uint16
	Value int32
}

func main() {
	// Substitua por seu device IR
	f, err := os.Open("/dev/input/event0")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var ev inputEvent
	for {
		err := binary.Read(f, binary.LittleEndian, &ev)
		if err != nil {
			panic(err)
		}

		if ev.Type == 1 { // EV_KEY
			var action string
			switch ev.Value {
			case 0:
				action = "key up"
			case 1:
				action = "key down"
			case 2:
				action = "key repeat"
			}

			fmt.Printf("Key event: code=%d action=%s\n", ev.Code, action)
		}
	}
}

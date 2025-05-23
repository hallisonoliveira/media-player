package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"log"
	"media-player/pkg/queue"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const TOPIC = "remote-control"

type CommandData struct {
	Timestamp string `json:"timestamp"`
	Key       string `json:"key"`
}

type InputEvent struct {
	Sec   uint64
	Usec  uint64
	Type  uint16
	Code  uint16
	Value int32
}

type Code int

const (
	KEY_PROG1           Code = 148
	KEY_SWITCHVIDEOMODE Code = 227
	KEY_REFRESH         Code = 173
	KEY_DVD             Code = 389
	KEY_MEDIA           Code = 226
	KEY_PAGEUP          Code = 104
	KEY_STOP            Code = 128
	KEY_REWIND          Code = 168
	KEY_PLAYPAUSE       Code = 164
	KEY_FASTFORWARD     Code = 208
	KEY_PAGEDOWN        Code = 109
	KEY_PREVIOUS        Code = 412
	KEY_UP              Code = 103
	KEY_NEXT            Code = 407
	KEY_LEFT            Code = 105
	KEY_OK              Code = 352
	KEY_RIGHT           Code = 106
	KEY_BACKSPACE       Code = 14
	KEY_DOWN            Code = 108
	KEY_INFO            Code = 358
	KEY_VOLUMEDOWN      Code = 114
	KEY_MUTE            Code = 113
	KEY_VOLUMEUP        Code = 115
)

func main() {
	log.SetFlags(0)

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	defer stop()

	run(ctx)
}

func run(ctx context.Context) {
	input, err := os.Open("/dev/input/event0")
	if err != nil {
		log.Fatalf("Error openning connection with IR input: %v", err)
	}
	defer input.Close()
	q := queue.NewQueue()

	var event InputEvent
	for {
		input.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		err := binary.Read(input, binary.LittleEndian, &event)
		var key string
		select {
		case <-ctx.Done():
			log.Println("Stopping service...")
			return
		default:
			if err != nil {
				if os.IsTimeout(err) {
					continue
				}
				log.Fatalf("Error reading IR input: %v", err)
			}
			if event.Value == 1 { //key down
				switch event.Code {
				case uint16(KEY_PROG1):
					key = "KEY_PROG1"
				case uint16(KEY_SWITCHVIDEOMODE):
					key = "KEY_SWITCHVIDEOMODE"
				case uint16(KEY_REFRESH):
					key = "KEY_REFRESH"
				case uint16(KEY_DVD):
					key = "KEY_DVD"
				case uint16(KEY_MEDIA):
					key = "KEY_MEDIA"
				case uint16(KEY_PAGEUP):
					key = "KEY_PAGEUP"
				case uint16(KEY_STOP):
					key = "KEY_STOP"
				case uint16(KEY_REWIND):
					key = "KEY_REWIND"
				case uint16(KEY_PLAYPAUSE):
					key = "KEY_PLAYPAUSE"
				case uint16(KEY_FASTFORWARD):
					key = "KEY_FASTFORWARD"
				case uint16(KEY_PAGEDOWN):
					key = "KEY_PAGEDOWN"
				case uint16(KEY_PREVIOUS):
					key = "KEY_PREVIOUS"
				case uint16(KEY_UP):
					key = "KEY_UP"
				case uint16(KEY_NEXT):
					key = "KEY_NEXT"
				case uint16(KEY_LEFT):
					key = "KEY_LEFT"
				case uint16(KEY_OK):
					key = "KEY_OK"
				case uint16(KEY_RIGHT):
					key = "KEY_RIGHT"
				case uint16(KEY_BACKSPACE):
					key = "KEY_BACKSPACE"
				case uint16(KEY_DOWN):
					key = "KEY_DOWN"
				case uint16(KEY_INFO):
					key = "KEY_INFO"
				case uint16(KEY_VOLUMEDOWN):
					key = "KEY_VOLUMEDOWN"
				case uint16(KEY_MUTE):
					key = "KEY_MUTE"
				case uint16(KEY_VOLUMEUP):
					key = "KEY_VOLUMEUP"
				}
				sendKey(&key, q)
			}
		}
	}
}

func sendKey(key *string, q *queue.Queue) {
	command := CommandData{
		Timestamp: time.Now().Format("02-01-2006T15:04:05.000"),
		Key:       *key,
	}

	data, err := json.Marshal(command)
	if err != nil {
		log.Printf("Error serializing command: %v", err)
		return
	}

	if err := q.Publish(TOPIC, string(data)); err != nil {
		log.Printf("Error publishing to topic %s: %v", TOPIC, err)
	}
}

package main

import (
	"context"
	"encoding/json"
	"log"
	"media-player/pkg/display"
	"media-player/pkg/queue"
	"os"
	"os/signal"
	"syscall"
)

type DateTime struct {
	Timestamp string `json:"timestamp"`
	Date      string `json:"date"`
	Time      string `json:"time"`
}

const DATETIME_TOPIC = "datetime"

func main() {
	log.SetFlags(0)
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	screen, err := display.NewDisplay()
	if err != nil {
		log.Fatalf("Error creating display: %v", err)
	}
	defer func() {
		log.Println("Closing display connections")
		screen.Close()
		log.Println("Service finished")
	}()

	// Goroutine to cancel the context
	go func() {
		<-sigChan
		log.Println("Stopping service...")
		cancel()
		screen.TurnBacklightOff()
		screen.Close()
	}()

	if err := screen.Clear(); err != nil {
		log.Printf("Error clearing display: %v", err)
	}

	if err := screen.TurnBacklightOn(); err != nil {
		log.Printf("Error turning on backlight: %v", err)
	}

	q := queue.NewQueue()
	if err := q.Subscribe(ctx, handleMessage(screen), DATETIME_TOPIC); err != nil {
		log.Fatalf("Error subscribing to topic: %v", err)
	}
}

func handleMessage(screen *display.Display) func(channel, message string) {
	return func(channel, message string) {
		switch channel {
		case DATETIME_TOPIC:
			var datetime DateTime
			if err := json.Unmarshal([]byte(message), &datetime); err != nil {
				log.Printf("Error parsing JSON for datetime: %v", err)
				return
			}
			if err := screen.ShowText(datetime.Date, 1, display.CENTER); err != nil {
				log.Printf("Error showing date: %v", err)
			}
			if err := screen.ShowText(datetime.Time, 2, display.CENTER); err != nil {
				log.Printf("Error showing time: %v", err)
			}
		default:
			log.Printf("Received message on unknown topic: %s", channel)
		}
	}
}

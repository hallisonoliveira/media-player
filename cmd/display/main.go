package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"media-player/pkg/display"
	"media-player/pkg/queue"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	mu               sync.Mutex
	disableDateTimer *time.Timer
)

type DateTime struct {
	Timestamp string `json:"timestamp"`
	Date      string `json:"date"`
	Time      string `json:"time"`
}

type Item struct {
	IsDir bool   `json:"is_dir"`
	Name  string `json:"name"`
	Path  string `json:"path"`
}

type Explorer struct {
	Timestamp     string `json:"timestamp"`
	CurrentDir    string `json:"current_dir"`
	HasNext       bool   `json:"has_next`
	HasPrevious   bool   `json:"has_previous`
	SelectedIndex int    `json:"selected_index"`
	Items         []Item `json:"items"`
}

const DATETIME_TOPIC = "datetime"
const FILE_EXPLORER_TOPIC = "file-explorer"

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
	if err := q.Subscribe(ctx, handleMessage(screen), DATETIME_TOPIC, FILE_EXPLORER_TOPIC); err != nil {
		log.Fatalf("Error subscribing to topic: %v", err)
	}
}

func handleMessage(screen *display.Display) func(channel, message string) {
	var allowDateTime = true

	return func(channel, message string) {
		switch channel {
		case FILE_EXPLORER_TOPIC:
			var explorer Explorer
			if err := json.Unmarshal([]byte(message), &explorer); err != nil {
				log.Printf("Error parsing JSON for explorer: %v", err)
				return
			}
			handleExplorerUpdate(&explorer, screen, &allowDateTime)

		case DATETIME_TOPIC:
			if !allowDateTime {
				return
			}

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

func handleExplorerUpdate(explorer *Explorer, screen *display.Display, allowDateTime *bool) {
	line1, line2 := getExplorerLines(explorer)

	if err := screen.Clear(); err != nil {
		log.Printf("Error cleaning display: %v", err)
	}
	if err := screen.ShowText(line1, 1, display.LEFT); err != nil {
		log.Printf("Error showing explorer path: %v", err)
	}
	if err := screen.ShowText(line2, 2, display.LEFT); err != nil {
		log.Printf("Error showing explorer path: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	*allowDateTime = false

	if disableDateTimer != nil {
		disableDateTimer.Stop()
	}

	disableDateTimer = time.AfterFunc(10*time.Second, func() {
		mu.Lock()
		defer mu.Unlock()
		*allowDateTime = true
		if err := screen.Clear(); err != nil {
			log.Printf("Error cleaning display: %v", err)
		}
	})
}

func getExplorerLines(explorer *Explorer) (string, string) {
	const displayWidth = 16
	upArrow := "|"
	downArrow := "|"

	var lines [2]string

	start := explorer.SelectedIndex / 2 * 2
	end := start + 2
	if end > len(explorer.Items) {
		end = len(explorer.Items)
	}

	for i := start; i < end; i++ {
		item := explorer.Items[i]
		itemName := item.Name
		if item.IsDir {
			itemName = fmt.Sprintf("-%s", item.Name)
		}
		prefix := " "
		if i == explorer.SelectedIndex {
			prefix = "+"
		}
		line := fmt.Sprintf("%s%s", prefix, itemName)

		if len(line) > displayWidth {
			line = line[:displayWidth]
		} else {
			line = fmt.Sprintf("%-16s", line)
		}

		lines[i-start] = line
	}

	for i := range lines {
		if len(lines[i]) == 0 {
			lines[i] = strings.Repeat(" ", displayWidth)
		}
	}

	if explorer.HasPrevious {
		lines[0] = lines[0][:(displayWidth-1)] + upArrow
	}
	if explorer.HasNext && (start+1 < len(explorer.Items)) {
		lines[1] = lines[1][:(displayWidth-1)] + downArrow
	}

	return lines[0], lines[1]
}

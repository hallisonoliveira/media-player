package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"media-player/pkg/queue"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const REMOTE_CONTROL_TOPIC = "remote-control"
const FILE_EXPLORER_TOPIC = "file-explorer"

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

type CommandData struct {
	Timestamp string `json:"timestamp"`
	Key       string `json:"key"`
}

func NewExplorer(startDir string) (*Explorer, error) {
	entries, err := os.ReadDir(startDir)
	if err != nil {
		return nil, err
	}

	var items []Item

	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(startDir, entry.Name())

		if entry.IsDir() || filepath.Ext(name) == ".mp3" {
			item := Item{
				IsDir: entry.IsDir(),
				Name:  name,
				Path:  fullPath,
			}
			items = append(items, item)
		}
	}

	return &Explorer{
		Timestamp:     time.Now().Format("02-01-2006T15:04:05.000"),
		CurrentDir:    startDir,
		Items:         items,
		SelectedIndex: 0,
		HasNext:       false,
		HasPrevious:   false,
	}, nil
}

func main() {
	root := "/home/hallison/media"
	explorer, err := NewExplorer(root)
	if err != nil {
		fmt.Println(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	q := queue.NewQueue()
	if err := q.Subscribe(ctx, handleCommand(q, explorer), REMOTE_CONTROL_TOPIC); err != nil {
		fmt.Printf("Error subscribing to topic: %v", err)
	}

	go func() {
		<-sigChan
		log.Println("Stopping service...")
		cancel()
	}()
	<-ctx.Done()
}

func (e *Explorer) RefreshTimestamp() {
	e.Timestamp = time.Now().Format("02-01-2006T15:04:05.000")
}

func (e *Explorer) Next() {
	if e.SelectedIndex <= len(e.Items)-1 {
		e.SelectedIndex++
	}
}

func (e *Explorer) Previous() {
	if e.SelectedIndex > 0 {
		e.SelectedIndex--
	}
}

func (e *Explorer) Enter() error {
	if len(e.Items) == 0 {
		return nil
	}

	selected := e.Items[e.SelectedIndex]

	if e.Items[e.SelectedIndex].IsDir {
		fullPath := filepath.Join(e.CurrentDir, selected.Name)
		newExplorer, err := NewExplorer(fullPath)
		if err != nil {
			return err
		}
		fmt.Println("ENTER - Explorer: ", newExplorer)
		*e = *newExplorer
	} else {
		fmt.Println("Playing music")
	}
	return nil
}

func (e *Explorer) Back() error {
	parent := filepath.Dir(e.CurrentDir)
	newExplorer, err := NewExplorer(parent)
	if err != nil {
		return err
	}
	*e = *newExplorer
	return nil
}

func handleCommand(q *queue.Queue, explorer *Explorer) func(topic, message string) {
	return func(topic, message string) {
		var command CommandData
		if err := json.Unmarshal([]byte(message), &command); err != nil {
			log.Printf("Error parsing JSON for CommandData: %v", err)
			return
		}

		switch command.Key {
		case "KEY_DOWN":
			explorer.Next()
			publish(explorer, q)
		case "KEY_UP":
			explorer.Previous()
			publish(explorer, q)
		case "KEY_OK":
			if err := explorer.Enter(); err != nil {
				fmt.Println("Error on enter:", err)
			}
			publish(explorer, q)
		case "KEY_BACKSPACE":
			if err := explorer.Back(); err != nil {
				fmt.Println("Error on back:", err)
			}
			publish(explorer, q)
		}
	}
}

func publish(explorer *Explorer, q *queue.Queue) {
	explorer.HasNext = explorer.SelectedIndex < len(explorer.Items)-1
	explorer.HasPrevious = explorer.SelectedIndex > 0
	explorer.RefreshTimestamp()
	stateJson, _ := json.Marshal(explorer)
	if err := q.Publish(FILE_EXPLORER_TOPIC, string(stateJson)); err != nil {
		fmt.Printf("Error publishing to topic %s: %v", FILE_EXPLORER_TOPIC, err)
	}
}

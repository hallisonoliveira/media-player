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
)

// {
// 	"has_next": true,
// 	"has_previous": true,
// 	"item_selected": 1,
// 	"items": [
// 	  "../",
// 	  "-Dance/",
// 	  "-Rock/",
// 	  "[ALL]",
// 	  "Some Music.mp3"
// 	]
// }

const COMMAND_TOPIC = "command"

type Item struct {
	IsDir bool   `json:"is_dir"`
	Name  string `json:"name"`
	Path  string `json:"path"`
}

type Explorer struct {
	CurrentDir    string `json:"current_dir"`
	HasNext       bool   `json:"has_next`
	HasPrevious   bool   `json:"has_previous`
	SelectedIndex int    `json:"selected_index"`
	Items         []Item `json:"items"`
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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	q := queue.NewQueue()
	if err := q.Subscribe(ctx, handleCommand(q, explorer), COMMAND_TOPIC); err != nil {
		fmt.Printf("Error subscribing to topic: %v", err)
	}

	go func() {
		<-sigChan
		log.Println("Stopping service...")
		cancel()
	}()
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
		// Enviar para o canal do player, ex: "player.load"
		// Aqui você pode publicar o caminho do arquivo MP3 no Redis
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
		switch message {
		case "next":
			explorer.Next()
		case "previous":
			explorer.Previous()
		case "enter":
			if err := explorer.Enter(); err != nil {
				fmt.Println("Error on enter:", err)
			}
		case "back":
			if err := explorer.Back(); err != nil {
				fmt.Println("Error on back:", err)
			}
		}

		explorer.HasNext = explorer.SelectedIndex < len(explorer.Items)-1
		explorer.HasPrevious = explorer.SelectedIndex > 0

		stateJson, _ := json.Marshal(explorer)
		if err := q.Publish("file-explorer", string(stateJson)); err != nil {
			fmt.Printf("Error publishing to topic %s: %v", "file-explorer", err)
		}
	}
}

package main

import (
	"fmt"
	"media-player/pkg/queue"
	"time"
)

func main() {
	q := queue.NewQueue()

	// Simula receber botões de um controle IR
	buttons := []string{"play", "pause", "next", "previous", "stop"}

	for {
		for _, btn := range buttons {
			fmt.Printf("Sending command: %s\n", btn)
			err := q.Publish("player_commands", btn)
			if err != nil {
				fmt.Printf("Error publishing: %v\n", err)
			}
			time.Sleep(5 * time.Second) // Espera um pouco pra simular "pessoa apertando o botão"
		}
	}
}

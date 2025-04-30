package main

import (
	"context"
	"encoding/json"
	"log"
	"media-player/pkg/queue"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const topic = "datetime"

type DateTime struct {
	Timestamp string `json:"timestamp"`
	Date      string `json:"date"`
	Time      string `json:"time"`
}

func main() {
	log.SetFlags(0)

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	defer stop()

	run(ctx)
}

func run(ctx context.Context) {
	q := queue.NewQueue()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	loc := time.FixedZone("GMT-3", -3*60*60)
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping service...")
			return

		case t := <-ticker.C:
			datetime := DateTime{
				Timestamp: t.In(loc).Format("02-01-2006T15:04:05.000"),
				Date:      t.In(loc).Format("02-01-2006"),
				Time:      t.In(loc).Format("15:04:05"),
			}

			data, err := json.Marshal(datetime)
			if err != nil {
				log.Printf("Error serializing DateTime: %v", err)
				continue
			}

			log.Printf("%s", data)

			if err := q.Publish(topic, string(data)); err != nil {
				log.Printf("Error publishing to topic %s: %v", topic, err)
			}
		}
	}
}

package queue

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type Queue struct {
	client *redis.Client
}

func NewQueue() *Queue {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	return &Queue{client: rdb}
}

func (q *Queue) Publish(channel, message string) error {
	return q.client.Publish(ctx, channel, message).Err()
}

func (q *Queue) Subscribe(ctx context.Context, handler func(channel, message string), channels ...string) error {
	sub := q.client.Subscribe(ctx, channels...)

	_, err := sub.Receive(ctx)
	if err != nil {
		return err
	}

	ch := sub.Channel()
	for {
		select {
		case msg := <-ch:
			if msg == nil {
				return nil
			}
			handler(msg.Channel, msg.Payload)
		case <-ctx.Done():
			return sub.Close()
		}
	}
}

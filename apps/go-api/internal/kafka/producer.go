package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type ExpenseEvent struct {
	Action     string  `json:"action"`
	ID         int64   `json:"id"`
	Month      int     `json:"month"`
	Year       int     `json:"year"`
	Amount     float64 `json:"amount"`
	PrevAmount float64 `json:"prev_amount,omitempty"`
}

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) PublishEvent(ctx context.Context, event ExpenseEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Value: data,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

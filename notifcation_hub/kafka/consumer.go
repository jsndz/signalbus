package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type Consumer struct{
	reader *kafka.Reader
}

func NewConsumer (topic string,brokers []string) *Consumer{
	return &Consumer{
		kafka.NewReader(kafka.ReaderConfig{
			Brokers:   []string{"localhost:9092", "localhost:9093", "localhost:9094"},
			Topic:     topic,
			Partition: 0,
			MaxBytes:  10e6, // 10MB
		}),
	}
}


func (c *Consumer ) ReadFromKafka(ctx context.Context)(*kafka.Message,error) {
	m, err := c.reader.ReadMessage(ctx)
	if err := c.reader.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}
	return &m,err
	
}

func (c *Consumer) Close() error {
    return c.reader.Close()
}


















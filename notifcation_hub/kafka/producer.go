package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type Producer struct{
	writer *kafka.Writer
}



func NewProducer(brokers string,topic string) *Producer{
	return &Producer{
		&kafka.Writer{
			Addr:     kafka.TCP("localhost:9092", "localhost:9093", "localhost:9094"),
			Topic:   topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
	
}

func(p Producer) WriteToKafka(ctx context.Context, key ,value string) error{
	err := p.writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(key),
			Value: []byte(value),
		},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
		return err
	}
	return nil
	
}
func (p *Producer) Close() error {
	return p.writer.Close()
}
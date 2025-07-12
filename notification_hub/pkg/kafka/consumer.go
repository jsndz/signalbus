package kafka

import (
	"context"

	"github.com/jsndz/signalbus/metrics"
	"github.com/segmentio/kafka-go"
)

type Consumer struct{
	reader *kafka.Reader
}

func NewConsumer (topic string,brokers []string) *Consumer{
	return &Consumer{
		kafka.NewReader(kafka.ReaderConfig{
			Brokers:   brokers,
			Topic:     topic,
			Partition: 0,
			MaxBytes:  10e6, // 10MB
		}),
	}
}


func (c *Consumer ) ReadFromKafka(ctx context.Context)(*kafka.Message,error) {
	m, err := c.reader.ReadMessage(ctx)
	if err != nil{
		metrics.KafkaSubscriberFailure.WithLabelValues(c.reader.Config().Topic).Inc()

		return nil, err
	}
	metrics.KafkaSubscriberSuccess.WithLabelValues(c.reader.Config().Topic).Inc()
	return &m,nil
}

func (c *Consumer) Close() error {
    return c.reader.Close()
}
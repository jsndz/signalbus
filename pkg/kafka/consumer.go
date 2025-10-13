package kafka

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"github.com/jsndz/signalbus/metrics"
	"github.com/jsndz/signalbus/pkg/utils"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func (c *Consumer) ReadFromKafka(ctx context.Context) (*kafka.Message, error) {
	m, err := c.reader.ReadMessage(ctx)
	if err != nil {
		metrics.KafkaSubscriberFailureTotal.WithLabelValues(c.reader.Config().Topic).Inc()
		return nil, err
	}
	go func() {
		for {
			if lag, err := c.reader.ReadLag(context.Background()); err == nil {
				metrics.KafkaConsumerLag.WithLabelValues(
					c.reader.Config().GroupID,
					c.reader.Config().Topic,
				).Set(float64(lag))
			}
			time.Sleep(10 * time.Second)
		}
	}()

	return &m, nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func NewConsumer(topic string, brokers []string, groupID string) *Consumer {
	metrics.KafkaRebalancesTotal.WithLabelValues(groupID).Inc()
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
			MaxBytes: 10e6, // 10MB
		}),
	}
}

func NewConsumerAvien(topic, groupID string) *Consumer {
	kafkaURL := utils.GetEnv("AVIEN_KAFKA_URL")

	keypair, caCertPool :=utils.Decode()

    dialer := &kafka.Dialer{
        Timeout:   10 * time.Second,
        DualStack: true,
        TLS: &tls.Config{
            Certificates: []tls.Certificate{keypair},
            RootCAs:      caCertPool,
        },
    }

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaURL},
		Topic:   topic,
		GroupID: groupID,
		Dialer:  dialer,
		MaxBytes: 10e6,
	})

	return &Consumer{
		reader: reader,
	}
}



func NewConsumerFromEnv(topic,groupID string) *Consumer {
	state := utils.GetEnv("STATE")
	

	switch state {
	case "prod":
		log.Println("Starting Kafka Consumer in PROD mode (Aiven)")
		return NewConsumerAvien(topic, groupID)
	case "dev":
		log.Println("Starting Kafka Consumer in DEV mode (local)")
		fallthrough
	default:
		broker := utils.GetEnv("KAFKA_BROKER")
		return NewConsumer(topic, []string{broker}, groupID)
	}
}

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

type Producer struct{
	writer *kafka.Writer
}



func NewProducer(brokers []string) *Producer{
	return &Producer{
		&kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Balancer: &kafka.LeastBytes{},
		},
	}
	
}

func(p *Producer) Publish(ctx context.Context,topic string, key ,value []byte) error{
	err := p.writer.WriteMessages(ctx,
		kafka.Message{
			Topic: topic,
			Key:   key,
			Value: value,
		},
	)
	if err != nil {
		log.Printf("failed to write messages: %v", err)
		metrics.KafkaPublisherFailure.WithLabelValues(topic).Inc()
		return err
	}
	metrics.KafkaPublisherSuccess.WithLabelValues(topic).Inc()
	return nil
	
}
func (p *Producer) Close() error {
	return p.writer.Close()
}


func NewProducerAvien()*Producer{
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

    writer := kafka.NewWriter(kafka.WriterConfig{
        Brokers:  []string{kafkaURL},
        Dialer:   dialer,
		
    })
	return &Producer{
		writer: writer,
	}
}


func NewProducerFromEnv() *Producer {
	state := utils.GetEnv("STATE")

	switch state {
	case "prod":
		log.Println("Starting Kafka Producer in PROD mode (Aiven)")
		return NewProducerAvien()
	case "dev":
		log.Println("Starting Kafka Producer in DEV mode (local)")
		fallthrough
	default:
		brokers := []string{utils.GetEnv("KAFKA_BROKER")}
		return NewProducer(brokers)
	}
}

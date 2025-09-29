package kafka

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"github.com/jsndz/signalbus/metrics"
	"github.com/jsndz/signalbus/pkg/utils"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
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

func (p *Producer) Publish(ctx context.Context, topic string, key, value []byte) error {
    tracer := otel.Tracer("kafka-producer")

    ctx, span := tracer.Start(ctx, "kafka.publish", trace.WithAttributes(
        attribute.String("messaging.system", "kafka"),
        attribute.String("messaging.destination", topic),
    ))
    defer span.End()

    headers := make([]kafka.Header, 0)
    carrier := propagation.MapCarrier{}
    otel.GetTextMapPropagator().Inject(ctx, carrier)

    for k, v := range carrier {
        headers = append(headers, kafka.Header{
            Key:   k,
            Value: []byte(v),
        })
    }

    err := p.writer.WriteMessages(ctx,
        kafka.Message{
            Topic:   topic,
            Key:     key,
            Value:   value,
            Headers: headers,
        },
    )

    if err != nil {
        log.Printf("failed to write messages: %v", err)
        metrics.KafkaPublishFailureTotal.WithLabelValues(topic).Inc()
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    span.SetStatus(codes.Ok, "message published")
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

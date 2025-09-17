package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/jsndz/signalbus/metrics"
	"github.com/jsndz/signalbus/pkg/gosms"
	"github.com/jsndz/signalbus/pkg/kafka"
	"github.com/jsndz/signalbus/pkg/repositories"
	"github.com/jsndz/signalbus/pkg/templates"
	"github.com/jsndz/signalbus/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const maxRetries = 3



func  HandleSMS(broker string, ctx context.Context, smsService gosms.Sender, logger *zap.Logger,tmplRepo *repositories.TemplateRepository,producer *kafka.Producer)  {
	topic:= "notification.sms"
	c := kafka.NewConsumerFromEnv(topic,"sms")
	defer c.Close()

	logger.Info("Starting Kafka consumer", zap.String("topic", topic), zap.String("broker", broker))

	for {
		select{
		case <-ctx.Done():
			logger.Info("Shutting down SignupConsumer", zap.String("topic", topic))
			return
		default:
			m,err:= c.ReadFromKafka(ctx)
			if err!= nil{
				logger.Error("Error reading Kafka message", zap.String("topic", topic), zap.Error(err))
				continue
			}
			var msg types.KafkaStreamData
            if err := json.Unmarshal(m.Value, &msg); err != nil {
                logger.Error("Failed to unmarshal SMS message",
                    zap.ByteString("raw", m.Value),
                    zap.Error(err),
                )
                continue
            }
			content ,err := templates.Render(msg.RecieverData,msg.GetTemplateData.TenantID.String(),msg.GetTemplateData.EventType,string(m.Key),msg.GetTemplateData.Locale,[]string{"html","text"},tmplRepo)
			if err != nil {
				logger.Error("Coudn't Render template",
					zap.Any("recieverData", msg.RecieverData),
					zap.Error(err),
				)
				continue
			}
			logger.Info("Kafka message received",
				zap.String("topic", topic),
				zap.ByteString("key", m.Key),
				zap.Int64("offset", m.Offset),
			)
			var user gosms.SMS
			recieverDataBytes, err := json.Marshal(msg.RecieverData)
			if err != nil {
				logger.Error("Failed to marshal RecieverData",
					zap.Any("recieverData", msg.RecieverData),
					zap.Error(err),
				)
				continue
			}
			if err := json.Unmarshal(recieverDataBytes, &user); err != nil {
				logger.Error("Failed to unmarshal SMS user",
					zap.ByteString("raw", recieverDataBytes),
					zap.Error(err),
				)
				continue
			}
			sms := gosms.NewSMS(user.To, string(content["text"]), gosms.WithIdempotencyKey(user.IdempotencyKey))
			SendSMSWithRetry(logger, smsService, sms,producer)
		}
	}
}

func  SendSMSWithRetry(Logger *zap.Logger,smsService gosms.Sender,sms gosms.SMS,producer *kafka.Producer) error {
	timer := prometheus.NewTimer(metrics.NotificationSendDuration.WithLabelValues("twilio", "sms_worker"))
	defer timer.ObserveDuration()

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := smsService.Send(sms)
		if err == nil {
			metrics.ExternalAPISuccess.WithLabelValues("twilio", "sms_worker").Inc()
			return nil
		}
		baseTime := 1 * time.Second
		backoffDelay := baseTime * time.Duration( 1 << ( attempt - 1 ) ) 
		jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
		waitTime := backoffDelay * jitter
		Logger.Warn("SMS send attempt failed, will retry",
			zap.Int("attempt", attempt),
			zap.Error(err),
			zap.Duration("retry_in", waitTime),
		)

		time.Sleep(waitTime)
	}

	err := fmt.Errorf("SendMail failed after %d retries", maxRetries)
	metrics.ExternalAPIFailure.WithLabelValues("twilio", "sms_worker").Inc()
	smsBytes, marshalErr := json.Marshal(sms)
	if marshalErr != nil {
		Logger.Error("Failed to marshal SMS for DLQ",
			zap.String("to", sms.To),
			zap.Error(marshalErr),
		)
		return marshalErr
	}
	producer.Publish(context.Background(), "notification.sms.dlq", []byte(sms.IdempotencyKey), smsBytes)

	Logger.Error("Final SMS send failure",
		zap.String("to", sms.To),
		zap.Error(err),
	)
	return err
}

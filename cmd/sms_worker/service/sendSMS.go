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
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const maxRetries = 3



func  HandleSMS(broker string, ctx context.Context, smsService gosms.Sender, logger *zap.Logger,tmpl *repositories.TemplateRepository)  {
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
			 logger.Info("Kafka message received",
				zap.String("topic", topic),
				zap.ByteString("key", m.Key),
				zap.Int64("offset", m.Offset),
			)
			var messageBody gosms.SMS
			
			if err:= json.Unmarshal(m.Value,&messageBody); err!=nil{
				logger.Error("Failed to unmarshal signup message",
					zap.ByteString("raw", m.Value),
					zap.Error(err),
				)
				continue
			}
			sms := gosms.NewSMS(messageBody.To,messageBody.Text,gosms.WithIdempotencyKey(messageBody.IdempotencyKey))
			SendSMSWithRetry(logger,smsService,sms)
		}
	}
}

func  SendSMSWithRetry(Logger *zap.Logger,smsService gosms.Sender,sms gosms.SMS) error {
	timer := prometheus.NewTimer(metrics.NotificationSendDuration.WithLabelValues("twilio", "sms_worker"))
	defer timer.ObserveDuration()

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := smsService.Send(sms)
		if err == nil {
			metrics.ExternalAPISuccess.WithLabelValues("twilio", "sms_worker").Inc()
			return nil
		}
		baseTime := 1 * time.Second
		backoffDelay := baseTime * time.Duration(1<<(attempt-1)) 
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

	Logger.Error("Final SMS send failure",
		zap.String("to", sms.To),
		zap.Error(err),
	)
	return err
}

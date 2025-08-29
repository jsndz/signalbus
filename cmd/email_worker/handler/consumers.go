package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jsndz/signalbus/metrics"
	"github.com/jsndz/signalbus/pkg/gomailer"
	"github.com/jsndz/signalbus/pkg/kafka"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const maxRetries = 3

func HandleMail(broker string, ctx context.Context, mailService gomailer.Mailer, logger *zap.Logger) {
	topic := "notification.email"
	c := kafka.NewConsumerFromEnv(topic,"email")
	defer c.Close()

	logger.Info("Starting Kafka consumer", zap.String("topic", topic), zap.String("broker", broker))

	for {
		select {
		case <-ctx.Done():
			logger.Info("Shutting down SignupConsumer", zap.String("topic", topic))
			return

		default:
			m, err := c.ReadFromKafka(ctx)
			if err != nil {
				logger.Error("Error reading Kafka message", zap.String("topic", topic), zap.Error(err))
				continue
			}

			logger.Info("Kafka message received",
				zap.String("topic", topic),
				zap.ByteString("key", m.Key),
				zap.Int64("offset", m.Offset),
			)

			var messageBody gomailer.Email
			err = json.Unmarshal(m.Value, &messageBody)
			if err != nil {
				logger.Error("Failed to unmarshal signup message",
					zap.ByteString("raw", m.Value),
					zap.Error(err),
				)
				continue
			}
			mail := gomailer.NewEmail(messageBody.From,messageBody.To,
				gomailer.WithHTML(messageBody.HTML),gomailer.WithText(messageBody.Text),
				gomailer.WithSubject(messageBody.Subject))

			SendEmailWithRetry(logger,mailService,mail);
		}
	}
}



func SendEmailWithRetry(Logger *zap.Logger,mailService gomailer.Mailer,mail gomailer.Email)(error){
	timer := prometheus.NewTimer(metrics.NotificationSendDuration.WithLabelValues("sendgrid", "email_worker"))
	defer timer.ObserveDuration()

	for attempt := 1; attempt <= maxRetries; attempt++ {

		err:= mailService.Send(mail)
		if err == nil {
			metrics.ExternalAPISuccess.WithLabelValues("sendgrid", "email_worker").Inc()

			return nil
		}

		waitTime := time.Duration(attempt*2) * time.Second

		Logger.Warn("Email send failed, will retry",
			zap.Int("attempt", attempt),
			zap.Duration("retry_in", waitTime),
			zap.Error(err),
		)

		time.Sleep(waitTime)
	}

	err := fmt.Errorf("SendMail failed after %d retries", maxRetries)
	metrics.ExternalAPIFailure.WithLabelValues("sendgrid", "email_worker").Inc()

	Logger.Error("Final email send failure",
		zap.Error(err),
	)
	return err
}

package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/jsndz/signalbus/metrics"
	"github.com/jsndz/signalbus/pkg/gomailer"
	"github.com/jsndz/signalbus/pkg/kafka"
	"github.com/jsndz/signalbus/pkg/repositories"
	"github.com/jsndz/signalbus/pkg/templates"
	"github.com/jsndz/signalbus/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const maxRetries = 3

func HandleMail(broker string, ctx context.Context, mailService gomailer.Mailer, logger *zap.Logger,tmplRepo *repositories.TemplateRepository) {
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
			var user gomailer.Email
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
			mail := gomailer.NewEmail(user.From,user.To,
				gomailer.WithHTML(string(content["html"])),gomailer.WithText(string(content["text"])),
				gomailer.WithSubject(user.Subject))

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
		baseTime := 1 * time.Second
		backoffDelay := baseTime * time.Duration(1<<(attempt-1)) 
		jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
		waitTime := backoffDelay * jitter

		Logger.Warn("Email send failed, will retry",
			zap.Int("attempt", attempt),
			zap.Duration("retry_in", waitTime),
			zap.Error(err),
		)

		time.Sleep(waitTime)
	}

	err := fmt.Errorf("SendMail failed after %d retries", maxRetries)
	metrics.ExternalAPIFailure.WithLabelValues("sendgrid", "email_worker").Inc()

	Logger.Error("Permanent email failure - message lost",
		zap.String("to", strings.Join(mail.To, ",")),
		zap.String("subject", mail.Subject),
		zap.String("reason", "max_retries_exceeded"),
	)
	return err
}

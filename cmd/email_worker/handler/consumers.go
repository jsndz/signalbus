package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

func HandleMail(broker string, ctx context.Context, mailService gomailer.Mailer, logger *zap.Logger,tmplRepo *repositories.TemplateRepository,producer *kafka.Producer) {
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
			log.Println(msg)
			content ,err := templates.Render(msg.InTemplateData,msg.GetTemplateData.TenantID.String(),"email",msg.GetTemplateData.EventType,msg.GetTemplateData.Locale,[]string{"html","text"},tmplRepo)
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
				logger.Error("Failed to unmarshal mail user",
				zap.ByteString("raw", recieverDataBytes),
				zap.Error(err),
			)
			continue
			}
			mail := gomailer.NewEmail(user.From,user.To,
				gomailer.WithHTML(string(content["html"])),gomailer.WithText(string(content["text"])),
				gomailer.WithSubject(user.Subject))

			SendEmailWithRetry(logger,mailService,mail,producer);
		}
	}
}



func SendEmailWithRetry(Logger *zap.Logger,mailService gomailer.Mailer,mail gomailer.Email,producer *kafka.Producer)(error){
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
	if err == nil {
		return nil
	}
	metrics.ExternalAPIFailure.WithLabelValues("sendgrid", "email_worker").Inc()

	Logger.Error("Permanent email failure - sending it to DLQ",
		zap.String("to", strings.Join(mail.To, ",")),
		zap.String("subject", mail.Subject),
		zap.String("reason", "max_retries_exceeded"),
	)
	mailBytes, err := json.Marshal(mail)
	if err != nil {
		Logger.Error("Failed to marshal email for DLQ",
			zap.String("to", strings.Join(mail.To, ",")),
			zap.String("subject", mail.Subject),
			zap.Error(err),
		)
		return err
	}
	producer.Publish(context.Background(), "notification.sms.dlq", []byte(mail.IdempotencyKey), mailBytes)
	return err
}

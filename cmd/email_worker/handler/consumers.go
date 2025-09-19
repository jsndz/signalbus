package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jsndz/signalbus/metrics"
	"github.com/jsndz/signalbus/pkg/gomailer"
	"github.com/jsndz/signalbus/pkg/kafka"
	"github.com/jsndz/signalbus/pkg/models"
	"github.com/jsndz/signalbus/pkg/repositories"
	"github.com/jsndz/signalbus/pkg/templates"
	"github.com/jsndz/signalbus/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const maxRetries = 3

func HandleMail(broker string, ctx context.Context, mailService gomailer.Mailer, logger *zap.Logger,tmplRepo *repositories.TemplateRepository,notificationRepo *repositories.NotificationRepository,producer *kafka.Producer) {
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

			SendEmailWithRetry(logger,mailService,mail,producer,msg.NotificationId,notificationRepo);
		}
	}
}



func SendEmailWithRetry(
    logger *zap.Logger,
    mailService gomailer.Mailer,
    mail gomailer.Email,
    producer *kafka.Producer,
    notificationID uuid.UUID,
    notificationRepo *repositories.NotificationRepository,
) error {
    timer := prometheus.NewTimer(metrics.NotificationSendDuration.WithLabelValues("sendgrid", "email_worker"))
    defer timer.ObserveDuration()

    for attempt := 1; attempt <= maxRetries; attempt++ {
        start := time.Now()
        err := mailService.Send(mail)
        latency := time.Since(start).Milliseconds()

        if err == nil {
            notificationRepo.UpdateStatus(notificationID, "delivered")
            notificationRepo.CreateAttempt(&models.DeliveryAttempt{
                NotificationID: notificationID,
                Channel:        "email",
                Provider:       "sendgrid",
                Status:         "delivered",
                Try:            attempt,
                LatencyMs:      latency,
            })
            metrics.ExternalAPISuccessTotal.WithLabelValues("sendgrid", "email_worker").Inc()
            return nil
        }

        backoffDelay := time.Second * time.Duration(1<<(attempt-1))
        jitter := time.Duration(rand.Intn(500)) * time.Millisecond
        waitTime := backoffDelay + jitter

        notificationRepo.CreateAttempt(&models.DeliveryAttempt{
            NotificationID: notificationID,
            Channel:        "email",
            Provider:       "sendgrid",
            Status:         "retrying",
            Error:          err.Error(),
            Try:            attempt,
            LatencyMs:      latency,
        })

        logger.Warn("Email send failed, will retry",
            zap.Int("attempt", attempt),
            zap.Duration("retry_in", waitTime),
            zap.Error(err),
        )

        time.Sleep(waitTime)
    }

    metrics.ExternalAPIFailureTotal.WithLabelValues("sendgrid", "email_worker").Inc()
    notificationRepo.UpdateStatus(notificationID, "failed")

    mailBytes, mErr := json.Marshal(mail)
    if mErr != nil {
        logger.Error("Failed to marshal email for DLQ",
            zap.String("to", strings.Join(mail.To, ",")),
            zap.String("subject", mail.Subject),
            zap.Error(mErr),
        )
        return mErr
    }

    logger.Error("Permanent email failure - sending to DLQ",
        zap.String("to", strings.Join(mail.To, ",")),
        zap.String("subject", mail.Subject),
    )

	producer.Publish(context.Background(), "notification.email.dlq", notificationID[:], mailBytes)
	notificationRepo.CreateAttempt(&models.DeliveryAttempt{
            NotificationID: notificationID,
            Channel:        "email",
            Provider:       "sendgrid",
            Status:         "dlq",
			Message: mailBytes, 
    })
    return fmt.Errorf("permanent email failure after %d retries", maxRetries)
}

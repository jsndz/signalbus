package handler

import (
	"context"
	"encoding/json"
	"fmt"
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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const maxRetries = 3

func HandleMail(broker string, 
    ctx context.Context, mailService gomailer.Mailer, 
    logger *zap.Logger,tmplRepo *repositories.TemplateRepository,
    notificationRepo *repositories.NotificationRepository,producer *kafka.Producer,
    provider string,tracer trace.Tracer,
) {
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
            msgCtx := ctx
            if len(m.Headers) > 0 {
                 carrier := make(map[string]string)
                for _, h := range m.Headers {
                    carrier[h.Key] = string(h.Value)
                }
                msgCtx = otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(carrier))
            }
            emailCtx, span := tracer.Start(msgCtx, "handle-email")
            func(){
                defer span.End() 
                var msg types.KafkaStreamData
                if err := json.Unmarshal(m.Value, &msg); err != nil {

                    logger.Error("Failed to unmarshal SMS message",
                        zap.ByteString("raw", m.Value),
                        zap.Error(err),
                    )
                    return
                }    
                 _, tmplspan := tracer.Start(emailCtx, "template extraction")

                var htmlContent, textContent string            
                if msg.GetTemplateData == nil {
                    htmlContent = msg.HTMLMessage
                    textContent = msg.TextMessage
                    
                    if htmlContent == "" && textContent == "" {
                        logger.Error("No content provided - neither template nor custom message")
                        return
                    }
                 }   else {
                    content ,err := templates.Render(
                        msg.InTemplateData,
                        msg.GetTemplateData.TenantID.String(),
                        "email",
                        msg.GetTemplateData.EventType,
                        msg.GetTemplateData.Locale,
                        []string{"html","text"},
                        tmplRepo)
               
                    if err != nil {
                        tmplspan.SetStatus(codes.Error, "couldn't extract template")
                        logger.Error("Coudn't Render template",
                            zap.Any("recieverData", msg.RecieverData),
                            zap.Error(err),
                        )
                        return
                    }
                    defer tmplspan.End()
                    htmlContent = string(content["html"])
                    textContent = string(content["text"])
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
                    return
                }
                if err := json.Unmarshal(recieverDataBytes, &user); err != nil {
                    logger.Error("Failed to unmarshal mail user",
                    zap.ByteString("raw", recieverDataBytes),
                    zap.Error(err),
                )
                return
                }
                mail := gomailer.NewEmail(user.From,user.To,
                    gomailer.WithHTML(htmlContent),gomailer.WithText(textContent),
                    gomailer.WithSubject(user.Subject))

                SendEmailWithRetry(emailCtx,logger,mailService,mail,producer,msg.NotificationId,notificationRepo,provider,tracer);
            }()
		}
	}
}

func SendEmailWithRetry(
    ctx context.Context,
    logger *zap.Logger,
    mailService gomailer.Mailer,
    mail gomailer.Email,
    producer *kafka.Producer,
    notificationID uuid.UUID,
    notificationRepo *repositories.NotificationRepository,
    provider string,
    tracer trace.Tracer,
) error {
    timer := prometheus.NewTimer(metrics.NotificationSendDuration.WithLabelValues(provider, "email_worker"))
    defer timer.ObserveDuration()
    _, sendSpan := tracer.Start(ctx, "send-email")
	defer sendSpan.End()
    for attempt := 1; attempt <= maxRetries; attempt++ {
        start := time.Now()
        apiTimer := prometheus.NewTimer(metrics.ExternalAPIDuration.WithLabelValues(provider, "email"))
        _,err := mailService.Send(mail)
        apiTimer.ObserveDuration()
        latency := time.Since(start).Milliseconds()

        if err == nil {
            metrics.NotificationsAttemptedTotal.WithLabelValues("email", "success", provider).Inc()

            notificationRepo.UpdateStatus(notificationID, "delivered")
            notificationRepo.CreateAttempt(&models.DeliveryAttempt{
                NotificationID: notificationID,
                Channel:        "email",
                Provider:       provider,
                Status:         "delivered",
                Try:            attempt,
                LatencyMs:      latency,
            })
            metrics.ExternalAPISuccessTotal.WithLabelValues(provider, "email_worker").Inc()

            return nil
        }
        sendSpan.AddEvent(fmt.Sprintf("Retry %d failed", attempt))
		sendSpan.RecordError(err)
        metrics.NotificationsAttemptedTotal.WithLabelValues("email", "failed", provider).Inc()

        backoffDelay := time.Second * time.Duration(1<<(attempt-1))
        jitter := time.Duration(rand.Intn(500)) * time.Millisecond
        waitTime := backoffDelay + jitter
   		metrics.NotificationRetriesTotal.WithLabelValues("provider_error","email").Inc()
        notificationRepo.CreateAttempt(&models.DeliveryAttempt{
            NotificationID: notificationID,
            Channel:        "email",
            Provider:       provider,
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

    metrics.NotificationsAttemptedTotal.WithLabelValues("email", "failed",provider).Inc()
    metrics.ExternalAPIFailureTotal.WithLabelValues(provider, "email_worker").Inc()
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
     _, dlqSpan := tracer.Start(ctx, "publish-dlq")
	defer dlqSpan.End()
    logger.Error("Permanent email failure - sending to DLQ",
        zap.String("to", strings.Join(mail.To, ",")),
        zap.String("subject", mail.Subject),
    )

   err:= producer.Publish(context.Background(), "notification.email.dlq", notificationID[:], mailBytes)

    if err != nil {
        logger.Error("couldnt send to DLQ",
            zap.String("to", strings.Join(mail.To, ",")),
            zap.String("subject", mail.Subject),
        )
        dlqSpan.RecordError(err)
        dlqSpan.SetStatus(codes.Error, err.Error())
    } else {
        dlqSpan.SetStatus(codes.Ok, "dlq published")
    }

	metrics.NotificationDLQTotal.WithLabelValues("provider_error","email")
    notificationRepo.CreateAttempt(&models.DeliveryAttempt{
        NotificationID: notificationID,
        Channel:        "email",
        Provider:       provider,
        Status:         "dlq",
        Message:        mailBytes,
    })
    return fmt.Errorf("permanent email failure after %d retries", maxRetries)
}

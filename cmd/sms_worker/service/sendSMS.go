package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jsndz/signalbus/metrics"
	"github.com/jsndz/signalbus/pkg/gosms"
	"github.com/jsndz/signalbus/pkg/kafka"
	"github.com/jsndz/signalbus/pkg/models"
	"github.com/jsndz/signalbus/pkg/repositories"
	"github.com/jsndz/signalbus/pkg/templates"
	"github.com/jsndz/signalbus/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const maxRetries = 3


func HandleSMS(
    broker string, ctx context.Context,
    smsService gosms.Sender,
    logger *zap.Logger,
    tmplRepo *repositories.TemplateRepository,
    notificationRepo *repositories.NotificationRepository,
    producer *kafka.Producer,
        provider string,

) {
    topic := "notification.sms"
    c := kafka.NewConsumerFromEnv(topic, "sms")
    defer c.Close()
    
    logger.Info("Starting Kafka consumer", zap.String("topic", topic), zap.String("broker", broker))
    
    for {
        select {
        case <-ctx.Done():
            logger.Info("Shutting down SMS Consumer", zap.String("topic", topic))
            return
        default:
            m, err := c.ReadFromKafka(ctx)
            if err != nil {
                logger.Error("Error reading Kafka message", zap.String("topic", topic), zap.Error(err))
                continue
            }
            
            if len(m.Headers) > 0 {
                carrier := make(map[string]string)
                for _, h := range m.Headers {
                    carrier[h.Key] = string(h.Value)
                }
            }
            
            func() {
                var msg types.KafkaStreamData
                if err := json.Unmarshal(m.Value, &msg); err != nil {
                    logger.Error("Failed to unmarshal SMS message",
                        zap.ByteString("raw", m.Value),
                        zap.Error(err),
                    )
                    return
                }

                logger.Info("Kafka message received",
                    zap.String("topic", topic),
                    zap.ByteString("key", m.Key),
                    zap.Int64("offset", m.Offset),
                )

                var textContent string

                if msg.GetTemplateData != nil {
                    content, err := templates.Render(
                        msg.InTemplateData,
                        msg.GetTemplateData.TenantID.String(),
                        "sms",  
                        msg.GetTemplateData.EventType,
                        msg.GetTemplateData.Locale,
                        []string{"text"},
                        tmplRepo,
                    )
                    if err != nil {
                        logger.Error("Couldn't render template",
                            zap.Any("recieverData", msg.RecieverData),
                            zap.Error(err),
                        )
                        return
                    }
                    textContent = string(content["text"])
                } else {
                    textContent = msg.TextMessage
                    if textContent == "" {
                        logger.Error("No content provided - neither template nor custom message")
                        return
                    }
                }

                var user gosms.SMS
                recieverDataBytes, err := json.Marshal(msg.RecieverData)
                if err != nil {
                    logger.Error("Failed to marshal RecieverData",
                        zap.Any("recieverData", msg.RecieverData),
                        zap.Error(err),
                    )
                    return
                }
                
                if err := json.Unmarshal(recieverDataBytes, &user); err != nil {
                    logger.Error("Failed to unmarshal SMS user",
                        zap.ByteString("raw", recieverDataBytes),
                        zap.Error(err),
                    )
                    return
                }

                sms := gosms.NewSMS(
                    user.To, 
                    textContent, 
                    gosms.WithIdempotencyKey(msg.IdempotencyKey), 
                )
                
                SendSMSWithRetry(
                    logger, 
                    smsService, 
                    sms, 
                    producer, 
                    msg.NotificationId, 
                    notificationRepo,provider,
                )
            }()
        }
    }
}
func SendSMSWithRetry(
    logger *zap.Logger,
    smsService gosms.Sender,
    sms gosms.SMS,
    producer *kafka.Producer,
    notificationID uuid.UUID,
    notificationRepo *repositories.NotificationRepository,
        provider string,

) error {
    timer := prometheus.NewTimer(metrics.NotificationSendDuration.WithLabelValues(provider, "sms_worker"))
    defer timer.ObserveDuration()
    for attempt := 1; attempt <= maxRetries; attempt++ {
        start := time.Now()
        apiTimer := prometheus.NewTimer(metrics.ExternalAPIDuration.WithLabelValues(provider, "email"))
        resp,err := smsService.Send(sms)
        apiTimer.ObserveDuration()
        latency := time.Since(start).Milliseconds()

        if err == nil {
            notificationRepo.UpdateStatus(notificationID, "delivered")
            notificationRepo.CreateAttempt(&models.DeliveryAttempt{
                NotificationID: notificationID,
                Channel:        "sms",
                Provider:       provider,
                Status:         "delivered",
                Try:            attempt,
                LatencyMs:      latency,
            })
            metrics.ExternalAPISuccessTotal.WithLabelValues(provider, "email_worker").Inc()
            metrics.NotificationsAttemptedTotal.WithLabelValues("sms", "success", provider).Inc()
            return nil
        }
        metrics.NotificationRetriesTotal.WithLabelValues("sms", provider, "provider_error").Inc()
        metrics.NotificationsAttemptedTotal.WithLabelValues("sms", "failed", provider).Inc()
        notificationRepo.CreateAttempt(&models.DeliveryAttempt{
            NotificationID: notificationID,
            Channel:        "sms",
            Provider:       provider,
            Status:         "failed",
            Error:          err.Error(),
            Try:            attempt,
            LatencyMs:      latency,
        })

        backoffDelay := time.Second * time.Duration(1<<(attempt-1))
        jitter := time.Duration(rand.Intn(500)) * time.Millisecond
        waitTime := backoffDelay + jitter

        logger.Warn("SMS send attempt failed, will retry",
            zap.Int("attempt", attempt),
            zap.Error(err),
            zap.Duration("retry_in", waitTime),
        )

        time.Sleep(waitTime)
    }

    metrics.NotificationsAttemptedTotal.WithLabelValues("sms", "failed", provider).Inc()
    metrics.ExternalAPIFailureTotal.WithLabelValues(provider, "sms_worker").Inc()
    notificationRepo.UpdateStatus(notificationID, "failed")

    smsBytes, marshalErr := json.Marshal(sms)


    if marshalErr != nil {
        logger.Error("Failed to marshal SMS for DLQ",
            zap.String("to", sms.To),
            zap.Error(marshalErr),
        )
        
        return marshalErr
    }

    err := producer.Publish(context.Background(), "notification.sms.dlq", notificationID[:], smsBytes)
    if err != nil{

    }
	metrics.NotificationDLQTotal.WithLabelValues("provider_error","sms")

    logger.Error("Final SMS send failure",
        zap.String("to", sms.To),
        zap.Error(fmt.Errorf("SendSMS failed after %d retries", maxRetries)),
    )

    return fmt.Errorf("SendSMS failed after %d retries", maxRetries)
}

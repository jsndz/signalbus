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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const maxRetries = 3



func  HandleSMS(
    broker string, ctx context.Context, 
    smsService gosms.Sender, 
    logger *zap.Logger,
    tmplRepo *repositories.TemplateRepository,
    notificationRepo *repositories.NotificationRepository,
    producer *kafka.Producer,
    tracer trace.Tracer,
) {
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
            msgCtx := ctx
            if len(m.Headers) > 0 {
                 carrier := make(map[string]string)
                for _, h := range m.Headers {
                    carrier[h.Key] = string(h.Value)
                }
                msgCtx = otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(carrier))
                
            }
            smsCtx, span := tracer.Start(msgCtx, "handle-sms")
            func(){
                defer span.End() 
                
                var msg types.KafkaStreamData
                if err := json.Unmarshal(m.Value, &msg); err != nil {
                    span.RecordError(err)
                    span.SetStatus(codes.Error, "failed to unmarshal Kafka message")
                    logger.Error("Failed to unmarshal SMS message",
                        zap.ByteString("raw", m.Value),
                        zap.Error(err),
                    )
                    return
                }
                _, tmplspan := tracer.Start(smsCtx, "template extraction")

                content ,err := templates.Render(msg.RecieverData,
                    msg.GetTemplateData.TenantID.String(),
                    msg.GetTemplateData.EventType,string(m.Key),
                    msg.GetTemplateData.Locale,[]string{"html","text"},tmplRepo)
                if err != nil {
                    tmplspan.SetStatus(codes.Error, "couldn't extract template")
                    logger.Error("Coudn't Render template",
                        zap.Any("recieverData", msg.RecieverData),
                        zap.Error(err),
                    )
                    return
                }
                tmplspan.End()

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
                    return
                }
                if err := json.Unmarshal(recieverDataBytes, &user); err != nil {
                    logger.Error("Failed to unmarshal SMS user",
                        zap.ByteString("raw", recieverDataBytes),
                        zap.Error(err),
                    )
                    return
                }
                
                sms := gosms.NewSMS(user.To, string(content["text"]), gosms.WithIdempotencyKey(user.IdempotencyKey))
                SendSMSWithRetry(smsCtx,logger, smsService, sms,producer,msg.NotificationId,notificationRepo,tracer)
            }()
            }
	}
}
func SendSMSWithRetry(
    	ctx context.Context,
    logger *zap.Logger,
    smsService gosms.Sender,
    sms gosms.SMS,
    producer *kafka.Producer,
    notificationID uuid.UUID,
    notificationRepo *repositories.NotificationRepository,
    tracer trace.Tracer,
) error {
    timer := prometheus.NewTimer(metrics.NotificationSendDuration.WithLabelValues("twilio", "sms_worker"))
    defer timer.ObserveDuration()
    _, sendSpan := tracer.Start(ctx, "send-sms")
	defer sendSpan.End()
    for attempt := 1; attempt <= maxRetries; attempt++ {
        start := time.Now()
        apiTimer := prometheus.NewTimer(metrics.ExternalAPIDuration.WithLabelValues("sendgrid", "email"))
        err := smsService.Send(sms)
        apiTimer.ObserveDuration()
        latency := time.Since(start).Milliseconds()

        if err == nil {
            notificationRepo.UpdateStatus(notificationID, "delivered")
            notificationRepo.CreateAttempt(&models.DeliveryAttempt{
                NotificationID: notificationID,
                Channel:        "sms",
                Provider:       "twilio",
                Status:         "delivered",
                Try:            attempt,
                LatencyMs:      latency,
            })
            metrics.ExternalAPISuccessTotal.WithLabelValues("sendgrid", "email_worker").Inc()
            metrics.NotificationsAttemptedTotal.WithLabelValues("sms", "success", "twilio").Inc()
            return nil
        }
        sendSpan.AddEvent(fmt.Sprintf("Retry %d failed", attempt))
		sendSpan.RecordError(err)
        metrics.NotificationRetriesTotal.WithLabelValues("sms", "twilio", "provider_error").Inc()
        metrics.NotificationsAttemptedTotal.WithLabelValues("sms", "failed", "twilio").Inc()
        notificationRepo.CreateAttempt(&models.DeliveryAttempt{
            NotificationID: notificationID,
            Channel:        "sms",
            Provider:       "twilio",
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

    metrics.NotificationsAttemptedTotal.WithLabelValues("sms", "failed", "twilio").Inc()
    metrics.ExternalAPIFailureTotal.WithLabelValues("twilio", "sms_worker").Inc()
    notificationRepo.UpdateStatus(notificationID, "failed")

    smsBytes, marshalErr := json.Marshal(sms)

    _, dlqSpan := tracer.Start(ctx, "publish-dlq")
	defer dlqSpan.End()

    if marshalErr != nil {
        logger.Error("Failed to marshal SMS for DLQ",
            zap.String("to", sms.To),
            zap.Error(marshalErr),
        )
        
        return marshalErr
    }

    err := producer.Publish(context.Background(), "notification.sms.dlq", notificationID[:], smsBytes)
    if err != nil{
        dlqSpan.RecordError(err)
		dlqSpan.SetStatus(codes.Error, err.Error())

    }
	metrics.NotificationDLQTotal.WithLabelValues("provider_error","sms")

    logger.Error("Final SMS send failure",
        zap.String("to", sms.To),
        zap.Error(fmt.Errorf("SendSMS failed after %d retries", maxRetries)),
    )

    return fmt.Errorf("SendSMS failed after %d retries", maxRetries)
}

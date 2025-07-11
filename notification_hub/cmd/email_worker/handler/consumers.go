package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jsndz/signalbus/cmd/email_worker/service"
	"github.com/jsndz/signalbus/pkg/kafka"
	"go.uber.org/zap"
)

type SignupRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
}

type SuccessRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required"`
	Amount   string `json:"amount" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Method   string `json:"method" binding:"required"`
}

func SignupConsumer(broker string, ctx context.Context, mailService *service.MailClient, logger *zap.Logger) {
	topic := "user_signup"
	c := kafka.NewConsumer(topic, []string{broker})
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

			var messageBody SignupRequest
			err = json.Unmarshal(m.Value, &messageBody)
			if err != nil {
				logger.Error("Failed to unmarshal signup message",
					zap.ByteString("raw", m.Value),
					zap.Error(err),
				)
				continue
			}

			logger.Info("Sending welcome email",
				zap.String("to_email", messageBody.Email),
				zap.String("username", messageBody.Username),
			)

			mailService.SendMailWithRetry(service.SendEmailRequest{
				ToName:    messageBody.Username,
				ToEmail:   messageBody.Email,
				Subject:   "Welcome to SignalBus!",
				PlainText: "Thank you for signing up. We're excited to have you on board.",
				HTMLBody:  "<h1>Welcome!</h1><p>Thank you for signing up. We're excited to have you on board.</p>",
			})
		}
	}
}

func PaymentSuccessConsumer(broker string, ctx context.Context, mailService *service.MailClient, logger *zap.Logger) {
	topic := "payment_success"
	c := kafka.NewConsumer(topic, []string{broker})
	defer c.Close()

	logger.Info("Starting Kafka consumer", zap.String("topic", topic), zap.String("broker", broker))

	for {
		select {
		case <-ctx.Done():
			logger.Info("Shutting down PaymentSuccessConsumer", zap.String("topic", topic))
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

			var messageBody SuccessRequest
			err = json.Unmarshal(m.Value, &messageBody)
			if err != nil {
				logger.Error("Failed to unmarshal payment message",
					zap.ByteString("raw", m.Value),
					zap.Error(err),
				)
				continue
			}

			logger.Info("Sending payment confirmation email",
				zap.String("to_email", messageBody.Email),
				zap.String("amount", messageBody.Amount),
				zap.String("method", messageBody.Method),
			)

			mailService.SendMail(service.SendEmailRequest{
				ToName:    messageBody.Username,
				ToEmail:   messageBody.Email,
				Subject:   "Your payment was successful!",
				PlainText: fmt.Sprintf("Hello, we have received your payment of %s via %s.", messageBody.Amount, messageBody.Method),
				HTMLBody:  fmt.Sprintf("<p>Hello, we have received your payment of <strong>%s</strong> via <em>%s</em>.</p>", messageBody.Amount, messageBody.Method),
			})
		}
	}
}

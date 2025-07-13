package service

import (
	"fmt"
	"time"

	"github.com/jsndz/signalbus/metrics"
	"github.com/jsndz/signalbus/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.uber.org/zap"
)

const maxRetries = 3

type MailClient struct {
	FromMail string
	Client   *sendgrid.Client
	Logger   *zap.Logger
}

type SendEmailRequest struct {
	ToName    string
	ToEmail   string
	Subject   string
	PlainText string
	HTMLBody  string
}

func NewMailClient(logger *zap.Logger) *MailClient {
	return &MailClient{
		FromMail: utils.GetEnv("SENDGRID_FROM_EMAIL"),
		Client:   sendgrid.NewSendClient(utils.GetEnv("SENDGRID_API_KEY")),
		Logger:   logger,
	}
}

func (m *MailClient) SendMail(emailReq SendEmailRequest) error {
	from := mail.NewEmail("SignalBus", m.FromMail)
	to := mail.NewEmail(emailReq.ToName, emailReq.ToEmail)
	message := mail.NewSingleEmail(from, emailReq.Subject, to, emailReq.PlainText, emailReq.HTMLBody)

	m.Logger.Info("Sending email",
		zap.String("to", emailReq.ToEmail),
		zap.String("subject", emailReq.Subject),
	)

	_, err := m.Client.Send(message)
	if err != nil {
		m.Logger.Error("SendGrid email failed",
			zap.String("to", emailReq.ToEmail),
			zap.Error(err),
		)
		metrics.ExternalAPIFailure.WithLabelValues("sendgrid", "email_worker").Inc()
		return err
	}

	m.Logger.Info("Email sent successfully",
		zap.String("to", emailReq.ToEmail),
		zap.String("subject", emailReq.Subject),
	)
	metrics.ExternalAPISuccess.WithLabelValues("sendgrid", "email_worker").Inc()
	return nil
}

func (m *MailClient) SendMailWithRetry(emailReq SendEmailRequest) error {
	timer := prometheus.NewTimer(metrics.NotificationSendDuration.WithLabelValues("sendgrid", "email_worker"))
	defer timer.ObserveDuration()

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := m.SendMail(emailReq)
		if err == nil {
			return nil
		}

		waitTime := time.Duration(attempt*2) * time.Second

		m.Logger.Warn("Email send failed, will retry",
			zap.String("to", emailReq.ToEmail),
			zap.Int("attempt", attempt),
			zap.Duration("retry_in", waitTime),
			zap.Error(err),
		)

		time.Sleep(waitTime)
	}

	err := fmt.Errorf("SendMail failed after %d retries", maxRetries)
	m.Logger.Error("Final email send failure",
		zap.String("to", emailReq.ToEmail),
		zap.Error(err),
	)
	return err
}

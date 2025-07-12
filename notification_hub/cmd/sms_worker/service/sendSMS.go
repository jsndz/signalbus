package service

import (
	"fmt"
	"time"

	"github.com/jsndz/signalbus/metrics"
	"github.com/jsndz/signalbus/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
	"go.uber.org/zap"
)

const maxRetries = 3

type SMSClient struct {
	FromNumber string
	Client     *twilio.RestClient
	Logger     *zap.Logger
}

func NewSMSClient(logger *zap.Logger) *SMSClient {
	accountSid := utils.GetEnv("TWILIO_ACCOUNT_SID")
	authToken := utils.GetEnv("TWILIO_AUTH_TOKEN")
	fromNumber := utils.GetEnv("TWILIO_PHONE_NUMBER")

	client := twilio.NewRestClientWithParams(
		twilio.ClientParams{
			Username: accountSid,
			Password: authToken,
		})

	return &SMSClient{
		FromNumber: fromNumber,
		Client:     client,
		Logger:     logger,
	}
}

func (c *SMSClient) SendSMS(toNumber string) error {
	body := "Welcome to SignalBus"
	c.Logger.Info("Sending SMS",
		zap.String("to", toNumber),
		zap.String("body", body),
	)

	params := &api.CreateMessageParams{}
	params.SetBody(body)
	params.SetFrom(c.FromNumber)
	params.SetTo(toNumber)

	resp, err := c.Client.Api.CreateMessage(params)
	if err != nil {
		c.Logger.Error("Twilio API error while sending SMS",
			zap.String("to", toNumber),
			zap.Error(err),
		)
		metrics.ExternalAPIFailure.WithLabelValues("twilio", "sms_worker").Inc()
		return err
	}

	if resp.Sid != nil {
		c.Logger.Info("SMS sent successfully",
			zap.String("to", toNumber),
			zap.String("sid", *resp.Sid),
		)
		metrics.ExternalAPISuccess.WithLabelValues("twilio", "sms_worker").Inc()
	} else {
		c.Logger.Error("SMS send failed: no SID returned",
			zap.String("to", toNumber),
		)
		metrics.ExternalAPIFailure.WithLabelValues("twilio", "sms_worker").Inc()
	}
	return nil
}

func (c *SMSClient) SendSMSWithRetry(toNumber string) error {
	timer := prometheus.NewTimer(metrics.NotificationSendDuration.WithLabelValues("twilio", "sms_worker"))
	defer timer.ObserveDuration()

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := c.SendSMS(toNumber)
		if err == nil {
			return nil
		}

		waitTime := time.Duration(attempt*2) * time.Second
		c.Logger.Warn("SMS send attempt failed, will retry",
			zap.String("to", toNumber),
			zap.Int("attempt", attempt),
			zap.Error(err),
			zap.Duration("retry_in", waitTime),
		)

		time.Sleep(waitTime)
	}

	finalErr := fmt.Errorf("failed to send SMS after %d retries", maxRetries)
	c.Logger.Error("Final SMS send failure",
		zap.String("to", toNumber),
		zap.Error(finalErr),
	)
	return finalErr
}

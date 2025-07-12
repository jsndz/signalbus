package service

import (
	"fmt"
	"log"
	"time"

	"github.com/jsndz/signalbus/metrics"
	"github.com/jsndz/signalbus/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
)
const maxRetries = 3

type SMSClient struct {
	FromNumber string
	Client *twilio.RestClient
}

func NewSMSClient() *SMSClient {
	accountSid := utils.GetEnv("TWILIO_ACCOUNT_SID")
	authToken := utils.GetEnv("TWILIO_AUTH_TOKEN")
	fromNumber := utils.GetEnv("TWILIO_PHONE_NUMBER") 
	client:= twilio.NewRestClientWithParams(
		twilio.ClientParams{
			Username: accountSid,
			Password: authToken,
		})
	return &SMSClient{
		FromNumber: fromNumber,
		Client: client,
		
	}
}



func (c *SMSClient) SendSMS(toNumber string) error {
	params := &api.CreateMessageParams{}
	body:= "Welcome to SignalBus"
	params.SetBody(body)
	params.SetFrom(c.FromNumber)
	params.SetTo(toNumber)

	resp, err := c.Client.Api.CreateMessage(params)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	if resp.Sid != nil {
		log.Printf(" SMS sent to %s! Message SID: %s\n", toNumber, *resp.Sid)	
		metrics.ExternalAPISuccess.WithLabelValues("twilio","sms_worker").Inc()
	} else {
		log.Println(resp.Body)
		metrics.ExternalAPIFailure.WithLabelValues("twilio","sms_worker").Inc()
	}
	return nil
}


func(c *SMSClient) SendSMSWithRetry(toNumber string ) error{
	timer := prometheus.NewTimer(metrics.NotificationSendDuration.WithLabelValues("twilio", "sms_worker"))
	defer timer.ObserveDuration()
	for attempt:=1; attempt<=maxRetries;attempt++{
		err := c.SendSMS(toNumber)
		if err == nil {
			return nil 
		}
		log.Printf("Send attempt %d failed: %v", attempt, err)
		waitTime := time.Duration(attempt*2) * time.Second
		log.Printf("Retrying in %v...", waitTime)
		time.Sleep(waitTime)
	}
	err := fmt.Errorf("failed to send SMS")
	log.Printf("SendMail failed after %d retries: %v", maxRetries, err)
	return err
}

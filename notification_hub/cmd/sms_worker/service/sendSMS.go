package service

import (
	"log"

	"github.com/jsndz/signalbus/pkg/utils"
	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
)

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



func (c *SMSClient)SendSMS(toNumber string) {
	params := &api.CreateMessageParams{}
	body:= "Welcome to SignalBus"
	params.SetBody(body)
	params.SetFrom(c.FromNumber)
	params.SetTo(toNumber)

	resp, err := c.Client.Api.CreateMessage(params)
	if err != nil {
		log.Println(err.Error())
		return
	}
	if resp.Sid != nil {
		log.Printf(" SMS sent to %s! Message SID: %s\n", toNumber, *resp.Sid)	
	} else {
		log.Println(resp.Body)
	}
	
}
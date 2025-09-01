package gosms

import (
	"context"
	"time"

	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
)

type TwilioSender struct{		
	FromNumber        string          `yaml:"fromNumber"`
    Timeout         time.Duration     `yaml:"timeout"`
    IdempotencyKey  string            `yaml:"idempotencyKey"`
    Headers         map[string]string `yaml:"headers,omitempty"`
	Client     *twilio.RestClient
	Ctx context.Context
}

func NewTwilioSender(accountSid,authToken,fromNumber string) *TwilioSender {
	client := twilio.NewRestClientWithParams(
	twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})

	return &TwilioSender{
		FromNumber:fromNumber,
		Client: client,
		Timeout:  10 * time.Second,
	}
}

func (t *TwilioSender) Send(s SMS) error {

	params := &api.CreateMessageParams{}
	params.SetBody(s.Text)
	params.SetFrom(t.FromNumber)
	params.SetTo(s.To)

	resp, err := t.Client.Api.CreateMessage(params)
	if err != nil {
		return err
	}
	if resp.Sid != nil {
		return nil
	}
	return nil
}
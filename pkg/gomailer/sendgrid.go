package gomailer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jsndz/signalbus/pkg/types"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct{
	Provider        string
	APIKey          string            `yaml:"apiKey"`
    BaseURL         string            `yaml:"baseURL"`
    Timeout         time.Duration     `yaml:"timeout"`
    FromName        string            `yaml:"fromName"`
    FromMail        string            `yaml:"fromMail"`
    IdempotencyKey  string            `yaml:"idempotencyKey"`
    Headers         map[string]string `yaml:"headers,omitempty"`
	Client   *sendgrid.Client
	Ctx context.Context
}


func NewSendGridMailer(apiKey, fromName, fromMail string) *SendGridMailer {
	return &SendGridMailer{
		APIKey:   apiKey,
		FromName: fromName,
		FromMail: fromMail,
		Timeout:  10 * time.Second,
		Client:   sendgrid.NewSendClient(apiKey),
	}
}

func (s *SendGridMailer) Send(e Email) (*types.SendResponse,error) {
	from := mail.NewEmail(s.FromName, e.From)

	var recipients []*mail.Email
	for _, email := range e.To {
		recipients = append(recipients, mail.NewEmail("", email))
	}

	message := mail.NewV3Mail()
	message.SetFrom(from)
	message.Subject = e.Subject

	p := mail.NewPersonalization()
	for _, r := range recipients {
		p.AddTos(r)
	}
	message.AddPersonalizations(p)

	if e.Text != "" {
		message.AddContent(mail.NewContent("text/plain", e.Text))
	}
	if e.HTML != "" {
		message.AddContent(mail.NewContent("text/html", e.HTML))
	}
	body := mail.GetRequestBody(message)
	request,err := http.NewRequestWithContext(s.Ctx, "POST", "https://api.sendgrid.com/v3/mail/send",bytes.NewReader(body))
	if err != nil {
		return nil,err
	}
	request.Header.Set("Authorization", "Bearer "+s.APIKey)
	request.Header.Set("Content-Type", "application/json")
	for k,v := range e.Headers{
		request.Header.Set(k, v)
	}
	if e.IdempotencyKey != "" {
		request.Header.Set("Idempotency-Key", e.IdempotencyKey)
	}
	client := &http.Client{Timeout: s.Timeout}
	resp,err:=client.Do(request)
	if err != nil {
		return nil,fmt.Errorf("sendgrid send error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil,fmt.Errorf("sendgrid API error: %d %s", resp.StatusCode, resp.Body)
	}
	bodyBytes, _ := io.ReadAll(resp.Body)

	res := &types.SendResponse{
		Provider: "sendgrid",
		ProviderID: "",
		Status:     "accepted",
		RawResponse: bodyBytes,
		Timestamp:  time.Now(),
	}
	return res,nil
}

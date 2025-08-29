package gomailer

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct{
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

type MailClient struct {
	FromMail string
	Client   *sendgrid.Client
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

func (s *SendGridMailer) Send(e Email) error {
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
		return err
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
		return fmt.Errorf("sendgrid send error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("sendgrid API error: %d %s", resp.StatusCode, resp.Body)
	}

	return nil
}

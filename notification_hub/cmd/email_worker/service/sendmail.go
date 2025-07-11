package service

import (
	"fmt"
	"log"
	"time"

	"github.com/jsndz/signalbus/pkg/utils"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

const maxRetries = 3

type MailClient struct {
	FromMail string
	Client *sendgrid.Client
}

type SendEmailRequest struct {
	ToName    string
	ToEmail   string
	Subject   string
	PlainText string
	HTMLBody  string
}

func NewMailClient() *MailClient {
	return &MailClient{
		FromMail: utils.GetEnv("SENDGRID_FROM_EMAIL"),
		Client:   sendgrid.NewSendClient(utils.GetEnv("SENDGRID_API_KEY")),
	}
}


func(m *MailClient) SendMail(emailReq SendEmailRequest) error{
	from := mail.NewEmail("SignalBus", m.FromMail)
	to := mail.NewEmail(emailReq.ToName, emailReq.ToEmail)
	message := mail.NewSingleEmail(
		from,
		emailReq.Subject,
		to,
		emailReq.PlainText,
		emailReq.HTMLBody,
	)
	response, err := m.Client.Send(message)
	if err != nil {
		log.Printf("SendGrid error: %v", err)
		return err
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
	return nil
}

func(m *MailClient) SendMailWithRetry(emailReq SendEmailRequest) error{

	for attempt:=1; attempt<=maxRetries;attempt++{
		err := m.SendMail(emailReq)
		if err == nil {
			return nil 
		}
		log.Printf("Send attempt %d failed: %v", attempt, err)
		waitTime := time.Duration(attempt*2) * time.Second
		log.Printf("Retrying in %v...", waitTime)
		time.Sleep(waitTime)
	}
	err := fmt.Errorf("failed to send email")
	log.Printf("SendMail failed after %d retries: %v", maxRetries, err)
	return err
}

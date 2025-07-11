package service

import (
	"fmt"
	"log"

	"github.com/jsndz/signalbus/pkg/utils"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type MailClient struct {
	FromMail string
	Client *sendgrid.Client
}

func NewMailClient() *MailClient {
	return &MailClient{
		FromMail: utils.GetEnv("SENDGRID_FROM_EMAIL"),
		Client:   sendgrid.NewSendClient(utils.GetEnv("SENDGRID_API_KEY")),
	}
}


func(m *MailClient) SendMail(username string, email string) {
	from := mail.NewEmail("SignalBus", m.FromMail)
	subject := "Welcome to SIGNALBUS*"
	to := mail.NewEmail(username, email)
	plainTextContent := "and easy to do anywhere, even with Go"
	htmlContent := "<strong>and easy to do anywhere, even with Go</strong>"
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)

	response, err := m.Client.Send(message)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
}
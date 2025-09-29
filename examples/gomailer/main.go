package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jsndz/signalbus/pkg/gomailer"
)

func main() {
	mailer := gomailer.NewSendGridMailer(
		"SG.fake_api_key",       
		"My App",                 
		"noreply@example.com",    
	)

	mailer.Timeout = 5 * time.Second
	mailer.Ctx = context.Background()

	email := gomailer.Email{
		From:    "noreply@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Welcome to GoMailer",
		Text:    "This is a plain text body",
		HTML:    "<h1>Hello from GoMailer!</h1><p>This is an HTML body.</p>",
		Headers: map[string]string{
			"X-Custom-Header": "GoMailerExample",
		},
		IdempotencyKey: "example-12345",
	}

	if _,err := mailer.Send(email); err != nil {
		log.Fatalf("failed to send email: %v", err)
	}

	fmt.Println(" Email sent successfully!")
}

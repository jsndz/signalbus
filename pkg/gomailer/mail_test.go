package gomailer_test

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/jsndz/signalbus/pkg/gomailer"
)

type MailpitAddress struct {
	Name    string `json:"Name"`
	Address string `json:"Address"`
}

type MailpitMessage struct {
	ID        string          `json:"ID"`
	MessageID string          `json:"MessageID"`
	Read      bool            `json:"Read"`
	From      MailpitAddress  `json:"From"`
	To        []MailpitAddress `json:"To"`
	Cc        []MailpitAddress `json:"Cc"`
	Bcc       []MailpitAddress `json:"Bcc"`
	ReplyTo   []MailpitAddress `json:"ReplyTo"`
	Subject   string          `json:"Subject"`
	Created   time.Time       `json:"Created"`
	Username  string          `json:"Username"`
	Tags      []string        `json:"Tags"`
	Size      int             `json:"Size"`
	Attachments int           `json:"Attachments"`
	Snippet   string          `json:"Snippet"`
}

type MailpitMessagesResponse struct {
	Total          int              `json:"total"`
	Unread         int              `json:"unread"`
	Count          int              `json:"count"`
	MessagesCount  int              `json:"messages_count"`
	MessagesUnread int              `json:"messages_unread"`
	Start          int              `json:"start"`
	Tags           []string         `json:"tags"`
	Messages       []MailpitMessage `json:"messages"`
}

func GetFromMailPit(t *testing.T)[]MailpitMessage {
	resp, err := http.Get("http://localhost:8025/api/v1/messages")
	if err != nil {
		t.Fatalf("failed to fetch messages: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var data struct {
		Messages []MailpitMessage `json:"messages"`
	}
	log.Println(string(body))
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("failed to parse mailpit response: %v", err)
	}
	return data.Messages
}

func TestSendMail(t *testing.T)  {
	mailer := &gomailer.SMTPMailer{
		Host: "localhost",
		Port: 1025,
		UseAuth: false,
		Timeout: 5 * time.Second,
	}

	email := gomailer.NewEmail(
		"test@example.com",
		[]string{"receiver@example.com"},
		gomailer.WithSubject("Hello Mailpit"),
		gomailer.WithText("This is a test email"),
	)
	if err := mailer.Send(email); err != nil {
		t.Fatalf("failed to send email: %v", err)
	}
	time.Sleep(200 * time.Millisecond)
	messages := GetFromMailPit(t)
	found := false
	for _,msg := range messages{
		if msg.Subject ==  "Hello Mailpit"{
			found =  true
			break
		}
	}

	if !found {
		t.Errorf("expected email with subject %q not found in Mailpit", "Hello Mailpit")
	} else {
		fmt.Println(" got Email successful")
	}

}
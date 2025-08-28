package gomailer

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SMTPMailer struct{
	Host     string
	Port     int
	Username string
	Password string
	UseAuth  bool
	TLSConfig *tls.Config
	Timeout time.Duration
	Headers map[string]string
	Ctx context.Context
	IdempotencyKey string
}


func (m *SMTPMailer) tlsConfig()*tls.Config{
	if m.TLSConfig !=nil{
		return m.TLSConfig
	}
	if m.Host == "localhost"{
		return &tls.Config{
			InsecureSkipVerify: true,
			ServerName: m.Host,
		}
	}
	return &tls.Config{
		ServerName: m.Host,
	}
}

func (m *SMTPMailer) Send (email Email) error{
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("From: %s\r\n", email.From))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ",")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))
	msg.WriteString("MIME-Version: 1.0\r\n")

	
	boundary := "SIGNALBUS_BOUNDARY"
	if len(email.Attachments) > 0 {
		msg.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary))
		msg.WriteString("\r\n")
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		if email.HTML != "" {
			msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
			msg.WriteString(email.HTML + "\r\n")
		} else {
			msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
			msg.WriteString(email.Text + "\r\n")
		}

		for _, path := range email.Attachments {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			encoded := base64.StdEncoding.EncodeToString(data)
			filename := filepath.Base(path)

			msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			msg.WriteString(fmt.Sprintf("Content-Type: application/octet-stream; name=\"%s\"\r\n", filename))
			msg.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", filename))
			msg.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")

			for i := 0; i < len(encoded); i += 76 {
				end := i + 76
				if end > len(encoded) {
					end = len(encoded)
				}
				msg.WriteString(encoded[i:end] + "\r\n")
			}
		}

		msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		if email.HTML != "" {
			msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
			msg.WriteString(email.HTML)
		} else {
			msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
			msg.WriteString(email.Text)
		}
	}
	smtpAddr := fmt.Sprintf("%s:%d",m.Host,m.Port)
	var auth smtp.Auth

	if m.UseAuth {
		auth =smtp.PlainAuth("",m.Username,m.Password,m.Host)
	}
	dialer := net.Dialer{Timeout: m.Timeout}
	conn, err := dialer.DialContext(m.Ctx,"tcp",smtpAddr)
	if err != nil {
		return err
	}

	var client *smtp.Client
	if m.Port == 465 {
		tlsConn := tls.Client(conn, m.tlsConfig())
		client, err = smtp.NewClient(tlsConn, m.Host)
	} else {
		client, err = smtp.NewClient(conn, m.Host)
		if err == nil {
			if ok, _ := client.Extension("STARTTLS"); ok {
				if err = client.StartTLS(m.tlsConfig()); err != nil {
					return err
				}
			}
		}
	}
	if err != nil {
		return err
	}
	defer client.Close()

	if m.UseAuth {
		if ok, _ := client.Extension("AUTH"); ok {
			if err = client.Auth(auth); err != nil {
				return err
			}
		}
	}

	if err = client.Mail(email.From); err != nil {
		return err
	}
	for _, recipient := range email.To {
		if err = client.Rcpt(recipient); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write([]byte(msg.String())); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}

	return client.Quit()
}
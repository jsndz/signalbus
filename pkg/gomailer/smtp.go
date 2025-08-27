package gomailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
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

func (m *SMTPMailer)Send(email Email) error{
	headers := make(map[string]string)
	headers["From"] = email.From
	headers["To"] = strings.Join(email.To,",")
	headers["Subject"] = email.Subject
	headers["MIME-Version"] = "1.0"
	if email.HTML != ""{
		headers["Content-Type"]= "text/html; charset=\"UTF-8\""
	} else{
		headers["Content-Type"]= "text/plain; charset=\"UTF-8\""
	}
	var msg strings.Builder

	for k,v :=range headers{
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	if email.HTML != "" {
		msg.WriteString("\r\n" + email.HTML)
	} else {
		msg.WriteString("\r\n" + email.Text)
	}

	smtpAddr := fmt.Sprintf("%s:%d",m.Host,m.Port)
	var auth smtp.Auth

	if m.UseAuth {
		auth =smtp.PlainAuth("",m.Username,m.Password,m.Host)
	}
	if m.Port == 465{
		conn, err := tls.Dial("tcp",smtpAddr,m.tlsConfig())

		if err!=nil{
			return err
		}

		c,err := smtp.NewClient(conn,m.Host)
		if err!=nil{
			return err
		}
		defer c.Close()

		if err= c.Auth(auth);  err!=nil{
			return err 
		}
		if err= c.Mail(m.Username);  err!=nil{
			return err 
		}
		 for _, recipient := range email.To {
            if err = c.Rcpt(recipient); err != nil {
                return err
            }
        }
		w , err := c.Data()
		if err!=nil{
			return err
		}
		_,err =w.Write([]byte(msg.String()))
		if err!=nil{
			return err
		}
		return w.Close()
	}
	return smtp.SendMail(smtpAddr,auth,email.From,email.To,[]byte(msg.String()))
}
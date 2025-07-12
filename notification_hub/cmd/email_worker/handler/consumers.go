package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jsndz/signalbus/cmd/email_worker/service"
	"github.com/jsndz/signalbus/pkg/kafka"
)
type SignupRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Phone string `json:"phone" binding:"required"`
}

type SuccessRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email string `json:"email" binding:"required"`
	Amount string `json:"amount" binding:"required"`
	Phone string `json:"phone" binding:"required"`
	Method string `json:"method" binding:"required"`
}


func SignupConsumer(broker string,ctx context.Context,mailService *service.MailClient){
	c:=kafka.NewConsumer("user_signup", []string{broker})
	defer c.Close()
	for{
		select{
		case <-ctx.Done():
			log.Println("COnsumer is shutting down")
			return
		default:
		m,err:=c.ReadFromKafka(ctx)
		if err!=nil{
			log.Print(err)
		}
		var	messageBody SignupRequest
		err = json.Unmarshal(m.Value, &messageBody)
		if err!=nil{
			log.Println("Couldn't Parse JSON")
		}
		mailService.SendMailWithRetry(service.SendEmailRequest{
			ToName:    messageBody.Username,
			ToEmail:   messageBody.Email,
			Subject:   "Welcome to SignalBus!",
			PlainText: "Thank you for signing up. We're excited to have you on board.",
			HTMLBody:  "<h1>Welcome!</h1><p>Thank you for signing up. We're excited to have you on board.</p>",
		})	}	
}
}


func PaymentSuccessConsumer(broker string,ctx context.Context,mailService *service.MailClient){
	
	c:=kafka.NewConsumer("payment_success",[]string{broker})
	defer c.Close()

	for{
		select{
		case <-ctx.Done():
			log.Println("COnsumer is shutting down")
			return
		default:
			m,err:=c.ReadFromKafka( ctx)
			if err!=nil{
				log.Print(err)
			}
			var	messageBody SuccessRequest
			err = json.Unmarshal(m.Value, &messageBody)
			if err!=nil{
				log.Println("COuldn't Parse JSON")
				continue
			}
			mailService.SendMail(service.SendEmailRequest{
				ToName:     messageBody.Username,
				ToEmail:   messageBody.Email, 
				Subject:   "Your payment was successful!",
				PlainText: fmt.Sprintf("Hello, we have received your payment of %s via %s.", messageBody.Amount, messageBody.Method),
				HTMLBody:  fmt.Sprintf("<p>Hello, we have received your payment of <strong>%s</strong> via <em>%s</em>.</p>", messageBody.Amount, messageBody.Method),
			})	
		}
		
	}
}
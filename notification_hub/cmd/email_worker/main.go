package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/cmd/email_worker/handler"
	"github.com/jsndz/signalbus/cmd/email_worker/service"
	"github.com/jsndz/signalbus/pkg/utils"
)



func main(){
	router:= gin.Default()
	broker:= utils.GetEnv("KAFKA_BROKER")
	log.Println(broker)
	ctx, cancel := context.WithCancel(context.Background())
	mailService:= service.NewMailClient()
	defer cancel()
	go handler.SignupConsumer(broker,ctx,mailService)
	go handler.PaymentSuccessConsumer(broker,ctx,mailService)

	if err:= router.Run(":3001");err !=nil{
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
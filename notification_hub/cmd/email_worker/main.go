package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/cmd/email_worker/service"
	"github.com/jsndz/signalbus/pkg/kafka"
	"github.com/jsndz/signalbus/pkg/utils"
)

func main(){
	router:= gin.Default()
	broker:= utils.GetEnv("KAFKA_BROKER")
	log.Println(broker)
	c:=kafka.NewConsumer("user_signup", []string{broker})
	ctx, cancel := context.WithCancel(context.Background())
	mailService:= service.NewMailClient()
	defer cancel()
	defer c.Close() 
	go func(){
		for{
			m,err:=c.ReadFromKafka(ctx)
			if err!=nil{
				log.Print(err)
			}
			mailService.SendMail(string(m.Key), string(m.Value))
		}
	}()

	if err:= router.Run(":3000");err !=nil{
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
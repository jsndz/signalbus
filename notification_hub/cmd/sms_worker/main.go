package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/cmd/sms_worker/service"
	"github.com/jsndz/signalbus/metrics"
	"github.com/jsndz/signalbus/middlewares"
	"github.com/jsndz/signalbus/pkg/kafka"
	"github.com/jsndz/signalbus/pkg/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type SignupRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Phone string `json:"phone" binding:"required"`
}

func main(){
	router:= gin.Default()
	metrics.InitSMSMetrics()
	router.Use(middlewares.MetricsMiddleware())
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/health",func(ctx *gin.Context) {
		ctx.JSON(http.StatusAccepted,gin.H{"message":"ok"})
	 })
	broker:= utils.GetEnv("KAFKA_BROKER")
	log.Println(broker)
	c:=kafka.NewConsumer("user_signup", []string{broker})
	ctx, cancel := context.WithCancel(context.Background())
	smsService:= service.NewSMSClient()
	defer cancel()
	defer c.Close() 
	go func(){
		for{
			m,err:=c.ReadFromKafka(ctx)
			if err!=nil{
				log.Print(err)
			}
			var	messageBody SignupRequest
			err = json.Unmarshal(m.Value, &messageBody)
			if err!=nil{
				log.Println("COuldn't Parse JSON")
			}
			smsService.SendSMS(messageBody.Phone)
		}
	}()

	if err:= router.Run(":3000");err !=nil{
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
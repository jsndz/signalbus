package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/cmd/email_worker/handler"
	"github.com/jsndz/signalbus/cmd/email_worker/service"
	"github.com/jsndz/signalbus/metrics"
	"github.com/jsndz/signalbus/middlewares"
	"github.com/jsndz/signalbus/pkg/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)



func main(){
	router:= gin.Default()
	broker:= utils.GetEnv("KAFKA_BROKER")
	metrics.InitEmailMetrics()
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/health",func(ctx *gin.Context) {
		ctx.JSON(http.StatusAccepted,gin.H{"message":"ok"})
	 })
	router.Use(middlewares.MetricsMiddleware())
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
package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/cmd/email_worker/handler"
	"github.com/jsndz/signalbus/cmd/email_worker/service"
	"github.com/jsndz/signalbus/logger"
	"github.com/jsndz/signalbus/metrics"
	"github.com/jsndz/signalbus/middlewares"
	"github.com/jsndz/signalbus/pkg/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	log, err := logger.InitLogger()
	if err != nil {
		panic("failed to initialize zap logger: " + err.Error())
	}
	defer log.Sync()

	log.Info("Starting email worker service")

	broker := utils.GetEnv("KAFKA_BROKER")
	log.Info("Kafka broker resolved", zap.String("broker", broker))

	metrics.InitEmailMetrics()

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middlewares.MetricsMiddleware())

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/health", func(ctx *gin.Context) {
		log.Info("Health check hit")
		ctx.JSON(http.StatusAccepted, gin.H{"message": "ok"})
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mailService := service.NewMailClient(log)
	log.Info("Mail service initialized")

	go handler.SignupConsumer(broker, ctx, mailService, log)
	go handler.PaymentSuccessConsumer(broker, ctx, mailService, log)

	if err := router.Run(":3001"); err != nil {
		log.Fatal("Failed to start HTTP server", zap.Error(err))
	}
}

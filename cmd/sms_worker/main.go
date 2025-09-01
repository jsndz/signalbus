package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/cmd/sms_worker/service"
	"github.com/jsndz/signalbus/logger"
	"github.com/jsndz/signalbus/metrics"
	"github.com/jsndz/signalbus/middlewares"
	"github.com/jsndz/signalbus/pkg/config"
	"github.com/jsndz/signalbus/pkg/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type SignupRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
}

func main() {
	log, err := logger.InitLogger()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer log.Sync()

	log.Info("Starting SMS worker")

	broker := utils.GetEnv("KAFKA_BROKER")
	log.Info("Kafka broker loaded", zap.String("broker", broker))

	metrics.InitSMSMetrics()
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
	cfg ,err := config.LoadConfig("./config.yaml")
	if err!=nil {
		log.Fatal(err.Error(), zap.Error(err))
	}
	Sender,err := config.BuildSender(cfg)
	if err!=nil {
		log.Fatal(err.Error(), zap.Error(err))
	}
	log.Info("Mail service initialized")
	go service.HandleSMS(broker, ctx, Sender, log)


	if err := router.Run(":3000"); err != nil {
		log.Fatal("Failed to start HTTP server", zap.Error(err))
	}
}

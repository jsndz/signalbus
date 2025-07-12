package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/cmd/sms_worker/service"
	"github.com/jsndz/signalbus/logger"
	"github.com/jsndz/signalbus/metrics"
	"github.com/jsndz/signalbus/middlewares"
	"github.com/jsndz/signalbus/pkg/kafka"
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

	topic := "user_signup"
	consumer := kafka.NewConsumer(topic, []string{broker})
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	smsService := service.NewSMSClient(log)

	go func() {
		log.Info("Started Kafka consumer", zap.String("topic", topic))
		for {
			msg, err := consumer.ReadFromKafka(ctx)
			if err != nil {
				log.Error("Error reading from Kafka", zap.Error(err))
				continue
			}

			log.Info("Kafka message received",
				zap.String("topic", topic),
				zap.ByteString("key", msg.Key),
				zap.Int64("offset", msg.Offset),
			)

			var payload SignupRequest
			if err := json.Unmarshal(msg.Value, &payload); err != nil {
				log.Error("Failed to parse Kafka message",
					zap.ByteString("raw", msg.Value),
					zap.Error(err),
				)
				continue
			}

			log.Info("Sending SMS", zap.String("phone", payload.Phone))
			smsService.SendSMS(payload.Phone)
		}
	}()

	if err := router.Run(":3000"); err != nil {
		log.Fatal("Failed to start HTTP server", zap.Error(err))
	}
}

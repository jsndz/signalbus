package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	api "github.com/jsndz/signalbus/cmd/api/internal"
	"github.com/jsndz/signalbus/logger"
	"github.com/jsndz/signalbus/metrics"
	"github.com/jsndz/signalbus/middlewares"
	"github.com/jsndz/signalbus/pkg/kafka"
	"github.com/jsndz/signalbus/pkg/utils"
)

func main() {
	broker := utils.GetEnv("KAFKA_BROKER")

	log, err := logger.InitLogger()
	if err != nil {
		panic("Failed to initialize zap logger: " + err.Error())
	}
	log.Info("Logger initialized")

	metrics.InitAPIMetrics()

	producer := kafka.NewProducer([]string{broker})
	log.Info("Kafka producer initialized", zap.String("broker", broker))

	router := gin.Default()
	router.Use(middlewares.MetricsMiddleware())

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusAccepted, gin.H{"message": "ok"})
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := router.Group("/api")
	api.Notify(v1.Group("/notify"), producer, log)

	go handleShutdown(producer, log)

	if err := router.Run(); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}
}

func handleShutdown(producer *kafka.Producer, log *zap.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Info("Shutdown signal received", zap.String("signal", sig.String()))

	if err := producer.Close(); err != nil {
		log.Error("Error closing Kafka producer", zap.Error(err))
	} else {
		log.Info("Kafka producer closed cleanly")
	}

	os.Exit(0)
}

package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/jsndz/signalbus/cmd/sms_worker/service"
	"github.com/jsndz/signalbus/logger"
	"github.com/jsndz/signalbus/metrics"
	"github.com/jsndz/signalbus/middlewares"
	"github.com/jsndz/signalbus/pkg/config"
	"github.com/jsndz/signalbus/pkg/database"
	"github.com/jsndz/signalbus/pkg/kafka"
	"github.com/jsndz/signalbus/pkg/repositories"
	"github.com/jsndz/signalbus/pkg/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()

	logr, err := logger.InitLogger()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logr.Sync()

	dsn := os.Getenv("TENANT_DB")
	notification_dns := os.Getenv("NOTIFICATION_DB")

	db, err := database.InitDB(dsn)
	if err != nil {
		panic("failed to initialize Database: " + err.Error())
	}
	broker := utils.GetEnv("KAFKA_BROKER")
	logr.Info("Kafka broker loaded", zap.String("broker", broker))
	tmplRepo := repositories.NewTemplateRepository(db)
	producer := kafka.NewProducer([]string{broker})
	notification_db,err := database.InitDB(notification_dns)
	if err != nil {
		panic("failed to initialize Database: " + err.Error())
	}
	notification_repo := repositories.NewNotificationRepository(notification_db)
	logr.Info("Starting SMS worker")


	metrics.InitWorkerMetrics()
	metrics.InitKafkaMetrics()

	cfg, err := config.LoadConfig("./config.yaml")
	if err != nil {
		logr.Fatal("failed to load config", zap.Error(err))
	}
	sender, err := config.BuildSender(cfg)
	if err != nil {
		logr.Fatal("failed to init sender", zap.Error(err))
	}
	logr.Info("SMS sender initialized")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go service.HandleSMS(broker, ctx, sender, logr, tmplRepo,notification_repo,producer)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	wrappedMux := middlewares.MetricsMiddleware(mux)
	go handleShutdown(producer, logr)

	if err := http.ListenAndServe(":3003", wrappedMux); err != nil {
		logr.Fatal("metrics server failed", zap.Error(err))
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

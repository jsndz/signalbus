package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/jsndz/signalbus/cmd/email_worker/handler"
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
		panic("failed to initialize zap logger: " + err.Error())
	}
	defer logr.Sync()
	dns := os.Getenv("TEMPLATE_DB")
	db,err := database.InitDB(dns)
	if err != nil {
		panic("failed to initialize Database: " + err.Error())
	}
	tmpl_repo := repositories.NewTemplateRepository(db)

	logr.Info("Starting email worker service")

	broker := utils.GetEnv("KAFKA_BROKER")
	logr.Info("Kafka broker resolved", zap.String("broker", broker))
	producer := kafka.NewProducer([]string{broker})

	metrics.InitEmailMetrics()

	mux :=  http.NewServeMux()
	
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg ,err := config.LoadConfig("./config.yaml")
	
	if err!=nil {
		logr.Fatal(err.Error(), zap.Error(err))
	}
	Mailer,err := config.BuildMailer(cfg)
	if err!=nil {
		logr.Fatal(err.Error(), zap.Error(err))
	}
	logr.Info("Mail service initialized")

	go handler.HandleMail(broker, ctx, Mailer, logr, tmpl_repo,producer)
	wrappedMux := middlewares.MetricsMiddleware(mux)
	go handleShutdown(producer, logr)

	if err := http.ListenAndServe(":3001", wrappedMux); err != nil {
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

package main

import (
	"context"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/jsndz/signalbus/cmd/sms_worker/service"
	"github.com/jsndz/signalbus/logger"
	"github.com/jsndz/signalbus/metrics"
	"github.com/jsndz/signalbus/middlewares"
	"github.com/jsndz/signalbus/pkg/config"
	"github.com/jsndz/signalbus/pkg/database"
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
	db, err := database.InitDB(dsn)
	if err != nil {
		panic("failed to initialize Database: " + err.Error())
	}
	tmplRepo := repositories.NewTemplateRepository(db)

	logr.Info("Starting SMS worker")

	broker := utils.GetEnv("KAFKA_BROKER")
	logr.Info("Kafka broker loaded", zap.String("broker", broker))

	metrics.InitSMSMetrics()

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
	go service.HandleSMS(broker, ctx, sender, logr, tmplRepo)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	wrappedMux := middlewares.MetricsMiddleware(mux)

	if err := http.ListenAndServe(":3001", wrappedMux); err != nil {
		logr.Fatal("metrics server failed", zap.Error(err))
	}
}

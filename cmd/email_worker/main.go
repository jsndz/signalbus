package main

import (
	"context"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/jsndz/signalbus/cmd/email_worker/handler"
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
		panic("failed to initialize zap logger: " + err.Error())
	}
	defer logr.Sync()
	dns := os.Getenv("TENANT_DB")
	db,err := database.InitDB(dns)
	if err != nil {
		panic("failed to initialize Database: " + err.Error())
	}
	tmpl_repo := repositories.NewTemplateRepository(db)

	logr.Info("Starting email worker service")

	broker := utils.GetEnv("KAFKA_BROKER")
	logr.Info("Kafka broker resolved", zap.String("broker", broker))

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

	go handler.HandleMail(broker, ctx, Mailer, logr, tmpl_repo)
	wrappedMux := middlewares.MetricsMiddleware(mux)

	if err := http.ListenAndServe(":3001", wrappedMux); err != nil {
		logr.Fatal("metrics server failed", zap.Error(err))
	}
}

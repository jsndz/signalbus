package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/jsndz/signalbus/cmd/notification_api/app/routes"
	"github.com/jsndz/signalbus/logger"
	"github.com/jsndz/signalbus/metrics"
	"github.com/jsndz/signalbus/middlewares"
	"github.com/jsndz/signalbus/pkg/database"
	"github.com/jsndz/signalbus/pkg/kafka"
	"github.com/jsndz/signalbus/pkg/models"
	"github.com/jsndz/signalbus/pkg/utils"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system env")
	}
	broker := utils.GetEnv("KAFKA_BROKER")
	tenant_dns := os.Getenv("TENANT_DB")
	tenant_db,err := database.InitDB(tenant_dns)
	if err != nil {
		panic("DB not init  " + err.Error())
	}
	template_dns := os.Getenv("TEMPLATE_DB")
	template_db,err := database.InitDB(template_dns)
	database.MigrateDB(template_db, &models.Template{})
	database.MigrateDB(tenant_db, &models.Tenant{},&models.Policy{},&models.APIKey{},&models.IdempotencyKey{})
	if err != nil {
		panic("DB not init  " + err.Error())
	}
	log, err := logger.InitLogger()
	if err != nil {
		panic("Failed to initialize zap logger: " + err.Error())
	}
	log.Info("Logger initialized")

	metrics.InitAPIMetrics()
	producer := kafka.NewProducer([]string{broker})
	log.Info("Kafka producer initialized", zap.String("broker", broker))

	router := gin.Default()
	router.Use(middlewares.GinMetricsMiddleware())

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusAccepted, gin.H{"message": "ok"})
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := router.Group("/api")
	routes.Notifications(v1.Group("/notify"), producer,tenant_db, log)
	routes.Tenants(v1.Group("/tenants"),tenant_db,log)
	routes.APIKeys(v1.Group("/keys"),tenant_db,log)
	routes.Policies(v1.Group("/policies"),tenant_db,log)

	routes.Templates(v1.Group("/templates"),template_db,log)
	go handleShutdown(producer, log)
	if err := router.Run(":3000"); err != nil {
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

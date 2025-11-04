package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/cmd/notification_api/app/internal/handler"
	"github.com/jsndz/signalbus/middlewares"
	"github.com/jsndz/signalbus/pkg/kafka"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Notifications(router *gin.RouterGroup, p *kafka.Producer, db *gorm.DB, redisClient *redis.Client, log *zap.Logger, tracer trace.Tracer) {
	notificationHandler := handler.NewNotificationHandler(db)
	notifyMiddleware := middlewares.MiddlewareConfig{
		RedisClient: redisClient,
		DB:          db,
	}

	router.POST("/", middlewares.NotificationMiddleware(&notifyMiddleware), notificationHandler.Notify(p, db, log, tracer))
	router.POST("/publish", middlewares.NotificationMiddleware(&notifyMiddleware), notificationHandler.Publish(p, db, log))
	router.GET("/:id", notificationHandler.GetNotification(log))
	router.POST("/:id/redrive", notificationHandler.RedriveNotification(log, p))
}

func Templates(r *gin.RouterGroup, db *gorm.DB, log *zap.Logger) {
	templateHandler := handler.NewTemplateHandler(db)
	r.POST("/", templateHandler.CreateTemplate)
	r.GET("/:id", templateHandler.GetTemplateByID)
	r.PUT("/:id", templateHandler.UpdateTemplate)
	r.DELETE("/:id", templateHandler.DeleteTemplate)
}

func Policies(r *gin.RouterGroup, db *gorm.DB, log *zap.Logger) {
	policyHandler := handler.NewPolicyHandler(db)

	r.POST("/", policyHandler.CreatePolicy)
	r.DELETE("/:id", policyHandler.DeletePolicy)
}

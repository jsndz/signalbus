package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/cmd/notification_api/app/internal/services"
	"github.com/jsndz/signalbus/pkg/kafka"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NotifyRequest struct {
    EventType      string                 `json:"event_type" binding:"required"`
    Tenant         string                 `json:"tenant" binding:"required"`
	ApiKey			string					`json:"api_key" binding:"required"`
    Data           map[string]interface{} `json:"data" binding:"required"`
    IdempotencyKey string                 `json:"idempotency_key" binding:"required"`
}

type NotificationMessage struct {
    IdempotencyKey string                 `json:"idempotency_key"`
    Data           map[string]interface{} `json:"data"`
}

func Notify(p *kafka.Producer,db *gorm.DB,log *zap.Logger)gin.HandlerFunc{
	return func(c *gin.Context) {
		s := services.NewTenantService(db)

		log.Info("Incoming HTTP request",
			zap.String("endpoint", "/Notify"),
			zap.String("method", c.Request.Method),
			zap.String("client_ip", c.ClientIP()),
		)
		var req NotifyRequest
		if err:= c.ShouldBindJSON(&req); err!=nil{
			c.JSON(http.StatusAccepted, gin.H{
				"error": "COuldn't unmarshal JSON",
			})
		}
				tenant, err := s.GetTenantByAPIKey(req.ApiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid API key",
			})
			return
		}

		ctx := context.Background()

		msg := NotificationMessage{
			IdempotencyKey: req.IdempotencyKey,
			Data:           req.Data,
		}

		for _, policy := range tenant.Policies {
			for _, channel := range policy.Channels {
				topic := "notification." + channel
				msgBytes, err := json.Marshal(msg)
				if err != nil {
					log.Error("failed to marshal message", zap.Error(err))
					continue
				}
				if err := p.Publish(ctx, topic, []byte(req.EventType), msgBytes); err != nil {
					log.Error("failed to publish message", zap.String("topic", topic), zap.Error(err))
				}
			}
		}
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Notification accepted",
		})
	}
}

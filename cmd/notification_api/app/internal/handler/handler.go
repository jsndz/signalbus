package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jsndz/signalbus/cmd/notification_api/app/internal/services"
	"github.com/jsndz/signalbus/pkg/kafka"
	"github.com/jsndz/signalbus/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NotifyRequest struct {
    EventType      string                 `json:"event_type" binding:"required"`
    Data           map[string]interface{} `json:"data" binding:"required"`
    IdempotencyKey string                 `json:"idempotency_key" binding:"required"`
}

type NotificationMessage struct {
    IdempotencyKey string                 `json:"idempotency_key"`
    Data           map[string]interface{} `json:"data"`
	TenantID	   uuid.UUID				`json:"tenant_id"`
	
}


func Notify(p *kafka.Producer, db *gorm.DB, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := services.NewTenantService(db)
		apiKey := c.GetHeader("X-API-Key")
		log.Info("Incoming HTTP request",
			zap.String("endpoint", "/notify"),
			zap.String("method", c.Request.Method),
			zap.String("client_ip", c.ClientIP()),
		)

		var req NotifyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "couldn't unmarshal JSON: " + err.Error(),
			})
			return
		}

		tenant, err := s.GetTenantByAPIKey(apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid API key",
			})
			return
		}
		var record models.IdempotencyKey
		if err := db.First(&record, "key = ?", req.IdempotencyKey).Error; err == nil {
			c.Data(record.StatusCode, "application/json", []byte(record.Response))
			return
		}
		msg := NotificationMessage{
			IdempotencyKey: req.IdempotencyKey,
			Data:           req.Data,
			TenantID:		tenant.ID,

		}

		msgBytes, err := json.Marshal(msg)
		if err != nil {
			log.Error("failed to marshal message", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		ctx := context.Background()
		for _, policy := range tenant.Policies {
			for _, channel := range policy.Channels {
				topic := "notification." + channel
				if err := p.Publish(ctx, topic, []byte(req.EventType), msgBytes); err != nil {
					log.Error("failed to publish message", zap.String("topic", topic), zap.Error(err))
				}
			}
		}
		resp := gin.H{"message": "Notification accepted"}
		respBytes, _ := json.Marshal(resp)
		db.Create(&models.IdempotencyKey{
			Key:        	req.IdempotencyKey,
			TenantID:  	 	tenant.ID,
			RequestHash: 	string(msgBytes),
			Response:   	string(respBytes),
			StatusCode: 	http.StatusAccepted,
		})
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Notification accepted",
		})
	}
}

func Publish(p *kafka.Producer, db *gorm.DB, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := services.NewTenantService(db)
		apiKey := c.GetHeader("X-API-Key")

		topic := c.Param("topic")
		log.Info("Incoming HTTP request",
			zap.String("endpoint", "/publish/"+topic),
			zap.String("method", c.Request.Method),
			zap.String("client_ip", c.ClientIP()),
		)

		var req NotifyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "couldn't unmarshal JSON: " + err.Error(),
			})
			return
		}

		tenantID, err := s.CheckForValidTenant(apiKey)
		if err != nil ||  tenantID == uuid.Nil{
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid API key",
			})
			return
		}
		var record models.IdempotencyKey
		if err := db.First(&record, "key = ?", req.IdempotencyKey).Error; err == nil {
			c.Data(record.StatusCode, "application/json", []byte(record.Response))
			return
		}
		msg := NotificationMessage{
			IdempotencyKey: req.IdempotencyKey,
			Data:           req.Data,
		}

		msgBytes, err := json.Marshal(msg)
		if err != nil {
			log.Error("failed to marshal message", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		ctx := context.Background()
		if err := p.Publish(ctx, "notification."+topic, []byte(req.EventType), msgBytes); err != nil {
			log.Error("failed to publish message", zap.String("topic", topic), zap.Error(err))
		}
		resp := gin.H{"message": "Notification accepted"}
		respBytes, _ := json.Marshal(resp)

		db.Create(&models.IdempotencyKey{
			Key:        	req.IdempotencyKey,
			TenantID:  	 	tenantID,
			RequestHash: 	string(msgBytes),
			Response:   	string(respBytes),
			StatusCode: 	http.StatusAccepted,
		})
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Notification accepted",
		})
	}
}



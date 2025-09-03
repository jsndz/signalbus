package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/pkg/kafka"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RecieverData struct{
	To string 
	Text string 
}

type NotifyRequest struct {
    EventType      string                 `json:"event_type" binding:"required"`
    Tenant         string                 `json:"tenant" binding:"required"`
    Data           map[string]interface{} `json:"data" binding:"required"`
    IdempotencyKey string                 `json:"idempotency_key" binding:"required"`
}


func Notify(p *kafka.Producer,db *gorm.DB,log *zap.Logger)gin.HandlerFunc{
	return func(c *gin.Context) {
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
		
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Accepted",
		})
	}
}





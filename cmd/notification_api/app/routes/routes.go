package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/cmd/notification_api/app/internal/handler"
	"github.com/jsndz/signalbus/pkg/kafka"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Notify(router *gin.RouterGroup,p *kafka.Producer,db *gorm.DB,log *zap.Logger ){
	router.GET("/notify",handler.Notify(p,db,log))
}
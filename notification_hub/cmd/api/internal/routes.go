package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/pkg/kafka"
	"go.uber.org/zap"
)

func Notify(router *gin.RouterGroup,p *kafka.Producer,log *zap.Logger ){
	router.POST("/test",Testing(log))
	router.POST("/signup",Signup(p,log))
	router.POST("/payment_success",PaymentSuccess(p,log))
}
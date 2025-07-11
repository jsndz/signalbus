package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/pkg/kafka"
)

func Notify(router *gin.RouterGroup,p *kafka.Producer ){
	router.POST("/test",Testing)
	router.POST("/signup",Signup(p))
}
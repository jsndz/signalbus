package api

import (
	"github.com/gin-gonic/gin"
)

func Notify(router *gin.RouterGroup){
	router.POST("/test",Testing)
}
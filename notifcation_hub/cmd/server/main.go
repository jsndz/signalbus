package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/cmd/api"
)

func main() {
  router := gin.Default()
  router.GET("/health",func(ctx *gin.Context) {
     ctx.JSON(http.StatusAccepted,gin.H{"message":"ok"})
     
  })
  v1:=router.Group("/api")
  api.Notify(v1.Group("/notify"))
  router.Run()
}

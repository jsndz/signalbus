package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	api "github.com/jsndz/signalbus/cmd/api/internal"
	"github.com/jsndz/signalbus/kafka"
)

func main() {
  router := gin.Default()
  router.GET("/health",func(ctx *gin.Context) {
     ctx.JSON(http.StatusAccepted,gin.H{"message":"ok"})
  })
  producer:=kafka.NewProducer([]string{"localhost:9092"}, "user_signup")

  v1:=router.Group("/api")
  go func ()  {
    quit:= make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutting down...")
		if err := producer.Close(); err != nil {
			log.Printf("Error closing Kafka producer: %v", err)
		}
		os.Exit(0)
  }()
  api.Notify(v1.Group("/notify"),producer)
  if err:= router.Run();err !=nil{
    fmt.Printf("Failed to start server: %v\n", err)
  }
}

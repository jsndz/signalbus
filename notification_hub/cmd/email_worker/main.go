package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/kafka"
)

func main(){
	router:= gin.Default()
	c:=kafka.NewConsumer("user_signup", []string{"localhost:9092"})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer c.Close() 
	go func(){
		for{
			m,err:=c.ReadFromKafka(ctx)
			if err!=nil{
				log.Print(err)
			}
			log.Println(string(m.Key),string(m.Value))
		}
	}()

	if err:= router.Run(":3000");err !=nil{
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
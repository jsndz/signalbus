package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/pkg/kafka"
)

type TestingRequset struct{
	Name string `json:"name"`
}
type SignupRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Phone string `json:"phone" binding:"required"`
}

type SuccessRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email string `json:"email" binding:"required"`
	Amount string `json:"amount" binding:"required"`
	Phone string `json:"phone" binding:"required"`
	Method string `json:"method" binding:"required"`
}


func Testing(c *gin.Context){
	var req TestingRequset
	if err:= c.ShouldBindJSON(&req) ; err!=nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"message":"Invalid request",
			"err":err.Error(),
		})
		return
	}
	log.Printf("%v",req)

	c.JSON(http.StatusAccepted,gin.H{
		"message":"Accepted",
	})
}


func Signup(producer *kafka.Producer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SignupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
				"err":     err.Error(),
			})
			return
		}
		messageBody,err := json.Marshal(req)
		if err!=nil{
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Couldnt Parse data",
				"error":   err.Error(),
			})
			return		
		}
		err = producer.Publish(context.Background(), "user_signup",[]byte(req.Email),messageBody)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to send Kafka message",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"message": "Signup event sent to Kafka",
		})
	}
}

func PaymentSuccess(producer *kafka.Producer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SuccessRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
				"err":     err.Error(),
			})
			return
		}
		messageBody,err := json.Marshal(req)
		if err!=nil{
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Couldnt Parse data",
				"error":   err.Error(),
			})
			return		
		}
		err = producer.Publish(context.Background(),"payment_success", []byte(req.Email),messageBody)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to send Kafka message",
				"error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Payment is Successful",
		})
	}
}

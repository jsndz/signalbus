package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/pkg/kafka"
	"go.uber.org/zap"
)

type TestingRequset struct {
	Name string `json:"name"`
}

type SignupRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
}

type SuccessRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required"`
	Amount   string `json:"amount" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Method   string `json:"method" binding:"required"`
}

func Testing(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("Incoming HTTP request",
			zap.String("endpoint", "/testing"),
			zap.String("method", c.Request.Method),
			zap.String("client_ip", c.ClientIP()),
		)

		var req TestingRequset
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warn("Invalid request payload",
				zap.String("endpoint", "/testing"),
				zap.Error(err),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
				"err":     err.Error(),
			})
			return
		}

		logger.Info("Testing payload accepted", zap.Any("payload", req))
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Accepted",
		})
	}
}

func Signup(producer *kafka.Producer, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("Incoming HTTP request",
			zap.String("endpoint", "/signup"),
			zap.String("method", c.Request.Method),
			zap.String("client_ip", c.ClientIP()),
		)

		var req SignupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warn("Invalid signup request",
				zap.String("email", req.Email),
				zap.Error(err),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
				"err":     err.Error(),
			})
			return
		}

		messageBody, err := json.Marshal(req)
		if err != nil {
			logger.Error("Failed to marshal signup request",
				zap.String("email", req.Email),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Couldn't parse data",
				"error":   err.Error(),
			})
			return
		}

		err = producer.Publish(context.Background(), "user_signup", []byte(req.Email), messageBody)
		if err != nil {
			logger.Error("Kafka publish failed",
				zap.String("topic", "user_signup"),
				zap.String("email", req.Email),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to send Kafka message",
				"error":   err.Error(),
			})
			return
		}

		logger.Info("Kafka message published",
			zap.String("topic", "user_signup"),
			zap.String("email", req.Email),
		)

		c.JSON(http.StatusAccepted, gin.H{
			"message": "Signup event sent to Kafka",
		})
	}
}

func PaymentSuccess(producer *kafka.Producer, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("Incoming HTTP request",
			zap.String("endpoint", "/payment-success"),
			zap.String("method", c.Request.Method),
			zap.String("client_ip", c.ClientIP()),
		)

		var req SuccessRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warn("Invalid payment success request",
				zap.String("email", req.Email),
				zap.Error(err),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
				"err":     err.Error(),
			})
			return
		}

		messageBody, err := json.Marshal(req)
		if err != nil {
			logger.Error("Failed to marshal payment request",
				zap.String("email", req.Email),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Couldn't parse data",
				"error":   err.Error(),
			})
			return
		}

		err = producer.Publish(context.Background(), "payment_success", []byte(req.Email), messageBody)
		if err != nil {
			logger.Error("Kafka publish failed",
				zap.String("topic", "payment_success"),
				zap.String("email", req.Email),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to send Kafka message",
				"error":   err.Error(),
			})
			return
		}

		logger.Info("Kafka message published",
			zap.String("topic", "payment_success"),
			zap.String("email", req.Email),
			zap.String("method", req.Method),
			zap.String("amount", req.Amount),
		)

		c.JSON(http.StatusAccepted, gin.H{
			"message": "Payment is Successful",
		})
	}
}

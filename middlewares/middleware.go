package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/metrics"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type MiddlewareConfig struct {
	RedisClient *redis.Client
	DB          *gorm.DB

}


func NotificationMiddleware(cfg *MiddlewareConfig) gin.HandlerFunc{
	return func(ctx *gin.Context) {
		shortKey:=fmt.Sprintf("ratelimit:%d",time.Now().Unix()/60)
		count ,_ := cfg.RedisClient.Incr(ctx,shortKey).Result()
		if count == 1 {
			cfg.RedisClient.Expire(ctx, shortKey, time.Minute)
		}
		if count > 10 { 
			metrics.HttpRateLimitRejectionsTotal.Add(1)
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}



		ctx.Next()
	}
}
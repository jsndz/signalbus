package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jsndz/signalbus/metrics"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type MiddlewareConfig struct {
	RedisClient *redis.Client
	DB          *gorm.DB
}

func NotificationMiddleware(cfg *MiddlewareConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientIP := ctx.ClientIP()
		key := fmt.Sprintf("ratelimit:%s", clientIP)
		now := time.Now().Unix()
		window := int64(60)
		limit := int64(10)
		cfg.RedisClient.ZRemRangeByScore(
			ctx,
			key,
			"0",
			fmt.Sprintf("%d", now-window),
		)
		count, err := cfg.RedisClient.ZCard(ctx, key).Result()
		if err != nil {
			ctx.Next()
			return
		}
		if count >= limit {
			metrics.HttpRateLimitRejectionsTotal.Add(1)
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		id := uuid.New().String()
		member := fmt.Sprintf("%d:%s", now, id)
		cfg.RedisClient.ZAdd(ctx, key, redis.Z{
			Score:  float64(now),
			Member: member,
		})

		cfg.RedisClient.Expire(ctx, key, time.Minute)

		ctx.Next()
	}
}

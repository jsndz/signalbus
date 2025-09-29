package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type MiddlewareConfig struct {
	RedisClient *redis.Client
	DB          *gorm.DB
}


func NotificationMiddleware(cfg *MiddlewareConfig) gin.HandlerFunc{
	return func(ctx *gin.Context) {
		//extract the api key and idem key
		api_key := ctx.GetHeader("X-API-Key")
		if api_key == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing API key"})
			return
		}
		//short term rate limiter on apikey
		//apikey:$API_KEY -> cnt
		//fixed window 10/sec
		shortKey:=fmt.Sprintf("ratelimit:api_key:%s:%d",api_key,time.Now().Unix()/60)
		count ,_ := cfg.RedisClient.Incr(ctx,shortKey).Result()
		if count == 1 {
			cfg.RedisClient.Expire(ctx, shortKey, time.Minute)
		}
		if count > 10 { 
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		//get tenant  ID + quota
		var keyRecord struct {
			ID       uuid.UUID
			QuotaDaily int
			QuotaMonthly int
		} 
		 result:=cfg.DB.Raw(`
			SELECT t.id, t.quota_daily, t.quota_monthly
			FROM tenants t
			JOIN api_keys a ON a.tenant_id = t.id
			WHERE a.hash = ?
		`, api_key).Scan(&keyRecord); 
		if result.Error != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
						"error": fmt.Sprintf("invalid api-key: %v", result.Error),
					})
				return
		}
		// long term rate limiter
		//tenantID:$tenant_ID 
		// check if exist in 
		// get quotamontly and QuotaDaily
		today := time.Now().Format("20060102")
		month := time.Now().Format("200601")
		dailyKey := fmt.Sprintf("quota:tenant:%s:daily:%s", keyRecord.ID, today)
		monthlyKey := fmt.Sprintf("quota:tenant:%s:monthly:%s", keyRecord.ID, month)
		dailyCount, _ := cfg.RedisClient.Incr(ctx, dailyKey).Result()
		if dailyCount == 1 {
			cfg.RedisClient.Expire(ctx, dailyKey, 24*time.Hour)
		}
		monthlyCount, _ := cfg.RedisClient.Incr(ctx, monthlyKey).Result()
		if monthlyCount == 1 {
			cfg.RedisClient.Expire(ctx, monthlyKey, 30*24*time.Hour)
		}
		if dailyCount > int64(keyRecord.QuotaDaily) || monthlyCount > int64(keyRecord.QuotaMonthly) {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "quota exceeded"})
			return
		}
		//set the api key
		ctx.Set("tenant_id",keyRecord.ID)


		ctx.Next()
	}
}
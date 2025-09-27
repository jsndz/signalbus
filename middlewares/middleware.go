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
type bodyWriter struct {
	gin.ResponseWriter
	body []byte
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

func NewNotificationMiddleware(cfg *MiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		apiKey := c.GetHeader("X-API-Key")
		idempotencyKey := c.GetHeader("X-Idempotency-Key")

		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing API key"})
			return
		}

		var tenantID uuid.UUID
		var apiKeyID uuid.UUID
		var quotaDaily, quotaMonthly int

		cacheKey := fmt.Sprintf("apikey:%s", apiKey)
		val, err := cfg.RedisClient.HGetAll(ctx, cacheKey).Result()
		if err == nil && len(val) > 0 {
			tenantID, _ = uuid.Parse(val["tenant_id"])
			apiKeyID, _ = uuid.Parse(val["api_key_id"])
			fmt.Sscanf(val["quota_daily"], "%d", &quotaDaily)
			fmt.Sscanf(val["quota_monthly"], "%d", &quotaMonthly)
		} else {
			var keyRecord struct {
				ID       uuid.UUID
				TenantID uuid.UUID
				QuotaDaily int
				QuotaMonthly int
			}
			err := cfg.DB.Raw(`
				SELECT api_keys.id, api_keys.tenant_id, tenants.quota_daily, tenants.quota_monthly
				FROM api_keys
				JOIN tenants ON tenants.id = api_keys.tenant_id
				WHERE api_keys.hash = ?	
			`, apiKey).Scan(&keyRecord).Error
			
			if err != nil || keyRecord.ID == uuid.Nil {
				newErr := fmt.Sprintf("invalid API KEY: %s", err.Error())
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error":newErr})
				return
			}
			tenantID = keyRecord.TenantID
			apiKeyID = keyRecord.ID
			quotaDaily = keyRecord.QuotaDaily
			quotaMonthly = keyRecord.QuotaMonthly

			cfg.RedisClient.HSet(ctx, cacheKey, map[string]interface{}{
				"tenant_id":    tenantID.String(),
				"api_key_id":   apiKeyID.String(),
				"quota_daily":  quotaDaily,
				"quota_monthly": quotaMonthly,
			})
			cfg.RedisClient.Expire(ctx, cacheKey, 10*time.Minute)
		}

		c.Set("tenant_id", tenantID)
		c.Set("api_key_id", apiKeyID)
		c.Set("idem_key",idempotencyKey)
	
		if idempotencyKey != "" {
			idempRedisKey := fmt.Sprintf("idempotency:%s:%s", tenantID.String(), idempotencyKey)
			resp, err := cfg.RedisClient.Get(ctx, idempRedisKey).Result()
			if err == nil {
				c.Data(http.StatusOK, "application/json", []byte(resp))
				c.Abort()
				return
			}
		}


		shortKey := fmt.Sprintf("ratelimit:apikey:%s:%d", apiKeyID.String(), time.Now().Unix()/60)
		count, _ := cfg.RedisClient.Incr(ctx, shortKey).Result()
		if count == 1 {
			cfg.RedisClient.Expire(ctx, shortKey, time.Minute)
		}
		if count > 10 { 
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		today := time.Now().Format("20060102")
		month := time.Now().Format("200601")
		dailyKey := fmt.Sprintf("quota:tenant:%s:daily:%s", tenantID.String(), today)
		monthlyKey := fmt.Sprintf("quota:tenant:%s:monthly:%s", tenantID.String(), month)

		dailyCount, _ := cfg.RedisClient.Incr(ctx, dailyKey).Result()
		if dailyCount == 1 {
			cfg.RedisClient.Expire(ctx, dailyKey, 24*time.Hour)
		}
		monthlyCount, _ := cfg.RedisClient.Incr(ctx, monthlyKey).Result()
		if monthlyCount == 1 {
			cfg.RedisClient.Expire(ctx, monthlyKey, 30*24*time.Hour)
		}
		if dailyCount > int64(quotaDaily) || monthlyCount > int64(quotaMonthly) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "quota exceeded"})
			return
		}

		bw := &bodyWriter{ResponseWriter: c.Writer}
		c.Writer = bw
		c.Next()


		if idempotencyKey != "" && c.Writer.Status() < 400 {
			idempRedisKey := fmt.Sprintf("idempotency:%s:%s", tenantID.String(), idempotencyKey)
			cfg.RedisClient.Set(ctx, idempRedisKey, bw.body, 24*time.Hour)
		}
	}
}
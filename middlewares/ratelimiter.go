package middlewares

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)
type NotifyRequest struct {
    EventType      string                 `json:"event_type" binding:"required"`
    Data           map[string]interface{} `json:"data" binding:"required"`
    IdempotencyKey string                 `json:"idempotency_key" binding:"required"`
}


type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	r        rate.Limit
	burst    int
}

func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        r,
		burst:    burst,
	}
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.r, rl.burst)
		rl.limiters[key] = limiter
	}
	return limiter
}

func (rl *RateLimiter) Middleware(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		

		limiter := rl.getLimiter(key)
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded, slow down",
			})
			return
		}
		c.Next()
	}
}



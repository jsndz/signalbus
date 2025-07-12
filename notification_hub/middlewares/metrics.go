package middlewares

import (
	"fmt"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/metrics"
)

func MetricsMiddleware() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		start:= time.Now()
		ctx.Next()
		duration := time.Since(start).Seconds()
		endpoint := ctx.FullPath()
		method:= ctx.Request.Method
		status:=fmt.Sprintf("%d", ctx.Writer.Status())
		metrics.HttpRequestTotal.WithLabelValues(endpoint,status,method).Inc()
		metrics.HttpRequestDuration.WithLabelValues(endpoint,method).Observe(duration)

	}
}
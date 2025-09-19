package middlewares

import (
	"fmt"
	"net/http"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/metrics"
)

type ResponseWriterWithStatus struct {
	http.ResponseWriter
	StatusCode int
}

func MetricsMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrappedWriter := &ResponseWriterWithStatus{ResponseWriter: w, StatusCode: http.StatusOK}
		next.ServeHTTP(wrappedWriter, r)
		duration := time.Since(start).Seconds()
		endpoint := r.URL.Path
		method := r.Method
		status := fmt.Sprintf("%d", wrappedWriter.StatusCode)
		metrics.HttpRequestsTotal.WithLabelValues(endpoint, status, method).Inc()
		metrics.HttpRequestDuration.WithLabelValues(endpoint, method).Observe(duration)
	})
}


func GinMetricsMiddleware() gin.HandlerFunc{ return func(ctx *gin.Context) { 
	start:= time.Now() 
	ctx.Next() 
	duration := time.Since(start).Seconds() 
	endpoint := ctx.FullPath() 
	method:= ctx.Request.Method 
	status:=fmt.Sprintf("%d", ctx.Writer.Status()) 
	metrics.HttpRequestsTotal.WithLabelValues(endpoint,status,method).Inc() 
	metrics.HttpRequestDuration.WithLabelValues(endpoint,method).Observe(duration) 
	} 
}
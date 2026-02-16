package metrics

import (
	"github.com/gin-gonic/gin"
	"time"
)

// PrometheusMiddleware собирает метрики по каждому HTTP-запросу
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next() // выполняем обработчик

		duration := time.Since(start).Seconds()
		HTTPRequestCount.WithLabelValues(c.Request.Method, c.FullPath()).Inc()
		HTTPRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
	}
}

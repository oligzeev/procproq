package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

/*
	healthCounter := promauto.NewCounter(prometheus.CounterOpts{Name: "health_metric", Help: "Health check counter"})
	router.GET("/health", func (c *gin.Context) {
		span, _ := opentracing.StartSpanFromContext(c.Request.Context(), c.Request.RequestURI)
		defer span.Finish()
		healthCounter.Inc()
		c.JSON(http.StatusOK, gin.H{"status": "up"})
	})
*/
func PrometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// Package metrics exposes Prometheus metrics for the go-api service:
// a /metrics handler (Go + process collectors) and a Gin middleware
// recording HTTP request duration histograms.
package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry is the service-wide Prometheus registry.
var Registry = prometheus.NewRegistry()

var httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "go_api_http_request_duration_seconds",
	Help:    "HTTP request duration in seconds, by route and status.",
	Buckets: prometheus.DefBuckets,
}, []string{"route", "method", "status"})

func init() {
	Registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		httpDuration,
	)
}

// Handler returns the Prometheus scrape handler for GET /metrics.
func Handler() gin.HandlerFunc {
	return gin.WrapH(promhttp.HandlerFor(Registry, promhttp.HandlerOpts{}))
}

// Middleware records request duration per route/method/status.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		httpDuration.WithLabelValues(
			c.FullPath(),
			c.Request.Method,
			strconv.Itoa(c.Writer.Status()),
		).Observe(time.Since(start).Seconds())
	}
}

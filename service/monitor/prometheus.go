package monitor

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	totalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Number of incoming requests",
		},
		[]string{"path"},
	)

	totalRequestsPerHTTPMethod = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_methods_total",
			Help: "Number of requests per HTTP method",
		},
		[]string{"method"},
	)

	httpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_response_time_seconds",
			Help: "Duration of HTTP requests",
		},
		[]string{"path"},
	)
)

func Setup() error {
	if err := prometheus.Register(totalRequests); err != nil {
		return err
	}
	if err := prometheus.Register(totalRequestsPerHTTPMethod); err != nil {
		return err
	}
	if err := prometheus.Register(httpDuration); err != nil {
		return err
	}

	return nil
}

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(c.Request.URL.Path))

		totalRequests.WithLabelValues(c.Request.URL.Path).Inc()
		totalRequestsPerHTTPMethod.WithLabelValues(c.Request.Method).Inc()

		c.Next()

		timer.ObserveDuration()
	}
}

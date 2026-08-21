package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsHandlerExposesPrometheusFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Middleware())
	r.GET("/metrics", Handler())
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Warm the histogram so the metric is registered with a sample.
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(httptest.NewRecorder(), req)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "go_api_http_request_duration_seconds")
	assert.Contains(t, body, `route="/ping"`)
	assert.Contains(t, body, "go_goroutines")
}

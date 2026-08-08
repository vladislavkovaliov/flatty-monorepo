package applications

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewApplication(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC)

	app := NewApplication(1, "resident", "development", "resident", "styles",
		"http://localhost:8082", "/external-resident", "/", now, now)

	assert.Equal(t, int64(1), app.ID())
	assert.Equal(t, "resident", app.Name())
	assert.Equal(t, "development", app.Env())
	assert.Equal(t, "resident", app.BundleJS())
	assert.Equal(t, "styles", app.StyleURL())
	assert.Equal(t, "http://localhost:8082", app.RemoteOrigin())
	assert.Equal(t, "/external-resident", app.ProxyBasePath())
	assert.Equal(t, "/", app.BasePath())
	assert.Equal(t, now, app.CreatedAt())
	assert.Equal(t, now, app.UpdatedAt())
}

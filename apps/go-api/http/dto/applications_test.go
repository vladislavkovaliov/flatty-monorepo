package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestApplicationResponse_JSONKeys(t *testing.T) {
	t.Parallel()

	now := time.Now()
	resp := ApplicationResponse{
		ID:            1,
		Name:          "resident",
		Env:           "development",
		BundleJs:      "resident",
		StyleUrl:      "styles",
		RemoteOrigin:  "http://localhost:8082",
		ProxyBasePath: "/external-resident",
		BasePath:      "/",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	b, err := json.Marshal(resp)
	assert.NoError(t, err)

	raw := string(b)
	assert.Contains(t, raw, `"name":"resident"`)
	assert.Contains(t, raw, `"bundle_js":"resident"`)
	assert.Contains(t, raw, `"style_url":"styles"`)
	assert.Contains(t, raw, `"remote_origin":"http://localhost:8082"`)
	assert.Contains(t, raw, `"proxy_base_path":"/external-resident"`)
	assert.Contains(t, raw, `"base_path":"/"`)
}

func TestCreateApplicationRequest_Fields(t *testing.T) {
	t.Parallel()

	req := CreateApplicationRequest{
		Name:          "settings",
		Env:           "development",
		BundleJs:      "settings",
		StyleUrl:      "styles",
		RemoteOrigin:  "http://localhost:8081",
		ProxyBasePath: "/external-settings",
		BasePath:      "/",
	}

	assert.Equal(t, "settings", req.Name)
	assert.Equal(t, "development", req.Env)
	assert.Equal(t, "settings", req.BundleJs)
	assert.Equal(t, "styles", req.StyleUrl)
	assert.Equal(t, "http://localhost:8081", req.RemoteOrigin)
	assert.Equal(t, "/external-settings", req.ProxyBasePath)
	assert.Equal(t, "/", req.BasePath)
}

func TestUpdateApplicationRequest_Fields(t *testing.T) {
	t.Parallel()

	req := UpdateApplicationRequest{Name: "resident", Env: "development"}
	assert.Equal(t, "resident", req.Name)
	assert.Equal(t, "development", req.Env)
	assert.Empty(t, req.BasePath) // handler defaults to "/"
}

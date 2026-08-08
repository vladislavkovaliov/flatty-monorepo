package dto

import "time"

type ApplicationResponse struct {
	ID            int64     `json:"id,omitempty" example:"1" binding:"required"`
	Name          string    `json:"name,omitempty" example:"resident" binding:"required"`
	Env           string    `json:"env,omitempty" example:"development" binding:"required"`
	BundleJs      string    `json:"bundle_js,omitempty" example:"resident" binding:"required"`
	StyleUrl      string    `json:"style_url,omitempty" example:"styles"`
	RemoteOrigin  string    `json:"remote_origin,omitempty" example:"http://localhost:8082" binding:"required"`
	ProxyBasePath string    `json:"proxy_base_path,omitempty" example:"/external-resident" binding:"required"`
	BasePath      string    `json:"base_path,omitempty" example:"/"`
	CreatedAt     time.Time `json:"created_at,omitempty" binding:"required"`
	UpdatedAt     time.Time `json:"updated_at,omitempty" binding:"required"`
}

// ListApplicationResponse intentionally has no Total: the registry is small and
// consumers (mfe-host hook, next.config boot fetch) only read .data.
type ListApplicationResponse struct {
	Data []ApplicationResponse `json:"data,omitempty" binding:"required"`
}

type CreateApplicationRequest struct {
	Name          string `json:"name,omitempty" form:"name" binding:"required"`
	Env           string `json:"env,omitempty" form:"env" binding:"required"`
	BundleJs      string `json:"bundle_js,omitempty" form:"bundle_js" binding:"required"`
	StyleUrl      string `json:"style_url,omitempty" form:"style_url"`
	RemoteOrigin  string `json:"remote_origin,omitempty" form:"remote_origin" binding:"required"`
	ProxyBasePath string `json:"proxy_base_path,omitempty" form:"proxy_base_path" binding:"required"`
	BasePath      string `json:"base_path,omitempty" form:"base_path"`
}

// UpdateApplicationRequest is the UPDATE type — do NOT bind the create type on
// update (this avoids replicating the pre-existing bug at http/handlers/category.go:87).
type UpdateApplicationRequest struct {
	Name          string `json:"name,omitempty" form:"name" binding:"required"`
	Env           string `json:"env,omitempty" form:"env" binding:"required"`
	BundleJs      string `json:"bundle_js,omitempty" form:"bundle_js" binding:"required"`
	StyleUrl      string `json:"style_url,omitempty" form:"style_url"`
	RemoteOrigin  string `json:"remote_origin,omitempty" form:"remote_origin" binding:"required"`
	ProxyBasePath string `json:"proxy_base_path,omitempty" form:"proxy_base_path" binding:"required"`
	BasePath      string `json:"base_path,omitempty" form:"base_path"`
}

type DeleteApplicationResponse struct {
	Data int64 `json:"data" example:"1" binding:"required"`
}

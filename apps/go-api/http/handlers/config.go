package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"flatty-budget/go-api/http/dto"
)

type ConfigHandler struct{}

func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{}
}

// CountUsers godoc
//
//	@Summary		Information about health status
//	@Description	Returns health check status
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	dto.HealthResponse
//	@Router			/health [get]
func (h *ConfigHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_ = ctx

	c.JSON(http.StatusOK, dto.HealthResponse{
		Status: "ok",
	})
}

package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func internalError(c *gin.Context, err error) {
	slog.Error("internal server error",
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
		"error", err,
	)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

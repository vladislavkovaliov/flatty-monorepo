package handlers

import (
	"context"
	user_settings_domain "flatty-budget/go-api/domains/user_settings"
	"flatty-budget/go-api/http/dto"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type userSettingsService interface {
	Get(ctx context.Context, userID string) (*user_settings_domain.UserSettings, error)
	Update(ctx context.Context, userID string, input *user_settings_domain.UserSettingsInput) (*user_settings_domain.UserSettings, error)
}

type UserSettingsHandler struct {
	service userSettingsService
}

func NewUserSettingsHandler(svc userSettingsService) *UserSettingsHandler {
	return &UserSettingsHandler{
		service: svc,
	}
}

// GetSettings returns the authenticated user's settings.
// @Summary      Get user settings
// @Description  Returns the authenticated user's settings. If no settings have been saved yet, returns sensible defaults (language: en, theme: system, timezone: UTC, date_format: YYYY-MM-DD).
// @Tags         user-settings
// @Security     Bearer
// @Success      200 {object} dto.UserSettingsResponse
// @Router       /user/me/settings [get]
func (h *UserSettingsHandler) GetSettings(c *gin.Context) {
	userID := c.GetString("userID")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	settings, err := h.service.Get(ctx, userID)
	if err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.UserSettingsResponse{
		UserID:     settings.UserID(),
		Language:   settings.Language(),
		Theme:      settings.Theme(),
		Timezone:   settings.Timezone(),
		DateFormat: settings.DateFormat(),
		CreatedAt:  settings.CreatedAt(),
		UpdatedAt:  settings.UpdatedAt(),
	})
}

// UpdateSettings updates the authenticated user's settings (partial update).
// @Summary      Update user settings
// @Description  Updates the authenticated user's settings. This is a partial update — only send the fields you want to change. Missing fields will keep their current values. If no settings exist yet, they will be created.
// @Tags         user-settings
// @Security     Bearer
// @Accept       json
// @Produce      json
// @Param        body body dto.UpdateUserSettingsRequest true "Settings to update (partial)"
// @Success      200 {object} dto.UserSettingsResponse
// @Failure      400 {object} map[string]string "Invalid request body"
// @Router       /user/me/settings [put]
func (h *UserSettingsHandler) UpdateSettings(c *gin.Context) {
	userID := c.GetString("userID")

	var req dto.UpdateUserSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	current, err := h.service.Get(ctx, userID)
	if err != nil {
		internalError(c, err)
		return
	}

	language := current.Language()
	if req.Language != nil {
		language = *req.Language
	}

	theme := current.Theme()
	if req.Theme != nil {
		theme = *req.Theme
	}

	timezone := current.Timezone()
	if req.Timezone != nil {
		timezone = *req.Timezone
	}

	dateFormat := current.DateFormat()
	if req.DateFormat != nil {
		dateFormat = *req.DateFormat
	}

	input := user_settings_domain.NewUserSettingsInput(language, theme, timezone, dateFormat)

	updated, err := h.service.Update(ctx, userID, input)
	if err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.UserSettingsResponse{
		UserID:     updated.UserID(),
		Language:   updated.Language(),
		Theme:      updated.Theme(),
		Timezone:   updated.Timezone(),
		DateFormat: updated.DateFormat(),
		CreatedAt:  updated.CreatedAt(),
		UpdatedAt:  updated.UpdatedAt(),
	})
}

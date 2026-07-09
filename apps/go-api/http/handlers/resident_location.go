package handlers

import (
	"context"
	"flatty-budget/go-api/http/dto"
	residentlocationservice "flatty-budget/go-api/services/resident_location"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ResidentLocationHandler struct {
	service *residentlocationservice.Service
}

func NewResidentLocationHandler(service *residentlocationservice.Service) *ResidentLocationHandler {
	return &ResidentLocationHandler{
		service: service,
	}
}

// CountResidentLocation godoc
//
//	@Summary		Count all resident location
//	@Description	Returns count resident location from the database
//	@Tags			resident_location
//	@Produce		json
//	@Success		200	{object}	dto.CountResponse
//	@Router			/resident-location/count [get]
func (h *ResidentLocationHandler) Count(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)

	defer cancel()

	total, err := h.service.Count(ctx)

	if err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.CountResponse{
		Total: total,
	})
}

// ListResidentLocation godoc
//
//	@Summary		List all resident location
//	@Description	Returns all resident location from the database
//	@Tags			resident_location
//	@Produce		json
//	@Success		200	{object}	dto.ListResidentLocationResponse
//	@Router			/resident-location [get]
//
//	@Param			limit	query	int	false	"Number of products to return (default 10)"
//	@Param			offset	query	int	false	"Number of products to skip (default 0)"
func (h *ResidentLocationHandler) List(c *gin.Context) {
	defaultLimit := 10
	defaultOffset := 0

	if limit, err := strconv.Atoi(c.Query("limit")); err == nil && limit > 0 {
		defaultLimit = limit
	}

	if offset, err := strconv.Atoi(c.Query("offset")); err == nil && offset >= 0 {
		defaultOffset = offset
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	residentLocations, total, err := h.service.List(ctx, defaultLimit, defaultOffset)

	if err != nil {
		internalError(c, err)
		return
	}

	res := make([]dto.ResidentLocationResponse, 0, len(residentLocations))

	for _, p := range residentLocations {
		res = append(res, dto.ResidentLocationResponse{
			ID:         p.ID(),
			Country:    p.Country(),
			City:       p.City(),
			PostalCode: p.PostalCode(),
			Street:     p.Street(),
			House:      p.House(),
			Apartment:  p.Apartment(),
			CreatedAt:  p.CreatedAt(),
			UpdatedAt:  p.UpdatedAt(),
		})
	}

	c.JSON(http.StatusOK, dto.ListResidentLocationResponse{
		Data:  res,
		Total: total,
	})
}

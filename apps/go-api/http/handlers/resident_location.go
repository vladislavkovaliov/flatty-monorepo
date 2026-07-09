package handlers

import (
	"context"
	"flatty-budget/go-api/http/dto"
	residentlocationservice "flatty-budget/go-api/services/resident_location"
	"net/http"
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

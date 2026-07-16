package handlers

import (
	"context"
	"flatty-budget/go-api/http/dto"
	userservice "flatty-budget/go-api/services/user"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service *userservice.Service
}

func NewUseHandler(service *userservice.Service) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// ListUser godoc
//
//	@Summary		Return list of user
//	@Description	Returns list of user from the database
//	@Tags			user
//	@Produce		json
//	@Success		200	{object}	dto.ListUserResponse
//	@Router			/user [get]
func (h *UserHandler) List(c *gin.Context) {
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

	users, err := h.service.List(ctx, defaultLimit, defaultOffset)

	if err != nil {
		internalError(c, err)
		return
	}

	res := make([]dto.UserResponse, 0, len(users))

	for _, u := range users {
		res = append(res, dto.UserResponse{
			ID:            u.ID(),
			Name:          u.Name(),
			Email:         u.Email(),
			EmailVerified: u.EmailVerified(),
			Image:         u.Image(),
			CreatedAt:     u.CreatedAt(),
			UpdatedAt:     u.UpdatedAt(),
		})
	}

	c.JSON(http.StatusOK, dto.ListUserResponse{
		Data:  res,
		Total: len(users),
	})
}

// GetUserByID godoc
//
//	@Summary		Return user by id
//	@Description	Returns user by id from the database
//	@Tags			user
//	@Produce		json
//	@Param          id      path        string     true    "User ID"
//	@Success		200	{object}	dto.UserResponse
//	@Router			/user/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idParam := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)

	defer cancel()

	user, err := h.service.GetUserByID(ctx, idParam)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, dto.UserResponse{
		ID:            user.ID(),
		Name:          user.Name(),
		Email:         user.Email(),
		EmailVerified: user.EmailVerified(),
		Image:         user.Image(),
		CreatedAt:     user.CreatedAt(),
		UpdatedAt:     user.UpdatedAt(),
	})
}

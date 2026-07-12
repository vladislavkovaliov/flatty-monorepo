package handlers

import (
	"context"
	"errors"
	"flatty-budget/go-api/domains/category"
	"flatty-budget/go-api/http/dto"
	categoryservice "flatty-budget/go-api/services/category"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type CategoryHandler struct {
	service *categoryservice.Service
}

func NewCategoryHandler(service *categoryservice.Service) *CategoryHandler {
	return &CategoryHandler{
		service: service,
	}
}

// CountCategories godoc
//
//	@Summary		Count all categories
//	@Description	Returns count categories from the database
//	@Tags			categories
//	@Produce		json
//	@Success		200	{object}	dto.CountResponse
//	@Router			/categories/count [get]
func (h *CategoryHandler) Count(c *gin.Context) {
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

// ListCategories godoc
//
//	@Summary		List all categories
//	@Description	Returns all categories from the database
//	@Tags			categories
//	@Produce		json
//	@Success		200	{object}	dto.ListCategoryResponse
//	@Router			/categories [get]
//
//	@Param			limit	query	int	false	"Number of products to return (default 10)"
//	@Param			offset	query	int	false	"Number of products to skip (default 0)"
func (h *CategoryHandler) List(c *gin.Context) {
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

	categories, total, err := h.service.List(ctx, defaultLimit, defaultOffset)

	if err != nil {
		internalError(c, err)
		return
	}

	res := make([]dto.CategoryResponse, 0, len(categories))

	for _, p := range categories {
		res = append(res, dto.CategoryResponse{
			ID:          p.ID(),
			Name:        p.Name(),
			Description: p.Description(),
			CreatedAt:   p.CreatedAt(),
			UpdatedAt:   p.UpdatedAt(),
		})
	}

	c.JSON(http.StatusOK, dto.ListCategoryResponse{
		Data:  res,
		Total: total,
	})
}

// CreateCategory godoc
//
//	@Summary		Create a category
//	@Description	Add a new category to the database
//	@Tags			categories
//	@Accept			json
//	@Produce		json
//	@Param			body	body	dto.CreateCategoryRequest	true	"Category data"
//	@Success		201		{object}	dto.CategoryResponse
//	@Failure		400		{object}	map[string]string
//	@Router			/categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var req dto.CreateCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	category, err := h.service.Create(ctx, category.NewCategoryInput(
		req.Name,
		req.Description,
	))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.CategoryResponse{
		ID:          category.ID(),
		Name:        category.Name(),
		Description: category.Description(),
		CreatedAt:   category.CreatedAt(),
		UpdatedAt:   category.UpdatedAt(),
	})
}

// UpdateCategory godoc
//
//	@Summary		Update a category
//	@Description	Update a category to the database
//	@Tags			categories
//	@Accept			json
//	@Produce		json
//	@Param          id      path        int     true    "Category ID"
//	@Param			body	body	    dto.UpdateCategoryRequest	true	"Category data"
//	@Success		201		{object}	dto.CategoryResponse
//	@Failure		400		{object}	map[string]string
//	@Router			/categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req dto.CreateCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	category, err := h.service.Update(ctx, id, category.NewCategoryInput(
		req.Name,
		req.Description,
	))

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		internalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.CategoryResponse{
		ID:          category.ID(),
		Name:        category.Name(),
		Description: category.Description(),
		CreatedAt:   category.CreatedAt(),
		UpdatedAt:   category.UpdatedAt(),
	})
}

// DeleteCategory godoc
//
//	@Summary		Delete a category
//	@Description	Delete a category to the database
//	@Tags			categories
//	@Accept			json
//	@Produce		json
//	@Param          id      path        int     true    "Category ID"
//	@Success		201		{object}	dto.DeleteCategoryResponse
//	@Failure		400		{object}	map[string]string
//	@Router			/categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	returningId, err := h.service.Delete(ctx, id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.DeleteCategoryResponse{
		Data: returningId,
	})
}

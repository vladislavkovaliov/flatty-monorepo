package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	applicationsdomain "flatty-budget/go-api/domains/applications"
	"flatty-budget/go-api/http/dto"
	"flatty-budget/go-api/internal/config"
	applicationsservice "flatty-budget/go-api/services/applications"
)

type ApplicationsHandler struct {
	service *applicationsservice.Service
	cfg     *config.Config
}

func NewApplicationsHandler(service *applicationsservice.Service, cfg *config.Config) *ApplicationsHandler {
	return &ApplicationsHandler{
		service: service,
		cfg:     cfg,
	}
}

// ListApplications godoc
//
//	@Summary		List applications for the current APP_ENV
//	@Description	Returns the micro-frontend registry for the environment selected by APP_ENV. Public (unauthenticated): shells and the entrypoint boot fetch rely on it.
//	@Tags			applications
//	@Produce		json
//	@Success		200	{object}	dto.ListApplicationResponse
//	@Router			/applications [get]
func (h *ApplicationsHandler) List(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	apps, err := h.service.ListByEnv(ctx, h.cfg.AppEnv)
	if err != nil {
		internalError(c, err)
		return
	}

	res := make([]dto.ApplicationResponse, 0, len(apps))

	for _, a := range apps {
		res = append(res, toApplicationResponse(a))
	}

	c.JSON(http.StatusOK, dto.ListApplicationResponse{Data: res})
}

// CreateApplication godoc
//
//	@Summary		Create an application registry row
//	@Description	Add a new application entry to the registry
//	@Tags			applications
//	@Accept			json
//	@Produce		json
//	@Param			body	body	dto.CreateApplicationRequest	true	"Application data"
//	@Success		201		{object}	dto.ApplicationResponse
//	@Failure		400		{object}	map[string]string
//	@Router			/applications [post]
func (h *ApplicationsHandler) Create(c *gin.Context) {
	var req dto.CreateApplicationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.BasePath == "" {
		req.BasePath = "/"
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	app, err := h.service.Create(ctx, applicationsdomain.NewApplicationInput(
		req.Name,
		req.Env,
		req.BundleJs,
		req.StyleUrl,
		req.RemoteOrigin,
		req.ProxyBasePath,
		req.BasePath,
	))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toApplicationResponse(app))
}

// UpdateApplication godoc
//
//	@Summary		Update an application registry row
//	@Description	Update an existing application entry
//	@Tags			applications
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int	true	"Application ID"
//	@Param			body	body	dto.UpdateApplicationRequest	true	"Application data"
//	@Success		200		{object}	dto.ApplicationResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Router			/applications/{id} [put]
func (h *ApplicationsHandler) Update(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req dto.UpdateApplicationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.BasePath == "" {
		req.BasePath = "/"
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	app, err := h.service.Update(ctx, id, applicationsdomain.NewApplicationInput(
		req.Name,
		req.Env,
		req.BundleJs,
		req.StyleUrl,
		req.RemoteOrigin,
		req.ProxyBasePath,
		req.BasePath,
	))

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, toApplicationResponse(app))
}

// DeleteApplication godoc
//
//	@Summary		Delete an application registry row
//	@Description	Remove an application entry from the registry
//	@Tags			applications
//	@Produce		json
//	@Param			id	path	int	true	"Application ID"
//	@Success		200	{object}	dto.DeleteApplicationResponse
//	@Failure		404	{object}	map[string]string
//	@Router			/applications/{id} [delete]
func (h *ApplicationsHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	returningID, err := h.service.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.DeleteApplicationResponse{Data: returningID})
}

func toApplicationResponse(a *applicationsdomain.Application) dto.ApplicationResponse {
	return dto.ApplicationResponse{
		ID:            a.ID(),
		Name:          a.Name(),
		Env:           a.Env(),
		BundleJs:      a.BundleJS(),
		StyleUrl:      a.StyleURL(),
		RemoteOrigin:  a.RemoteOrigin(),
		ProxyBasePath: a.ProxyBasePath(),
		BasePath:      a.BasePath(),
		CreatedAt:     a.CreatedAt(),
		UpdatedAt:     a.UpdatedAt(),
	}
}

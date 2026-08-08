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
	applicationsservice "flatty-budget/go-api/services/applications"
)

type AdminApplicationsHandler struct {
	service *applicationsservice.Service
}

func NewAdminApplicationsHandler(service *applicationsservice.Service) *AdminApplicationsHandler {
	return &AdminApplicationsHandler{
		service: service,
	}
}

// ListPage renders the full admin page (layout → applications_list → applications_table).
func (h *AdminApplicationsHandler) ListPage(c *gin.Context) {
	apps, err := h.listAll(c)
	if err != nil {
		internalError(c, err)
		return
	}

	c.HTML(http.StatusOK, "layout", gin.H{"Groups": groupByEnv(apps)})
}

// CreateRow handles the htmx hx-post form and re-renders the table fragment.
func (h *AdminApplicationsHandler) CreateRow(c *gin.Context) {
	var req dto.CreateApplicationRequest

	if err := c.ShouldBind(&req); err != nil {
		c.String(http.StatusBadRequest, "invalid form data: %s", err.Error())
		return
	}

	if req.BasePath == "" {
		req.BasePath = "/"
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if _, err := h.service.Create(ctx, applicationsdomain.NewApplicationInput(
		req.Name,
		req.Env,
		req.BundleJs,
		req.StyleUrl,
		req.RemoteOrigin,
		req.ProxyBasePath,
		req.BasePath,
	)); err != nil {
		internalError(c, err)
		return
	}

	h.renderTable(c)
}

// EditRowForm returns the inline <tr> edit form for one row (htmx hx-get swap).
func (h *AdminApplicationsHandler) EditRowForm(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	apps, err := h.listAll(c)
	if err != nil {
		internalError(c, err)
		return
	}

	for _, a := range apps {
		if a.ID() == id {
			c.HTML(http.StatusOK, "applications_row_edit", gin.H{"Application": toApplicationResponse(a)})
			return
		}
	}

	c.Status(http.StatusNotFound)
}

// UpdateRow handles the inline edit form (htmx hx-put) and re-renders the table fragment.
func (h *AdminApplicationsHandler) UpdateRow(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	var req dto.UpdateApplicationRequest

	if err := c.ShouldBind(&req); err != nil {
		c.String(http.StatusBadRequest, "invalid form data: %s", err.Error())
		return
	}

	if req.BasePath == "" {
		req.BasePath = "/"
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if _, err := h.service.Update(ctx, id, applicationsdomain.NewApplicationInput(
		req.Name,
		req.Env,
		req.BundleJs,
		req.StyleUrl,
		req.RemoteOrigin,
		req.ProxyBasePath,
		req.BasePath,
	)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Status(http.StatusNotFound)
			return
		}

		internalError(c, err)
		return
	}

	h.renderTable(c)
}

// DeleteRow handles the hx-delete button and re-renders the table fragment.
func (h *AdminApplicationsHandler) DeleteRow(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if _, err := h.service.Delete(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Status(http.StatusNotFound)
			return
		}

		internalError(c, err)
		return
	}

	h.renderTable(c)
}

func (h *AdminApplicationsHandler) listAll(c *gin.Context) ([]*applicationsdomain.Application, error) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	return h.service.ListAll(ctx)
}

// renderTable re-renders the table fragment — the htmx response after create/update/delete.
func (h *AdminApplicationsHandler) renderTable(c *gin.Context) {
	apps, err := h.listAll(c)
	if err != nil {
		internalError(c, err)
		return
	}

	c.HTML(http.StatusOK, "applications_table", gin.H{"Groups": groupByEnv(apps)})
}

func groupByEnv(apps []*applicationsdomain.Application) map[string][]dto.ApplicationResponse {
	groups := make(map[string][]dto.ApplicationResponse)

	for _, a := range apps {
		env := a.Env()
		groups[env] = append(groups[env], toApplicationResponse(a))
	}

	return groups
}

package handlers

import (
	"context"
	"errors"
	expensedomain "flatty-budget/go-api/domains/expenses"
	"flatty-budget/go-api/http/dto"
	expensesservice "flatty-budget/go-api/services/expenses"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type ExpenseHandler struct {
	service *expensesservice.Service
}

func NewExpenseHandler(service *expensesservice.Service) *ExpenseHandler {
	return &ExpenseHandler{
		service: service,
	}
}

// CountExpenses godoc
//
//	@Summary		Count all expenses
//	@Description	Returns count of expenses from the database
//	@Tags			expenses
//	@Produce		json
//	@Success		200	{object}	dto.CountResponse
//	@Router			/expenses/count [get]
//	@Param			residentLocationId	query	int	false	"Resident location ID"
func (h *ExpenseHandler) Count(c *gin.Context) {
	residentLocationID, ok := parseResidentLocationID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "residentLocationId is required"})
		return
	}
	userID := c.GetString("userID")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	total, err := h.service.Count(ctx, residentLocationID, userID)

	if err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.CountResponse{
		Total: total,
	})
}

// ListExpenses godoc
//
//	@Summary		List all expenses
//	@Description	Returns all expenses from the database
//	@Tags			expenses
//	@Produce		json
//	@Success		200	{object}	dto.ListExpenseWithCategoryResponse
//	@Router			/expenses [get]
//
//	@Param			residentLocationId	query	int	false	"Resident location ID"
//	@Param			limit	query	int	false	"Number of items to return (default 10)"
//	@Param			offset	query	int	false	"Number of items to skip (default 0)"
func (h *ExpenseHandler) List(c *gin.Context) {
	residentLocationID, ok := parseResidentLocationID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "residentLocationId is required"})
		return
	}
	userID := c.GetString("userID")

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

	expenses, total, err := h.service.List(ctx, residentLocationID, userID, defaultLimit, defaultOffset)

	if err != nil {
		internalError(c, err)
		return
	}

	res := make([]dto.ExpenseWithCategoryResponse, 0, len(expenses))

	for _, e := range expenses {
		category := e.Category()

		categoryRes := dto.CategoryResponse{
			ID:          category.ID(),
			Name:        category.Name(),
			Description: category.Description(),
			CreatedAt:   category.CreatedAt(),
			UpdatedAt:   category.UpdatedAt(),
		}

		res = append(res, dto.ExpenseWithCategoryResponse{
			ID:                 e.ID(),
			ResidentLocationID: e.ResidentLocationID(),
			CategoryID:         e.CategoryID(),
			Amount:             e.Amount(),
			Month:              e.Month(),
			Year:               e.Year(),
			CreatedAt:          e.CreatedAt(),
			UpdatedAt:          e.UpdatedAt(),
			Description:        e.Description(),
			UserID:             e.UserID(),
			Category:           categoryRes,
		})
	}

	c.JSON(http.StatusOK, dto.ListExpenseWithCategoryResponse{
		Data:  res,
		Total: total,
	})
}

// CreateExpense godoc
//
//	@Summary		Create an expense
//	@Description	Add a new expense to the database
//	@Tags			expenses
//	@Accept			json
//	@Produce		json
//	@Param			body	body	dto.CreateExpenseRequest	true	"Expense data"
//	@Success		201		{object}	dto.ExpenseResponse
//	@Failure		400		{object}	map[string]string
//	@Router			/expenses [post]
func (h *ExpenseHandler) Create(c *gin.Context) {
	var req dto.CreateExpenseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	expense, err := h.service.Create(ctx, expensedomain.NewExpenseInput(
		req.ResidentLocationID,
		req.CategoryID,
		req.Amount,
		req.Description,
		req.Month,
		req.Year,
	), userID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ExpenseResponse{
		ID:                 expense.ID(),
		ResidentLocationID: expense.ResidentLocationID(),
		CategoryID:         expense.CategoryID(),
		Amount:             expense.Amount(),
		Month:              expense.Month(),
		Year:               expense.Year(),
		CreatedAt:          expense.CreatedAt(),
		UpdatedAt:          expense.UpdatedAt(),
		Description:        expense.Description(),
		UserID:             expense.UserID(),
	})
}

// UpdateExpense godoc
//
//	@Summary		Update an expense
//	@Description	Update an expense in the database
//	@Tags			expenses
//	@Accept			json
//	@Produce		json
//	@Param          id      path        int     true    "Expense ID"
//	@Param			body	body	    dto.UpdateExpenseRequest	true	"Expense data"
//	@Success		201		{object}	dto.ExpenseResponse
//	@Failure		400		{object}	map[string]string
//	@Router			/expenses/{id} [put]
func (h *ExpenseHandler) Update(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req dto.UpdateExpenseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	expense, err := h.service.Update(ctx, id, expensedomain.NewExpenseInput(
		req.ResidentLocationID,
		req.CategoryID,
		req.Amount,
		req.Description,
		req.Month,
		req.Year,
	), userID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		internalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ExpenseResponse{
		ID:                 expense.ID(),
		ResidentLocationID: expense.ResidentLocationID(),
		CategoryID:         expense.CategoryID(),
		Amount:             expense.Amount(),
		Month:              expense.Month(),
		Year:               expense.Year(),
		CreatedAt:          expense.CreatedAt(),
		UpdatedAt:          expense.UpdatedAt(),
		Description:        expense.Description(),
		UserID:             expense.UserID(),
	})
}

// DeleteExpense godoc
//
//	@Summary		Delete an expense
//	@Description	Delete an expense from the database
//	@Tags			expenses
//	@Accept			json
//	@Produce		json
//	@Param          id      path        int     true    "Expense ID"
//	@Success		200		{object}	dto.DeleteExpenseResponse
//	@Failure		400		{object}	map[string]string
//	@Router			/expenses/{id} [delete]
func (h *ExpenseHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	returningID, err := h.service.Delete(ctx, id, userID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.DeleteExpenseResponse{
		Data: returningID,
	})
}

// GetYearsAndMonths godoc
//
//	@Summary		List all grouped expenses by year and month
//	@Description	Returns all grouped expenses by year and month from the database
//	@Tags			expenses
//	@Produce		json
//	@Success		200	{object}	dto.ListYearAndMonthResponse
//	@Router			/expenses/get-years-and-months [get]
//
//	@Param			residentLocationId	query	int	false	"Resident location ID"
func (h *ExpenseHandler) GetYearsAndMonths(c *gin.Context) {
	residentLocationID, ok := parseResidentLocationID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "residentLocationId is required"})
		return
	}
	userID := c.GetString("userID")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	items, err := h.service.GetYearsAndMonths(ctx, residentLocationID, userID)
	if err != nil {
		internalError(c, err)
		return
	}

	res := make([]dto.YearAndMonthResponse, 0, len(items))

	for _, y := range items {
		res = append(res, dto.YearAndMonthResponse{
			Year:     y.Year(),
			Month:    y.Month(),
			Expenses: y.Expenses(),
		})
	}

	c.JSON(http.StatusOK, dto.ListYearAndMonthResponse{
		Data: res,
	})
}

// GetExpensesByYearMonth godoc
//
//	@Summary		List all expenses by year and month
//	@Description	Returns all expenses by year and month from the database
//	@Tags			expenses
//	@Produce		json
//	@Success		200	{object}	dto.ListExpenseWithCategoryResponse
//	@Router			/expenses/get-by-year-month [get]
//
//	@Param			residentLocationId	query	int	true	"Resident location ID"
//	@Param			year	            query	int	true	"Year"
//	@Param			month	            query	int	true	"Month"
func (h *ExpenseHandler) GetExpensesByYearMonth(c *gin.Context) {
	residentLocationID, ok := parseResidentLocationID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "residentLocationId is required"})
		return
	}

	year, ok := parseYear(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "year is required"})
		return
	}

	month, ok := parseMonth(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "month is required"})
		return
	}

	userID := c.GetString("userID")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	expenses, err := h.service.GetExpensesByYearMonth(ctx, residentLocationID, year, month, userID)

	if err != nil {
		internalError(c, err)
		return
	}

	res := make([]dto.ExpenseWithCategoryResponse, 0, len(expenses))

	for _, e := range expenses {
		category := e.Category()

		categoryRes := dto.CategoryResponse{
			ID:          category.ID(),
			Name:        category.Name(),
			Description: category.Description(),
			CreatedAt:   category.CreatedAt(),
			UpdatedAt:   category.UpdatedAt(),
		}

		res = append(res, dto.ExpenseWithCategoryResponse{
			ID:                 e.ID(),
			ResidentLocationID: e.ResidentLocationID(),
			CategoryID:         e.CategoryID(),
			Amount:             e.Amount(),
			Month:              e.Month(),
			Year:               e.Year(),
			CreatedAt:          e.CreatedAt(),
			UpdatedAt:          e.UpdatedAt(),
			Description:        e.Description(),
			UserID:             e.UserID(),
			Category:           categoryRes,
		})
	}

	c.JSON(http.StatusOK, dto.ListExpenseWithCategoryResponse{
		Data:  res,
		Total: len(expenses),
	})
}

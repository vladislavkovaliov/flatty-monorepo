package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"flatty-budget/go-api/http/dto"
	totalservice "flatty-budget/go-api/services/expense_stats"

	"github.com/gin-gonic/gin"
)

type ExpenseStatsHandler struct {
	totalSvc    *totalservice.MonthlyTotalService
	averageSvc  *totalservice.MonthlyAverageService
}

func NewExpenseStatsHandler(
	totalSvc *totalservice.MonthlyTotalService,
	averageSvc *totalservice.MonthlyAverageService,
) *ExpenseStatsHandler {
	return &ExpenseStatsHandler{
		totalSvc:   totalSvc,
		averageSvc: averageSvc,
	}
}

func parseMonthYear(c *gin.Context) (*int, *int) {
	var month, year *int
	if m := c.Query("month"); m != "" {
		if v, err := strconv.Atoi(m); err == nil && v >= 1 && v <= 12 {
			month = &v
		}
	}
	if y := c.Query("year"); y != "" {
		if v, err := strconv.Atoi(y); err == nil && v >= 2000 {
			year = &v
		}
	}
	return month, year
}

// ListTotals godoc
//
//	@Summary		List monthly totals
//	@Description	Returns monthly expense totals, optionally filtered by month/year
//	@Tags			expense-stats
//	@Produce		json
//	@Param			month	query	int	false	"Month (1-12)"
//	@Param			year	query	int	false	"Year (2000+)"
//	@Success		200	{object}	dto.ListMonthlyTotalResponse
//	@Router			/expenses/stats/totals [get]
func (h *ExpenseStatsHandler) ListTotals(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	month, year := parseMonthYear(c)

	totals, err := h.totalSvc.List(ctx, month, year)
	if err != nil {
		internalError(c, err)
		return
	}

	res := make([]dto.MonthlyTotalResponse, 0, len(totals))
	for _, t := range totals {
		res = append(res, dto.MonthlyTotalResponse{
			Month:      t.Month(),
			Year:       t.Year(),
			TotalSpent: t.TotalSpent(),
		})
	}

	c.JSON(http.StatusOK, dto.ListMonthlyTotalResponse{Data: res})
}

// ListAverages godoc
//
//	@Summary		List monthly averages
//	@Description	Returns monthly expense averages, optionally filtered by month/year
//	@Tags			expense-stats
//	@Produce		json
//	@Param			month	query	int	false	"Month (1-12)"
//	@Param			year	query	int	false	"Year (2000+)"
//	@Success		200	{object}	dto.ListMonthlyAverageResponse
//	@Router			/expenses/stats/averages [get]
func (h *ExpenseStatsHandler) ListAverages(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	month, year := parseMonthYear(c)

	averages, err := h.averageSvc.List(ctx, month, year)
	if err != nil {
		internalError(c, err)
		return
	}

	res := make([]dto.MonthlyAverageResponse, 0, len(averages))
	for _, a := range averages {
		res = append(res, dto.MonthlyAverageResponse{
			Month:         a.Month(),
			Year:          a.Year(),
			AverageAmount: a.AverageAmount(),
			ExpenseCount:  a.ExpenseCount(),
		})
	}

	c.JSON(http.StatusOK, dto.ListMonthlyAverageResponse{Data: res})
}

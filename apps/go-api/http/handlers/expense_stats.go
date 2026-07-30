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
	totalSvc   *totalservice.MonthlyTotalService
	averageSvc *totalservice.MonthlyAverageService
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

func parseResidentLocationID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Query("residentLocationId"), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// ListTotals godoc
//
//	@Summary		List monthly totals
//	@Description	Returns monthly expense totals for a resident location owned by the caller, optionally filtered by month/year
//	@Tags			expense-stats
//	@Produce		json
//	@Param			residentLocationId	query	int	true	"Resident Location ID"
//	@Param			month	query	int	false	"Month (1-12)"
//	@Param			year	query	int	false	"Year (2000+)"
//	@Success		200	{object}	dto.ListMonthlyTotalResponse
//	@Failure		400	{object}	map[string]string
//	@Router			/expenses/stats/totals [get]
func (h *ExpenseStatsHandler) ListTotals(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	residentLocationID, ok := parseResidentLocationID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "residentLocationId is required"})
		return
	}
	userID := c.GetString("userID")

	month, year := parseMonthYear(c)

	totals, err := h.totalSvc.List(ctx, residentLocationID, userID, month, year)
	if err != nil {
		internalError(c, err)
		return
	}

	res := make([]dto.MonthlyTotalResponse, 0, len(totals))
	for _, t := range totals {
		res = append(res, dto.MonthlyTotalResponse{
			ResidentLocationID: t.ResidentLocationID(),
			Month:              t.Month(),
			Year:               t.Year(),
			TotalSpent:         t.TotalSpent(),
		})
	}

	c.JSON(http.StatusOK, dto.ListMonthlyTotalResponse{Data: res})
}

// ListAverages godoc
//
//	@Summary		List monthly averages
//	@Description	Returns monthly expense averages for a resident location owned by the caller, optionally filtered by month/year
//	@Tags			expense-stats
//	@Produce		json
//	@Param			residentLocationId	query	int	true	"Resident Location ID"
//	@Param			month	query	int	false	"Month (1-12)"
//	@Param			year	query	int	false	"Year (2000+)"
//	@Success		200	{object}	dto.ListMonthlyAverageResponse
//	@Failure		400	{object}	map[string]string
//	@Router			/expenses/stats/averages [get]
func (h *ExpenseStatsHandler) ListAverages(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	residentLocationID, ok := parseResidentLocationID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "residentLocationId is required"})
		return
	}
	userID := c.GetString("userID")

	month, year := parseMonthYear(c)

	averages, err := h.averageSvc.List(ctx, residentLocationID, userID, month, year)
	if err != nil {
		internalError(c, err)
		return
	}

	res := make([]dto.MonthlyAverageResponse, 0, len(averages))
	for _, a := range averages {
		res = append(res, dto.MonthlyAverageResponse{
			ResidentLocationID: a.ResidentLocationID(),
			Month:              a.Month(),
			Year:               a.Year(),
			AverageAmount:      a.AverageAmount(),
			ExpenseCount:       a.ExpenseCount(),
		})
	}

	c.JSON(http.StatusOK, dto.ListMonthlyAverageResponse{Data: res})
}

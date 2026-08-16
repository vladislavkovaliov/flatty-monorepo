package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

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

func parseYear(c *gin.Context) (int64, bool) {
	y, err := strconv.ParseInt(c.Query("year"), 10, 64)
	if err != nil || y < 0 {
		return 0, false
	}

	return y, true
}

func parseMonth(c *gin.Context) (int64, bool) {
	m, err := strconv.ParseInt(c.Query("month"), 10, 64)
	if err != nil || m < 1 || m > 12 {
		return 0, false
	}

	return m, true
}

func parseResidentLocationID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Query("residentLocationId"), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}

	return id, true
}

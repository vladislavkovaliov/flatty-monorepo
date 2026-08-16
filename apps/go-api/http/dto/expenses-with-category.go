package dto

import (
	"time"
)

type ExpenseWithCategoryResponse struct {
	ID                 int64            `json:"id,omitempty" example:"1" binding:"required"`
	ResidentLocationID int64            `json:"resident_location_id,omitempty" example:"1" binding:"required"`
	CategoryID         int64            `json:"category_id,omitempty" example:"1" binding:"required"`
	Amount             float64          `json:"amount,omitempty" example:"150.50" binding:"required"`
	Description        string           `json:"description" example:"Some text" binding:"required"`
	Month              int              `json:"month,omitempty" example:"7" binding:"required"`
	Year               int              `json:"year,omitempty" example:"2026" binding:"required"`
	CreatedAt          time.Time        `json:"created_at,omitempty" example:"2026-07-13T12:00:00Z" binding:"required"`
	UpdatedAt          time.Time        `json:"updated_at,omitempty" example:"2026-07-13T12:00:00Z" binding:"required"`
	UserID             string           `json:"userID" example:"123458697" binding:"required"`
	Category           CategoryResponse `json:"category" binding:"required"`
}

type ListExpenseWithCategoryResponse struct {
	Data  []ExpenseWithCategoryResponse `json:"data,omitempty" binding:"required"`
	Total int                           `json:"total,omitempty" binding:"required"`
}

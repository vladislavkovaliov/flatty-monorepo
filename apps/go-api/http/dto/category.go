package dto

import "time"

type CategoryResponse struct {
	ID          int64     `json:"id,omitempty" example:"1" binding:"required"`
	Name        string    `json:"name,omitempty" example:"utilities" binding:"required"`
	Description string    `json:"description,omitempty" example:"Коммунальные платежи" binding:"required"`
	CreatedAt   time.Time `json:"created_at,omitempty" example:"2026-07-09 08:34:05.796617" binding:"required"`
	UpdatedAt   time.Time `json:"updated_at,omitempty" example:"2026-07-09 08:34:05.796617" binding:"required"`
}

type ListCategoryResponse struct {
	Data  []CategoryResponse `json:"data,omitempty" binding:"required"`
	Total int                `json:"total,omitempty" binding:"required"`
}

type CreateCategoryRequest struct {
	Name        string `json:"name,omitempty" example:"utilities" binding:"required"`
	Description string `json:"description,omitempty" example:"Коммунальные платежи" binding:"required"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name,omitempty" example:"utilities" binding:"required"`
	Description string `json:"description,omitempty" example:"Коммунальные платежи" binding:"required"`
}

type DeleteCategoryResponse struct {
	Data int64 `json:"data" example:"1" binding:"required"`
}

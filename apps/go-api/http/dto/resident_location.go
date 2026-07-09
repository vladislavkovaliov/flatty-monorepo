package dto

import "time"

type CountResponse struct {
	Total int `json:"total" example:"138" binding:"required"`
}

type ResidentLocationResponse struct {
	ID         int64     `json:"id,omitempty" example:"123" binding:"required"`
	Country    string    `json:"country,omitempty" example:"Poland" binding:"required"`
	City       string    `json:"city,omitempty" example:"Warsaw" binding:"required"`
	PostalCode string    `json:"postal_code,omitempty" example:"00-945" binding:"required"`
	Street     string    `json:"street,omitempty" example:"Bobr" binding:"required"`
	House      string    `json:"house,omitempty" example:"1" binding:"required"`
	Apartment  string    `json:"apartment,omitempty" example:"2" binding:"required"`
	CreatedAt  time.Time `json:"created_at,omitempty" example:"2026-07-09 08:34:05.796617" binding:"required"`
	UpdatedAt  time.Time `json:"updated_at,omitempty" example:"2026-07-09 08:34:05.796617" binding:"required"`
}

type ListResidentLocationResponse struct {
	Data  []ResidentLocationResponse `json:"data,omitempty" binding:"required"`
	Total int                        `json:"total,omitempty" binding:"required"`
}

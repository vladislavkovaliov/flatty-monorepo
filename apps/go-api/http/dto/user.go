package dto

import "time"

type UserResponse struct {
	ID            string    `json:"id,omitempty" example:"1" binding:"required"`
	Name          string    `json:"name" example:"clark" binding:"required"`
	Email         string    `json:"email" example:"superman@gmail.com" binding:"required"`
	EmailVerified bool      `json:"email_verified" example:"true" binding:"required"`
	Image         *string   `json:"image" example:"string"`
	CreatedAt     time.Time `json:"created_at" example:"2020-04-13t12:00:00Z" binding:"required"`
	UpdatedAt     time.Time `json:"updated_at" example:"2020-04-13t12:00:00Z" binding:"required"`
}

type ListUserResponse struct {
	Data  []UserResponse `json:"data" binding:"required"`
	Total int            `json:"total" binding:"required"`
}

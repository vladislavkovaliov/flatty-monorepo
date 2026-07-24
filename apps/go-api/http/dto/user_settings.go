package dto

import "time"

type UserSettingsResponse struct {
	UserID     string    `json:"user_id"     example:"00000000-0000-0000-0000-000000000001"`
	Language   string    `json:"language"    example:"en"`
	Theme      string    `json:"theme"       example:"dark"`
	Timezone   string    `json:"timezone"    example:"UTC+0"`
	DateFormat string    `json:"date_format" example:"YYYY-MM-DD"`
	CreatedAt  time.Time `json:"created_at"  example:"2024-01-15T10:30:00Z"`
	UpdatedAt  time.Time `json:"updated_at"  example:"2024-01-15T10:30:00Z"`
}

type UpdateUserSettingsRequest struct {
	Language   *string `json:"language"    example:"en"`
	Theme      *string `json:"theme"       example:"dark"`
	Timezone   *string `json:"timezone"    example:"UTC+0"`
	DateFormat *string `json:"date_format" example:"YYYY-MM-DD"`
}

package dto

type MonthlyTotalResponse struct {
	Month      int     `json:"month" example:"6"`
	Year       int     `json:"year" example:"2026"`
	TotalSpent float64 `json:"total_spent" example:"1500.50"`
}

type ListMonthlyTotalResponse struct {
	Data []MonthlyTotalResponse `json:"data"`
}

type MonthlyAverageResponse struct {
	Month         int     `json:"month" example:"6"`
	Year          int     `json:"year" example:"2026"`
	AverageAmount float64 `json:"average_amount" example:"187.50"`
	ExpenseCount  int     `json:"expense_count" example:"8"`
}

type ListMonthlyAverageResponse struct {
	Data []MonthlyAverageResponse `json:"data"`
}

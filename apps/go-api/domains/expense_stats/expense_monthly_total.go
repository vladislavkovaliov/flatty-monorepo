package expense_stats

type ExpenseMonthlyTotal struct {
	month      int
	year       int
	totalSpent float64
}

func NewExpenseMonthlyTotal(month, year int, totalSpent float64) *ExpenseMonthlyTotal {
	return &ExpenseMonthlyTotal{
		month:      month,
		year:       year,
		totalSpent: totalSpent,
	}
}

func (e *ExpenseMonthlyTotal) Month() int {
	return e.month
}

func (e *ExpenseMonthlyTotal) Year() int {
	return e.year
}

func (e *ExpenseMonthlyTotal) TotalSpent() float64 {
	return e.totalSpent
}

package expense_stats

type ExpenseMonthlyTotal struct {
	residentLocationID int64
	month              int
	year               int
	totalSpent         float64
}

func NewExpenseMonthlyTotal(residentLocationID int64, month, year int, totalSpent float64) *ExpenseMonthlyTotal {
	return &ExpenseMonthlyTotal{
		residentLocationID: residentLocationID,
		month:              month,
		year:               year,
		totalSpent:         totalSpent,
	}
}

func (e *ExpenseMonthlyTotal) ResidentLocationID() int64 {
	return e.residentLocationID
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

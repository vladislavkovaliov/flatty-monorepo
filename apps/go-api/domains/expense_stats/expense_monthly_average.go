package expense_stats

type ExpenseMonthlyAverage struct {
	residentLocationID int64
	month              int
	year               int
	averageAmount      float64
	expenseCount       int
}

func NewExpenseMonthlyAverage(residentLocationID int64, month, year int, averageAmount float64, expenseCount int) *ExpenseMonthlyAverage {
	return &ExpenseMonthlyAverage{
		residentLocationID: residentLocationID,
		month:              month,
		year:               year,
		averageAmount:      averageAmount,
		expenseCount:       expenseCount,
	}
}

func (e *ExpenseMonthlyAverage) ResidentLocationID() int64 {
	return e.residentLocationID
}

func (e *ExpenseMonthlyAverage) Month() int {
	return e.month
}

func (e *ExpenseMonthlyAverage) Year() int {
	return e.year
}

func (e *ExpenseMonthlyAverage) AverageAmount() float64 {
	return e.averageAmount
}

func (e *ExpenseMonthlyAverage) ExpenseCount() int {
	return e.expenseCount
}

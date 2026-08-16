package expenses

type YearAndMonth struct {
	year      int64
	month     int64
	expcenses int64
}

func (y *YearAndMonth) Year() int64 {
	return y.year
}

func (y *YearAndMonth) Month() int64 {
	return y.month
}

func (y *YearAndMonth) Expenses() int64 {
	return y.expcenses
}

func NewYearAndMonth(year, month, expenses int64) *YearAndMonth {
	return &YearAndMonth{
		year:      year,
		month:     month,
		expcenses: expenses,
	}
}

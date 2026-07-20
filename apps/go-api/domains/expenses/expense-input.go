package expenses

type ExpenseInput struct {
	residentLocationID int64
	categoryID         int64
	amount             float64
	month              int
	year               int
	description        string
}

func (i *ExpenseInput) ResidentLocationID() int64 {
	return i.residentLocationID
}

func (i *ExpenseInput) CategoryID() int64 {
	return i.categoryID
}

func (i *ExpenseInput) Amount() float64 {
	return i.amount
}

func (i *ExpenseInput) Month() int {
	return i.month
}

func (i *ExpenseInput) Year() int {
	return i.year
}

func (i *ExpenseInput) Description() string {
	return i.description
}

func NewExpenseInput(
	residentLocationID, categoryID int64,
	amount float64,
	description string,
	month, year int,
) *ExpenseInput {
	return &ExpenseInput{
		residentLocationID: residentLocationID,
		categoryID:         categoryID,
		amount:             amount,
		description:        description,
		month:              month,
		year:               year,
	}
}

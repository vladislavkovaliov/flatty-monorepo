package expenses

import "time"

type Expense struct {
	id                 int64
	residentLocationID int64
	categoryID         int64
	amount             float64
	month              int
	year               int
	createdAt          time.Time
	updatedAt          time.Time
	description        string
}

func (e *Expense) ID() int64 {
	return e.id
}

func (e *Expense) ResidentLocationID() int64 {
	return e.residentLocationID
}

func (e *Expense) CategoryID() int64 {
	return e.categoryID
}

func (e *Expense) Amount() float64 {
	return e.amount
}

func (e *Expense) Month() int {
	return e.month
}

func (e *Expense) Year() int {
	return e.year
}

func (e *Expense) CreatedAt() time.Time {
	return e.createdAt
}

func (e *Expense) UpdatedAt() time.Time {
	return e.updatedAt
}

func (e *Expense) Description() string {
	return e.description
}

func NewExpense(
	id, residentLocationID, categoryID int64,
	amount float64,
	description string,
	month, year int,
	createdAt, updatedAt time.Time,
) *Expense {
	return &Expense{
		id:                 id,
		residentLocationID: residentLocationID,
		categoryID:         categoryID,
		amount:             amount,
		description:        description,
		month:              month,
		year:               year,
		createdAt:          createdAt,
		updatedAt:          updatedAt,
	}
}

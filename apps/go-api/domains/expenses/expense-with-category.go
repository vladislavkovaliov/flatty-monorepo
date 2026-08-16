package expenses

import (
	"flatty-budget/go-api/domains/category"
	"time"
)

type ExpenseWithCategory struct {
	id                 int64
	residentLocationID int64
	categoryID         int64
	amount             float64
	month              int
	year               int
	createdAt          time.Time
	updatedAt          time.Time
	description        string
	userID             string
	category           category.Category
}

func (e *ExpenseWithCategory) ID() int64 {
	return e.id
}

func (e *ExpenseWithCategory) ResidentLocationID() int64 {
	return e.residentLocationID
}

func (e *ExpenseWithCategory) CategoryID() int64 {
	return e.categoryID
}

func (e *ExpenseWithCategory) Amount() float64 {
	return e.amount
}

func (e *ExpenseWithCategory) Month() int {
	return e.month
}

func (e *ExpenseWithCategory) Year() int {
	return e.year
}

func (e *ExpenseWithCategory) CreatedAt() time.Time {
	return e.createdAt
}

func (e *ExpenseWithCategory) UpdatedAt() time.Time {
	return e.updatedAt
}

func (e *ExpenseWithCategory) Description() string {
	return e.description
}

func (e *ExpenseWithCategory) UserID() string {
	return e.userID
}

func (e *ExpenseWithCategory) Category() category.Category {
	return e.category
}

func NewExpenseWithCategory(
	id, residentLocationID, categoryID int64,
	amount float64,
	description string,
	month, year int,
	createdAt, updatedAt time.Time,
	userID string,
	category category.Category,
) *ExpenseWithCategory {
	return &ExpenseWithCategory{
		id:                 id,
		residentLocationID: residentLocationID,
		categoryID:         categoryID,
		amount:             amount,
		description:        description,
		month:              month,
		year:               year,
		createdAt:          createdAt,
		updatedAt:          updatedAt,
		userID:             userID,
		category:           category,
	}
}

package resident_location

import (
	"time"
)

type ResidentLocation struct {
	id         int64
	userID     string
	country    string
	city       string
	postalCode string
	street     string
	house      string
	apartment  string
	createdAt  time.Time
	updatedAt  time.Time
}

func (r *ResidentLocation) ID() int64 {
	return r.id
}

func (r *ResidentLocation) UserID() string {
	return r.userID
}

func (r *ResidentLocation) Country() string {
	return r.country
}

func (r *ResidentLocation) City() string {
	return r.city
}

func (r *ResidentLocation) PostalCode() string {
	return r.postalCode
}

func (r *ResidentLocation) Street() string {
	return r.street
}

func (r *ResidentLocation) House() string {
	return r.house
}

func (r *ResidentLocation) Apartment() string {
	return r.apartment
}

func (r *ResidentLocation) CreatedAt() time.Time {
	return r.createdAt
}

func (r *ResidentLocation) UpdatedAt() time.Time {
	return r.updatedAt
}

func NewResidentLocation(id int64, userID, country, city, postalCode, street, house, apartment string, createdAt, updatedAt time.Time) *ResidentLocation {
	return &ResidentLocation{
		id:         id,
		userID:     userID,
		country:    country,
		city:       city,
		street:     street,
		postalCode: postalCode,
		house:      house,
		apartment:  apartment,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}
}

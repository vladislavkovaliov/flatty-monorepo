package resident_location

type ResidentLocationInput struct {
	country    string
	city       string
	postalCode string
	street     string
	house      string
	apartment  string
}

func (r *ResidentLocationInput) Country() string {
	return r.country
}

func (r *ResidentLocationInput) City() string {
	return r.city
}

func (r *ResidentLocationInput) PostalCode() string {
	return r.postalCode
}

func (r *ResidentLocationInput) Street() string {
	return r.street
}

func (r *ResidentLocationInput) House() string {
	return r.house
}

func (r *ResidentLocationInput) Apartment() string {
	return r.apartment
}

func NewResidentLocationInput(country, city, postalCode, street, house, apartment string) *ResidentLocationInput {
	return &ResidentLocationInput{
		country:    country,
		city:       city,
		postalCode: postalCode,
		street:     street,
		house:      house,
		apartment:  apartment,
	}
}

package category

import "time"

type Category struct {
	id          int64
	name        string
	description string
	createdAt   time.Time
	updatedAt   time.Time
}

func (c *Category) ID() int64 {
	return c.id
}

func (c *Category) Name() string {
	return c.name
}

func (c *Category) Description() string {
	return c.description
}

func (c *Category) CreatedAt() time.Time {
	return c.createdAt
}

func (c *Category) UpdatedAt() time.Time {
	return c.updatedAt
}

func NewCategory(id int64, name, description string, createdAt, updatedAt time.Time) *Category {
	return &Category{
		id:          id,
		name:        name,
		description: description,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

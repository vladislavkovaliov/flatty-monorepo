package category

type CategoryInput struct {
	name        string
	description string
}

func (c *CategoryInput) Name() string {
	return c.name
}

func (c *CategoryInput) Description() string {
	return c.description
}

func NewCategoryInput(name, description string) *CategoryInput {
	return &CategoryInput{
		name:        name,
		description: description,
	}
}

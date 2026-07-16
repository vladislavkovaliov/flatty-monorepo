package user

import (
	"time"
)

type User struct {
	id            string
	name          string
	email         string
	emailVarified bool
	image         *string
	createdAt     time.Time
	updatedAt     time.Time
}

func (u *User) ID() string {
	return u.id
}

func (u *User) Name() string {
	return u.name
}

func (u *User) Email() string {
	return u.email
}

func (u *User) EmailVerified() bool {
	return u.emailVarified
}

func (u *User) Image() *string {
	return u.image
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

func NewUser(id, name, email string, emailVarified bool, image *string, createdAt, updatedAt time.Time) *User {
	return &User{
		id:            id,
		name:          name,
		email:         email,
		emailVarified: emailVarified,
		image:         image,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

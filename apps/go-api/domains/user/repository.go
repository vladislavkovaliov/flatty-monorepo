package user

import "context"

type UserRepository interface {
	List(ctx context.Context, limit, offset int) ([]*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
}

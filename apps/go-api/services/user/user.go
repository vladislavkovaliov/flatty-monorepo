package user

import (
	"context"
	userdomain "flatty-budget/go-api/domains/user"
)

type Service struct {
	repo userdomain.UserRepository
}

func NewUserService(repo userdomain.UserRepository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]*userdomain.User, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *Service) GetUserByID(ctx context.Context, userId string) (*userdomain.User, error) {
	return s.repo.GetUserByID(ctx, userId)
}

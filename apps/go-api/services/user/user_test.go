package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"flatty-budget/go-api/domains/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockRepo implements userdomain.UserRepository.
type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) List(ctx context.Context, limit, offset int) ([]*user.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*user.User), args.Error(1)
}

func (m *mockRepo) GetUserByID(ctx context.Context, id string) (*user.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*user.User), args.Error(1)
}

func strPtr(s string) *string {
	return &s
}

func TestService_List(t *testing.T) {
	t.Parallel()

	now := time.Now()
	u1 := user.NewUser("id-1", "Alice", "alice@test.com", true, nil, now, now)
	u2 := user.NewUser("id-2", "Bob", "bob@test.com", true, strPtr("img.jpg"), now, now)

	type listCase struct {
		name          string
		limit, offset int
		repoRes       []*user.User
		repoErr       error
		want          []*user.User
		wantErr       string
	}

	cases := []listCase{
		{
			name:    "success",
			limit:   10,
			offset:  0,
			repoRes: []*user.User{u1, u2},
			want:    []*user.User{u1, u2},
		},
		{
			name:    "empty list",
			limit:   10,
			offset:  0,
			repoRes: []*user.User{},
			want:    []*user.User{},
		},
		{
			name:    "repo error",
			limit:   10,
			offset:  0,
			repoErr: errors.New("db error"),
			wantErr: "db error",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mockRepo)
			svc := NewUserService(repo)

			repo.On("List", mock.Anything, tc.limit, tc.offset).Return(tc.repoRes, tc.repoErr)

			got, err := svc.List(context.Background(), tc.limit, tc.offset)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_GetUserByID(t *testing.T) {
	t.Parallel()

	now := time.Now()
	expected := user.NewUser("id-1", "Alice", "alice@test.com", true, nil, now, now)

	type getCase struct {
		name    string
		userID  string
		repoRes *user.User
		repoErr error
		want    *user.User
		wantErr string
	}

	cases := []getCase{
		{
			name:   "success",
			userID: "id-1",
			repoRes: expected,
			want:   expected,
		},
		{
			name:    "not found",
			userID:  "non-existent",
			repoErr: errors.New("user not found"),
			wantErr: "user not found",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mockRepo)
			svc := NewUserService(repo)

			repo.On("GetUserByID", mock.Anything, tc.userID).Return(tc.repoRes, tc.repoErr)

			got, err := svc.GetUserByID(context.Background(), tc.userID)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}

			repo.AssertExpectations(t)
		})
	}
}

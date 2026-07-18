package residentlocation

import (
	"context"
	"errors"
	"testing"
	"time"

	"flatty-budget/go-api/domains/resident_location"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockRepo implements residentlocationdomain.Repository.
type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) Count(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *mockRepo) List(ctx context.Context, limit, offset int, userID string) ([]*resident_location.ResidentLocation, error) {
	args := m.Called(ctx, limit, offset, userID)
	return args.Get(0).([]*resident_location.ResidentLocation), args.Error(1)
}

func (m *mockRepo) Create(ctx context.Context, input *resident_location.ResidentLocationInput, userID string) (*resident_location.ResidentLocation, error) {
	args := m.Called(ctx, input, userID)
	return args.Get(0).(*resident_location.ResidentLocation), args.Error(1)
}

func (m *mockRepo) Update(ctx context.Context, id int64, input *resident_location.ResidentLocationInput, userID string) (*resident_location.ResidentLocation, error) {
	args := m.Called(ctx, id, input, userID)
	return args.Get(0).(*resident_location.ResidentLocation), args.Error(1)
}

func (m *mockRepo) Delete(ctx context.Context, id int64, userID string) (int64, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(int64), args.Error(1)
}

func TestService_Count(t *testing.T) {
	t.Parallel()

	type countCase struct {
		name    string
		repoErr error
		repoRes int
		want    int
		wantErr string
	}

	cases := []countCase{
		{
			name:    "success",
			repoRes: 3,
			want:    3,
		},
		{
			name:    "repo error",
			repoErr: errors.New("db error"),
			wantErr: "db error",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mockRepo)
			svc := New(repo)

			repo.On("Count", mock.Anything, "test-user-id").Return(tc.repoRes, tc.repoErr)

			got, err := svc.Count(context.Background(), "test-user-id")

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Equal(t, 0, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_List(t *testing.T) {
	t.Parallel()

	now := time.Now()
	loc1 := resident_location.NewResidentLocation(1, "test-user-id", "USA", "NYC", "10001", "Main St", "10", "1A", now, now)
	loc2 := resident_location.NewResidentLocation(2, "test-user-id", "USA", "LA", "90001", "Sunset Blvd", "20", "2B", now, now)

	type listCase struct {
		name          string
		limit, offset int

		listRepoErr error
		listRepoRes []*resident_location.ResidentLocation

		countRepoErr error
		countRepoRes int

		want      []*resident_location.ResidentLocation
		wantTotal int
		wantErr   string
	}

	cases := []listCase{
		{
			name:    "success",
			limit:   10,
			offset:  0,

			listRepoRes: []*resident_location.ResidentLocation{loc1, loc2},
			countRepoRes: 2,

			want:      []*resident_location.ResidentLocation{loc1, loc2},
			wantTotal: 2,
		},
		{
			name:    "list repo error",
			limit:   10,
			offset:  0,

			listRepoErr: errors.New("list error"),

			wantErr: "list error",
		},
		{
			name:    "count repo error",
			limit:   10,
			offset:  0,

			listRepoRes: []*resident_location.ResidentLocation{loc1},
			countRepoErr: errors.New("count error"),

			wantErr: "count error",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mockRepo)
			svc := New(repo)

			repo.On("List", mock.Anything, tc.limit, tc.offset, "test-user-id").Return(tc.listRepoRes, tc.listRepoErr)

			if tc.listRepoErr == nil {
				repo.On("Count", mock.Anything, "test-user-id").Return(tc.countRepoRes, tc.countRepoErr)
			}

			got, total, err := svc.List(context.Background(), tc.limit, tc.offset, "test-user-id")

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Nil(t, got)
				assert.Equal(t, 0, total)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
				assert.Equal(t, tc.wantTotal, total)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_Create(t *testing.T) {
	t.Parallel()

	now := time.Now()
	input := resident_location.NewResidentLocationInput("USA", "NYC", "10001", "Main St", "10", "1A")
	expected := resident_location.NewResidentLocation(1, "test-user-id", "USA", "NYC", "10001", "Main St", "10", "1A", now, now)

	type createCase struct {
		name    string
		input   *resident_location.ResidentLocationInput
		repoRes *resident_location.ResidentLocation
		repoErr error
		want    *resident_location.ResidentLocation
		wantErr string
	}

	cases := []createCase{
		{
			name:    "success",
			input:   input,
			repoRes: expected,
			want:    expected,
		},
		{
			name:    "repo error",
			input:   input,
			repoErr: errors.New("create error"),
			wantErr: "create error",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mockRepo)
			svc := New(repo)

			repo.On("Create", mock.Anything, tc.input, "test-user-id").Return(tc.repoRes, tc.repoErr)

			got, err := svc.Create(context.Background(), tc.input, "test-user-id")

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

func TestService_Update(t *testing.T) {
	t.Parallel()

	now := time.Now()
	input := resident_location.NewResidentLocationInput("Canada", "Toronto", "M5A", "Bay St", "100", "3C")
	expected := resident_location.NewResidentLocation(1, "test-user-id", "Canada", "Toronto", "M5A", "Bay St", "100", "3C", now, now)

	type updateCase struct {
		name    string
		id      int64
		input   *resident_location.ResidentLocationInput
		repoRes *resident_location.ResidentLocation
		repoErr error
		want    *resident_location.ResidentLocation
		wantErr string
	}

	cases := []updateCase{
		{
			name:    "success",
			id:      1,
			input:   input,
			repoRes: expected,
			want:    expected,
		},
		{
			name:    "repo error",
			id:      1,
			input:   input,
			repoErr: errors.New("update error"),
			wantErr: "update error",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mockRepo)
			svc := New(repo)

			repo.On("Update", mock.Anything, tc.id, tc.input, "test-user-id").Return(tc.repoRes, tc.repoErr)

			got, err := svc.Update(context.Background(), tc.id, tc.input, "test-user-id")

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

func TestService_Delete(t *testing.T) {
	t.Parallel()

	type deleteCase struct {
		name       string
		id         int64
		repoReturn int64
		repoErr    error
		want       int64
		wantErr    string
	}

	cases := []deleteCase{
		{
			name:       "success",
			id:         1,
			repoReturn: 1,
			want:       1,
		},
		{
			name:    "repo error",
			id:      1,
			repoErr: errors.New("delete error"),
			wantErr: "delete error",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mockRepo)
			svc := New(repo)

			repo.On("Delete", mock.Anything, tc.id, "test-user-id").Return(tc.repoReturn, tc.repoErr)

			got, err := svc.Delete(context.Background(), tc.id, "test-user-id")

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Equal(t, int64(0), got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}

			repo.AssertExpectations(t)
		})
	}
}

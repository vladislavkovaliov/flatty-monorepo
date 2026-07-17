package category

import (
	"context"
	"errors"
	"testing"
	"time"

	"flatty-budget/go-api/domains/category"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockRepo implements categorydomain.Repository.
type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *mockRepo) List(ctx context.Context, limit, offset int) ([]*category.Category, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*category.Category), args.Error(1)
}

func (m *mockRepo) Create(ctx context.Context, input *category.CategoryInput) (*category.Category, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*category.Category), args.Error(1)
}

func (m *mockRepo) Update(ctx context.Context, id int64, input *category.CategoryInput) (*category.Category, error) {
	args := m.Called(ctx, id, input)
	return args.Get(0).(*category.Category), args.Error(1)
}

func (m *mockRepo) Delete(ctx context.Context, id int64) (int64, error) {
	args := m.Called(ctx, id)
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
			repoRes: 5,
			want:    5,
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

			repo.On("Count", mock.Anything).Return(tc.repoRes, tc.repoErr)

			got, err := svc.Count(context.Background())

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
	cat1 := category.NewCategory(1, "cat1", "desc1", now, now)
	cat2 := category.NewCategory(2, "cat2", "desc2", now, now)

	type listCase struct {
		name          string
		limit, offset int

		listRepoErr error
		listRepoRes []*category.Category

		countRepoErr error
		countRepoRes int

		want    []*category.Category
		wantTotal int
		wantErr string
	}

	cases := []listCase{
		{
			name:    "success",
			limit:   10,
			offset:  0,

			listRepoRes: []*category.Category{cat1, cat2},
			countRepoRes: 2,

			want:      []*category.Category{cat1, cat2},
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

			listRepoRes: []*category.Category{cat1},
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

			repo.On("List", mock.Anything, tc.limit, tc.offset).Return(tc.listRepoRes, tc.listRepoErr)

			if tc.listRepoErr == nil {
				repo.On("Count", mock.Anything).Return(tc.countRepoRes, tc.countRepoErr)
			}

			got, total, err := svc.List(context.Background(), tc.limit, tc.offset)

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
	input := category.NewCategoryInput("new-cat", "new-desc")
	expected := category.NewCategory(1, "new-cat", "new-desc", now, now)

	type createCase struct {
		name    string
		input   *category.CategoryInput
		repoRes *category.Category
		repoErr error
		want    *category.Category
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

			repo.On("Create", mock.Anything, tc.input).Return(tc.repoRes, tc.repoErr)

			got, err := svc.Create(context.Background(), tc.input)

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
	input := category.NewCategoryInput("updated-cat", "updated-desc")
	expected := category.NewCategory(1, "updated-cat", "updated-desc", now, now)

	type updateCase struct {
		name    string
		id      int64
		input   *category.CategoryInput
		repoRes *category.Category
		repoErr error
		want    *category.Category
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

			repo.On("Update", mock.Anything, tc.id, tc.input).Return(tc.repoRes, tc.repoErr)

			got, err := svc.Update(context.Background(), tc.id, tc.input)

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

			repo.On("Delete", mock.Anything, tc.id).Return(tc.repoReturn, tc.repoErr)

			got, err := svc.Delete(context.Background(), tc.id)

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

package dto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserSettingsResponse_JSONTags(t *testing.T) {
	t.Parallel()

	now := time.Now()
	resp := UserSettingsResponse{
		UserID:     "user-1",
		Language:   "en",
		Theme:      "dark",
		Timezone:   "America/New_York",
		DateFormat: "MM/DD/YYYY",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	assert.Equal(t, "user-1", resp.UserID)
	assert.Equal(t, "en", resp.Language)
	assert.Equal(t, "dark", resp.Theme)
	assert.Equal(t, "America/New_York", resp.Timezone)
	assert.Equal(t, "MM/DD/YYYY", resp.DateFormat)
	assert.Equal(t, now, resp.CreatedAt)
	assert.Equal(t, now, resp.UpdatedAt)
}

func TestUpdateUserSettingsRequest_PointerFields(t *testing.T) {
	t.Parallel()

	req := UpdateUserSettingsRequest{}
	assert.Nil(t, req.Language)
	assert.Nil(t, req.Theme)
	assert.Nil(t, req.Timezone)
	assert.Nil(t, req.DateFormat)

	theme := "dark"
	req = UpdateUserSettingsRequest{Theme: &theme}
	assert.Nil(t, req.Language)
	assert.NotNil(t, req.Theme)
	assert.Equal(t, "dark", *req.Theme)
	assert.Nil(t, req.Timezone)
	assert.Nil(t, req.DateFormat)
}

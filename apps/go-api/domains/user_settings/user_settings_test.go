package user_settings

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewUserSettings(t *testing.T) {
	t.Parallel()
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	s := NewUserSettings("user-1", "en", "dark", "America/New_York", "MM/DD/YYYY", now, now)
	assert.Equal(t, "user-1", s.UserID())
	assert.Equal(t, "en", s.Language())
	assert.Equal(t, "dark", s.Theme())
	assert.Equal(t, "America/New_York", s.Timezone())
	assert.Equal(t, "MM/DD/YYYY", s.DateFormat())
	assert.True(t, now.Equal(s.CreatedAt()))
	assert.True(t, now.Equal(s.UpdatedAt()))
}

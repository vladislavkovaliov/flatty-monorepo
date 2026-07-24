package user_settings

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewUserSettingsInput(t *testing.T) {
	t.Parallel()
	input := NewUserSettingsInput("en", "dark", "America/New_York", "MM/DD/YYYY")
	assert.Equal(t, "en", input.Language())
	assert.Equal(t, "dark", input.Theme())
	assert.Equal(t, "America/New_York", input.Timezone())
	assert.Equal(t, "MM/DD/YYYY", input.DateFormat())
}

package resident_location

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResidentLocation(t *testing.T) {
	now := time.Now()
	loc := NewResidentLocation(1, "user-abc", "US", "NYC", "10001", "Main St", "10", "1A", now, now)

	assert.Equal(t, int64(1), loc.ID())
	assert.Equal(t, "user-abc", loc.UserID())
	assert.Equal(t, "US", loc.Country())
	assert.Equal(t, "NYC", loc.City())
	assert.Equal(t, "10001", loc.PostalCode())
	assert.Equal(t, "Main St", loc.Street())
	assert.Equal(t, "10", loc.House())
	assert.Equal(t, "1A", loc.Apartment())
	assert.True(t, now.Equal(loc.CreatedAt()))
	assert.True(t, now.Equal(loc.UpdatedAt()))
}

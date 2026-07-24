package user_settings

import "time"

type UserSettings struct {
	userID     string
	language   string
	theme      string
	timezone   string
	dateFormat string
	createdAt  time.Time
	updatedAt  time.Time
}

func (s *UserSettings) UserID() string { return s.userID }
func (s *UserSettings) Language() string { return s.language }
func (s *UserSettings) Theme() string { return s.theme }
func (s *UserSettings) Timezone() string { return s.timezone }
func (s *UserSettings) DateFormat() string { return s.dateFormat }
func (s *UserSettings) CreatedAt() time.Time { return s.createdAt }
func (s *UserSettings) UpdatedAt() time.Time { return s.updatedAt }

func NewUserSettings(userID, language, theme, timezone, dateFormat string, createdAt, updatedAt time.Time) *UserSettings {
	return &UserSettings{userID: userID, language: language, theme: theme, timezone: timezone, dateFormat: dateFormat, createdAt: createdAt, updatedAt: updatedAt}
}

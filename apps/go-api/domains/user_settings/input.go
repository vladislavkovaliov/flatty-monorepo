package user_settings

type UserSettingsInput struct {
	language   string
	theme      string
	timezone   string
	dateFormat string
}

func (i *UserSettingsInput) Language() string { return i.language }
func (i *UserSettingsInput) Theme() string { return i.theme }
func (i *UserSettingsInput) Timezone() string { return i.timezone }
func (i *UserSettingsInput) DateFormat() string { return i.dateFormat }

func NewUserSettingsInput(language, theme, timezone, dateFormat string) *UserSettingsInput {
	return &UserSettingsInput{language: language, theme: theme, timezone: timezone, dateFormat: dateFormat}
}

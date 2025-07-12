package users

type User struct {
	ID    string `db:"id"`
	Email string `db:"email"`
}

type Settings struct {
	WebhookURL   string
	RemindBefore int
}

type SettingsInput struct {
	WebhookURL   *string
	RemindBefore *int
}

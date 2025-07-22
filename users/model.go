package users

type User struct {
	ID    string `db:"id"`
	Email string `db:"email"`
}

type Settings struct {
	WebhookURL        string
	ReminderThreshold int
}

type SettingsInput struct {
	WebhookURL        *string
	ReminderThreshold *int
}

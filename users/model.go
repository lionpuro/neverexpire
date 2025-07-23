package users

import "github.com/lionpuro/neverexpire/notifications"

type User struct {
	ID    string `db:"id"`
	Email string `db:"email"`
}

type Settings struct {
	WebhookURL        string
	WebhookProvider   *notifications.WebhookProvider
	ReminderThreshold int
}

type SettingsInput struct {
	WebhookURL        *string
	WebhookProvider   *notifications.WebhookProvider
	ReminderThreshold *int
}

package users

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/lionpuro/neverexpire/db"
)

type Repository struct {
	db db.Connection
}

func NewRepository(conn db.Connection) *Repository {
	return &Repository{db: conn}
}

func (r *Repository) ByID(ctx context.Context, id string) (User, error) {
	rows, err := r.db.Query(ctx, `SELECT id, email FROM users WHERE id = $1`, id)
	if err != nil {
		return User{}, err
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (r *Repository) Create(ctx context.Context, id, email string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, email) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, id, email)
	return err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

func (r *Repository) Settings(ctx context.Context, userID string) (Settings, error) {
	q := `SELECT webhook_url, reminder_threshold FROM settings WHERE user_id = $1`
	row := r.db.QueryRow(ctx, q, userID)
	var vals Settings
	if err := row.Scan(&vals.WebhookURL, &vals.ReminderThreshold); err != nil {
		if db.IsErrNoRows(err) {
			return Settings{}, nil
		}
		return Settings{}, err
	}
	return vals, nil
}

func (r *Repository) SaveSettings(ctx context.Context, userID string, settings SettingsInput) (Settings, error) {
	q := `
	INSERT INTO settings (user_id, webhook_url, reminder_threshold)
	VALUES (
		$1,
		COALESCE($2, ''),
		COALESCE($3, 0)
	)
	ON CONFLICT (user_id) DO UPDATE
	SET webhook_url = COALESCE($2, settings.webhook_url),
		reminder_threshold = COALESCE($3, settings.reminder_threshold)
	RETURNING webhook_url, reminder_threshold`
	var s Settings
	row := r.db.QueryRow(ctx, q, userID, settings.WebhookURL, settings.ReminderThreshold)
	if err := row.Scan(&s.WebhookURL, &s.ReminderThreshold); err != nil {
		return Settings{}, err
	}
	return s, nil
}

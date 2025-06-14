package user

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lionpuro/neverexpire/db"
	"github.com/lionpuro/neverexpire/model"
)

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(dbpool *pgxpool.Pool) *Repository {
	return &Repository{DB: dbpool}
}

func (r *Repository) ByID(ctx context.Context, id string) (model.User, error) {
	rows, err := r.DB.Query(ctx, `SELECT id, email FROM users WHERE id = $1`, id)
	if err != nil {
		return model.User{}, err
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.User])
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (r *Repository) Create(id, email string) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	_, err := r.DB.Exec(ctx, `
		INSERT INTO users (id, email) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, id, email)
	return err
}

func (r *Repository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	_, err := r.DB.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

func (r *Repository) Settings(ctx context.Context, userID string) (model.Settings, error) {
	q := `SELECT webhook_url, remind_before FROM settings WHERE user_id = $1`
	row := r.DB.QueryRow(ctx, q, userID)
	var vals model.Settings
	if err := row.Scan(&vals.WebhookURL, &vals.RemindBefore); err != nil {
		if db.IsErrNoRows(err) {
			return model.Settings{}, nil
		}
		return model.Settings{}, err
	}
	return vals, nil
}

func (r *Repository) SaveSettings(ctx context.Context, userID string, settings model.SettingsInput) (model.Settings, error) {
	q := `
	INSERT INTO settings (user_id, webhook_url, remind_before)
	VALUES (
		$1,
		COALESCE($2, ''),
		COALESCE($3, 0)
	)
	ON CONFLICT (user_id) DO UPDATE
	SET webhook_url = COALESCE($2, settings.webhook_url),
		remind_before = COALESCE($3, settings.remind_before)
	RETURNING webhook_url, remind_before`
	var s model.Settings
	row := r.DB.QueryRow(ctx, q, userID, settings.WebhookURL, settings.RemindBefore)
	if err := row.Scan(&s.WebhookURL, &s.RemindBefore); err != nil {
		return model.Settings{}, err
	}
	return s, nil
}

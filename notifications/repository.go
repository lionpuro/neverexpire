package notifications

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/lionpuro/neverexpire/db"
)

type Repository struct {
	DB db.Connection
}

func NewRepository(conn db.Connection) *Repository {
	return &Repository{DB: conn}
}

func (r *Repository) AllDue(ctx context.Context) ([]Notification, error) {
	q := `
	SELECT
		s.webhook_url as endpoint,
		n.id,
		n.user_id,
		n.host_id,
		n.notification_type,
		n.body,
		n.due,
		n.delivered_at,
		attempts,
		deleted_after
	FROM notifications n
	INNER JOIN settings s
		ON n.user_id = s.user_id
	WHERE
		n.deleted_after > (now() at time zone 'utc')
		AND n.delivered_at IS NULL
		AND n.attempts < 3
		AND n.due <= (now() at time zone 'utc')
		AND s.webhook_url != ''
	ORDER BY due`
	rows, err := r.DB.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifs, err := pgx.CollectRows(rows, pgx.RowToStructByName[Notification])
	if err != nil {
		return nil, err
	}
	return notifs, nil
}

func (r *Repository) Create(ctx context.Context, n NotificationInput) error {
	q := `
	INSERT INTO notifications (
		user_id,
		host_id,
		notification_type,
		body,
		due,
		attempts,
		deleted_after
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (host_id, due) DO NOTHING`
	_, err := r.DB.Exec(ctx, q, n.UserID, n.HostID, n.Type, n.Body, n.Due, n.Attempts, n.DeletedAfter)
	return err
}

func (r *Repository) Update(ctx context.Context, id int, n NotificationUpdate) error {
	q := `
	UPDATE notifications
	SET delivered_at = COALESCE($1, delivered_at),
		attempts = COALESCE($2, attempts)
	WHERE id = $3`
	_, err := r.DB.Exec(ctx, q, n.DeliveredAt, n.Attempts, id)
	return err
}

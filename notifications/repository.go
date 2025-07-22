package notifications

import (
	"context"

	"github.com/lionpuro/neverexpire/db"
)

type Repository struct {
	db db.Connection
}

func NewRepository(conn db.Connection) *Repository {
	return &Repository{db: conn}
}

func (r *Repository) Upsert(ctx context.Context, n Notification) error {
	sql := `
	INSERT INTO notifications (
		user_id,
		host_id,
		notification_type,
		body,
		due,
		delivered_at,
		attempts,
		deleted_after
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (user_id, host_id, due) DO UPDATE SET
		delivered_at = EXCLUDED.delivered_at,
		attempts     = EXCLUDED.attempts
	`
	_, err := r.db.Exec(ctx, sql,
		n.UserID,
		n.HostID,
		n.Type,
		n.Body,
		n.Due,
		n.DeliveredAt,
		n.Attempts,
		n.DeletedAfter,
	)
	return err
}

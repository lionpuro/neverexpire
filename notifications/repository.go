package notifications

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/lionpuro/neverexpire/db"
	"github.com/lionpuro/neverexpire/logging"
)

type Repository struct {
	db db.Connection
}

func NewRepository(conn db.Connection) *Repository {
	return &Repository{db: conn}
}

func (r *Repository) AllByUser(ctx context.Context, uid string) ([]AppNotification, error) {
	sql := `
	SELECT
		n.id,
		n.user_id,
		n.host_id,
		n.notification_type,
		n.body,
		n.due,
		n.delivered_at,
		n.read_at,
		n.attempts,
		n.deleted_after,
		n.created_at
	FROM notifications n
	WHERE
		n.user_id = $1
		AND n.deleted_after > (now() at time zone 'utc')
	ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, sql, uid)
	if err != nil {
		return nil, err
	}
	notifs, err := pgx.CollectRows(rows, pgx.RowToStructByName[AppNotification])
	if err != nil {
		return nil, err
	}
	return notifs, nil
}

func (r *Repository) Update(ctx context.Context, uid string, input []NotificationUpdate) error {
	sql := `
	UPDATE notifications n
	SET
		body = COALESCE($3, n.body),
		delivered_at = COALESCE($4, n.delivered_at),
		read_at = COALESCE($5, n.read_at),
		attempts = COALESCE($6, n.attempts)
	WHERE id = $1 AND user_id = $2`
	tx, err := r.db.Begin(ctx)
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			logging.DefaultLogger().Error("failed to roll back tx", "error", err.Error())
		}
	}()
	if err != nil {
		return err
	}
	for _, update := range input {
		_, err := tx.Exec(ctx, sql, update.ID, uid, update.Body, update.DeliveredAt, update.ReadAt, update.Attempts)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
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

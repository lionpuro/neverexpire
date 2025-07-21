package hosts

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

func (r *Repository) ByID(ctx context.Context, userID string, id int) (Host, error) {
	row := r.db.QueryRow(ctx, `
	SELECT
		h.id,
		h.hostname,
		h.dns_names,
		h.ip_address,
		h.issued_by,
		h.status,
		h.expires_at,
		h.checked_at,
		h.latency,
		h.signature,
		h.error_message
	FROM hosts h
	INNER JOIN user_hosts uh
		ON h.id = uh.host_id
	WHERE h.id = $1 AND uh.user_id = $2`, id, userID)
	var result Host
	var errStr *string
	err := row.Scan(
		&result.ID,
		&result.Hostname,
		&result.Certificate.DNSNames,
		&result.Certificate.IP,
		&result.Certificate.IssuedBy,
		&result.Certificate.Status,
		&result.Certificate.ExpiresAt,
		&result.Certificate.CheckedAt,
		&result.Certificate.Latency,
		&result.Certificate.Signature,
		&errStr,
	)
	if err != nil {
		return Host{}, err
	}
	if errStr != nil {
		result.Certificate.Error = errors.New(*errStr)
	}

	return result, nil
}

func (r *Repository) ByName(ctx context.Context, userID, name string) (Host, error) {
	row := r.db.QueryRow(ctx, `
	SELECT
		h.id,
		h.hostname,
		h.dns_names,
		h.ip_address,
		h.issued_by,
		h.status,
		h.expires_at,
		h.checked_at,
		h.latency,
		h.signature,
		h.error_message
	FROM hosts h
	INNER JOIN user_hosts uh
		ON h.id = uh.host_id
	WHERE h.hostname = $1 AND uh.user_id = $2`, name, userID)
	var result Host
	var errStr *string
	err := row.Scan(
		&result.ID,
		&result.Hostname,
		&result.Certificate.DNSNames,
		&result.Certificate.IP,
		&result.Certificate.IssuedBy,
		&result.Certificate.Status,
		&result.Certificate.ExpiresAt,
		&result.Certificate.CheckedAt,
		&result.Certificate.Latency,
		&result.Certificate.Signature,
		&errStr,
	)
	if err != nil {
		return Host{}, err
	}
	if errStr != nil {
		result.Certificate.Error = errors.New(*errStr)
	}

	return result, nil
}

func (r *Repository) All(ctx context.Context) ([]Host, error) {
	order := fmt.Sprintf(
		"array[%d, %d, %d]",
		CertificateStatusUnknown,
		CertificateStatusOffline,
		CertificateStatusInvalid,
	)
	q := fmt.Sprintf(`
		SELECT
			id,
			hostname,
			dns_names,
			ip_address,
			issued_by,
			status,
			expires_at,
			checked_at,
			latency,
			signature,
			error_message
		FROM hosts
		ORDER BY
			array_position(%s, status),
			expires_at`,
		order,
	)
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []Host
	for rows.Next() {
		var h Host
		var errStr *string
		err := rows.Scan(
			&h.ID,
			&h.Hostname,
			&h.Certificate.DNSNames,
			&h.Certificate.IP,
			&h.Certificate.IssuedBy,
			&h.Certificate.Status,
			&h.Certificate.ExpiresAt,
			&h.Certificate.CheckedAt,
			&h.Certificate.Latency,
			&h.Certificate.Signature,
			&errStr,
		)
		if err != nil {
			return nil, err
		}
		if errStr != nil {
			h.Certificate.Error = errors.New(*errStr)
		}
		hosts = append(hosts, h)
	}

	return hosts, nil
}

func (r *Repository) Expiring(ctx context.Context) ([]HostWithUser, error) {
	q := `
	SELECT
		h.id,
		h.hostname,
		h.dns_names,
		h.ip_address,
		h.issued_by,
		h.status,
		h.expires_at,
		h.checked_at,
		h.latency,
		h.signature,
		h.error_message,
		u.id as user_id,
		u.email as user_email,
		s.webhook_url,
		s.remind_before
	FROM hosts h
	INNER JOIN user_hosts uh
		ON h.id = uh.host_id
	INNER JOIN users u
		ON uh.user_id = u.id
	INNER JOIN settings s
		ON u.id = s.user_id
	WHERE (h.expires_at - (s.remind_before * interval '1 second')) <= (now() at time zone 'utc')
	AND NOT EXISTS(
		SELECT 1 FROM notifications n
		WHERE n.host_id = h.id
		AND n.due = (h.expires_at - (s.remind_before * interval '1 second'))
		AND n.delivered_at IS NOT NULL
		AND n.attempts < 3
	)
	FOR UPDATE SKIP LOCKED`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []HostWithUser
	for rows.Next() {
		var record HostWithUser
		var errStr *string
		err := rows.Scan(
			&record.Host.ID,
			&record.Host.Hostname,
			&record.Host.Certificate.DNSNames,
			&record.Host.Certificate.IP,
			&record.Host.Certificate.IssuedBy,
			&record.Host.Certificate.Status,
			&record.Host.Certificate.ExpiresAt,
			&record.Host.Certificate.CheckedAt,
			&record.Host.Certificate.Latency,
			&record.Host.Certificate.Signature,
			&errStr,
			&record.User.ID,
			&record.User.Email,
			&record.Settings.WebhookURL,
			&record.Settings.RemindBefore,
		)
		if err != nil {
			return nil, err
		}
		if errStr != nil {
			record.Host.Certificate.Error = errors.New(*errStr)
		}
		hosts = append(hosts, record)
	}

	return hosts, nil
}

func (r *Repository) AllByUser(ctx context.Context, userID string) ([]Host, error) {
	order := fmt.Sprintf(
		"array[%d, %d, %d]",
		CertificateStatusUnknown,
		CertificateStatusOffline,
		CertificateStatusInvalid,
	)
	q := fmt.Sprintf(`
		SELECT
			h.id,
			h.hostname,
			h.dns_names,
			h.ip_address,
			h.issued_by,
			h.status,
			h.expires_at,
			h.checked_at,
			h.latency,
			h.signature,
			h.error_message
		FROM hosts h
		INNER JOIN user_hosts uh
			ON h.id = uh.host_id
		WHERE uh.user_id = $1
		ORDER BY
			array_position(%s, status),
			expires_at,
			hostname`,
		order,
	)
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []Host
	for rows.Next() {
		var h Host
		var errStr *string
		err := rows.Scan(
			&h.ID,
			&h.Hostname,
			&h.Certificate.DNSNames,
			&h.Certificate.IP,
			&h.Certificate.IssuedBy,
			&h.Certificate.Status,
			&h.Certificate.ExpiresAt,
			&h.Certificate.CheckedAt,
			&h.Certificate.Latency,
			&h.Certificate.Signature,
			&errStr,
		)
		if err != nil {
			return nil, err
		}
		if errStr != nil {
			h.Certificate.Error = errors.New(*errStr)
		}
		hosts = append(hosts, h)
	}

	return hosts, nil
}

func (r *Repository) Create(ctx context.Context, uid string, hosts []Host) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			logging.DefaultLogger().Error("failed to rollback tx", "error", err.Error())
		}
	}()

	for _, h := range hosts {
		var id int
		var errStr *string = nil
		if h.Certificate.Error != nil {
			str := h.Certificate.Error.Error()
			errStr = &str
		}
		err := tx.QueryRow(ctx, `
		INSERT INTO hosts (
			hostname,
			dns_names,
			ip_address,
			issued_by,
			status,
			expires_at,
			checked_at,
			latency,
			signature,
			error_message
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (hostname) DO UPDATE SET
			dns_names      = EXCLUDED.dns_names,
			ip_address     = EXCLUDED.ip_address,
			issued_by      = EXCLUDED.issued_by,
			status         = EXCLUDED.status,
			expires_at     = EXCLUDED.expires_at,
			checked_at     = EXCLUDED.checked_at,
			latency        = EXCLUDED.latency,
			signature      = EXCLUDED.signature,
			error_message  = EXCLUDED.error_message
		RETURNING id
		`,
			h.Hostname,
			h.Certificate.DNSNames,
			h.Certificate.IP,
			h.Certificate.IssuedBy,
			h.Certificate.Status,
			h.Certificate.ExpiresAt,
			h.Certificate.CheckedAt,
			h.Certificate.Latency,
			h.Certificate.Signature,
			errStr,
		).Scan(&id)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx,
			`INSERT INTO user_hosts (host_id, user_id) VALUES ($1, $2)`,
			id, uid,
		)
		if err != nil {
			str := `duplicate key value violates unique constraint "uq_user_hosts_user_id_host_id"`
			if strings.Contains(err.Error(), str) {
				return fmt.Errorf("already tracking %s", h.Hostname)
			}
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, uid string, id int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			logging.DefaultLogger().Error("failed to rollback tx", "error", err.Error())
		}
	}()
	_, err = tx.Exec(ctx, `DELETE FROM user_hosts WHERE host_id = $1 AND user_id = $2`, id, uid)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		DELETE FROM hosts
		WHERE id = $1
		AND NOT EXISTS (
			SELECT 1 FROM user_hosts uh
			WHERE uh.host_id = $1
		)`, id)
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, hosts []Host) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			logging.DefaultLogger().Error("failed to rollback tx", "error", err.Error())
		}
	}()

	for _, h := range hosts {
		var errStr *string = nil
		if h.Certificate.Error != nil {
			str := h.Certificate.Error.Error()
			errStr = &str
		}
		_, err := tx.Exec(ctx, `
		UPDATE hosts
		SET
			dns_names = $1,
			ip_address = $2,
			issued_by = $3,
			status = $4,
			expires_at = $5,
			checked_at = $6,
			latency = $7,
			signature = $8,
			error_message = $9,
			updated_at = (now() at time zone 'utc')
		WHERE id = $10
		`,
			h.Certificate.DNSNames,
			h.Certificate.IP,
			h.Certificate.IssuedBy,
			h.Certificate.Status,
			h.Certificate.ExpiresAt,
			h.Certificate.CheckedAt,
			h.Certificate.Latency,
			h.Certificate.Signature,
			errStr,
			h.ID,
		)
		if err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

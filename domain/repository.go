package domain

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lionpuro/neverexpire/db"
	"github.com/lionpuro/neverexpire/logging"
)

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(dbpool *pgxpool.Pool) *Repository {
	return &Repository{DB: dbpool}
}

func (r *Repository) ByID(ctx context.Context, userID string, id int) (Domain, error) {
	row := r.DB.QueryRow(ctx, `
	SELECT
		d.id,
		d.domain_name,
		d.dns_names,
		d.ip_address,
		d.issued_by,
		d.status,
		d.expires_at,
		d.checked_at,
		d.latency,
		d.signature,
		d.error_message
	FROM domains d
	INNER JOIN user_domains ud
		ON d.id = ud.domain_id
	WHERE d.id = $1 AND ud.user_id = $2`, id, userID)
	var result Domain
	var errStr *string
	err := row.Scan(
		&result.ID,
		&result.DomainName,
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
		return Domain{}, err
	}
	if errStr != nil {
		result.Certificate.Error = errors.New(*errStr)
	}

	return result, nil
}

func (r *Repository) All(ctx context.Context) ([]Domain, error) {
	order := fmt.Sprintf(
		"array[%d, %d, %d]",
		CertificateStatusUnknown,
		CertificateStatusOffline,
		CertificateStatusInvalid,
	)
	q := fmt.Sprintf(`
		SELECT
			id,
			domain_name,
			dns_names,
			ip_address,
			issued_by,
			status,
			expires_at,
			checked_at,
			latency,
			signature,
			error_message
		FROM domains
		ORDER BY
			array_position(%s, status),
			expires_at`,
		order,
	)
	rows, err := r.DB.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []Domain
	for rows.Next() {
		var d Domain
		var errStr *string
		err := rows.Scan(
			&d.ID,
			&d.DomainName,
			&d.Certificate.DNSNames,
			&d.Certificate.IP,
			&d.Certificate.IssuedBy,
			&d.Certificate.Status,
			&d.Certificate.ExpiresAt,
			&d.Certificate.CheckedAt,
			&d.Certificate.Latency,
			&d.Certificate.Signature,
			&errStr,
		)
		if err != nil {
			return nil, err
		}
		if errStr != nil {
			d.Certificate.Error = errors.New(*errStr)
		}
		domains = append(domains, d)
	}

	return domains, nil
}

func (r *Repository) Expiring(ctx context.Context) ([]DomainWithUser, error) {
	q := `
	SELECT
		d.id,
		d.domain_name,
		d.dns_names,
		d.ip_address,
		d.issued_by,
		d.status,
		d.expires_at,
		d.checked_at,
		d.latency,
		d.signature,
		d.error_message,
		u.id as user_id,
		u.email as user_email,
		s.webhook_url,
		s.remind_before
	FROM domains d
	INNER JOIN user_domains ud
		ON d.id = ud.domain_id
	INNER JOIN users u
		ON ud.user_id = u.id
	INNER JOIN settings s
		ON u.id = s.user_id
	WHERE (d.expires_at - (s.remind_before * interval '1 second')) <= (now() at time zone 'utc')
	AND NOT EXISTS(
		SELECT 1 FROM notifications n
		WHERE n.domain_id = d.id
		AND n.due = (d.expires_at - (s.remind_before * interval '1 second'))
		AND n.delivered_at IS NOT NULL
		AND n.attempts < 3
	)
	FOR UPDATE SKIP LOCKED`
	rows, err := r.DB.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []DomainWithUser
	for rows.Next() {
		var record DomainWithUser
		var errStr *string
		err := rows.Scan(
			&record.Domain.ID,
			&record.Domain.DomainName,
			&record.Domain.Certificate.DNSNames,
			&record.Domain.Certificate.IP,
			&record.Domain.Certificate.IssuedBy,
			&record.Domain.Certificate.Status,
			&record.Domain.Certificate.ExpiresAt,
			&record.Domain.Certificate.CheckedAt,
			&record.Domain.Certificate.Latency,
			&record.Domain.Certificate.Signature,
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
			record.Domain.Certificate.Error = errors.New(*errStr)
		}
		domains = append(domains, record)
	}

	return domains, nil
}

func (r *Repository) AllByUser(ctx context.Context, userID string) ([]Domain, error) {
	order := fmt.Sprintf(
		"array[%d, %d, %d]",
		CertificateStatusUnknown,
		CertificateStatusOffline,
		CertificateStatusInvalid,
	)
	q := fmt.Sprintf(`
		SELECT
			d.id,
			d.domain_name,
			d.dns_names,
			d.ip_address,
			d.issued_by,
			d.status,
			d.expires_at,
			d.checked_at,
			d.latency,
			d.signature,
			d.error_message
		FROM domains d
		INNER JOIN user_domains ud
			ON d.id = ud.domain_id
		WHERE ud.user_id = $1
		ORDER BY
			array_position(%s, status),
			expires_at,
			domain_name`,
		order,
	)
	rows, err := r.DB.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []Domain
	for rows.Next() {
		var d Domain
		var errStr *string
		err := rows.Scan(
			&d.ID,
			&d.DomainName,
			&d.Certificate.DNSNames,
			&d.Certificate.IP,
			&d.Certificate.IssuedBy,
			&d.Certificate.Status,
			&d.Certificate.ExpiresAt,
			&d.Certificate.CheckedAt,
			&d.Certificate.Latency,
			&d.Certificate.Signature,
			&errStr,
		)
		if err != nil {
			return nil, err
		}
		if errStr != nil {
			d.Certificate.Error = errors.New(*errStr)
		}
		domains = append(domains, d)
	}

	return domains, nil
}

func (r *Repository) Create(uid string, domains []Domain) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			logging.DefaultLogger().Error("failed to rollback tx", "error", err.Error())
		}
	}()

	for _, d := range domains {
		var id int
		var errStr *string = nil
		if d.Certificate.Error != nil {
			str := d.Certificate.Error.Error()
			errStr = &str
		}
		err := tx.QueryRow(ctx, `
		INSERT INTO domains (
			domain_name,
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
		ON CONFLICT (domain_name) DO UPDATE SET
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
			d.DomainName,
			d.Certificate.DNSNames,
			d.Certificate.IP,
			d.Certificate.IssuedBy,
			d.Certificate.Status,
			d.Certificate.ExpiresAt,
			d.Certificate.CheckedAt,
			d.Certificate.Latency,
			d.Certificate.Signature,
			errStr,
		).Scan(&id)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx,
			`INSERT INTO user_domains (domain_id, user_id) VALUES ($1, $2)`,
			id, uid,
		)
		if err != nil {
			str := `duplicate key value violates unique constraint "uq_user_domains_user_id_domain_id"`
			if strings.Contains(err.Error(), str) {
				return fmt.Errorf("already tracking %s", d.DomainName)
			}
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *Repository) Delete(uid string, id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			logging.DefaultLogger().Error("failed to rollback tx", "error", err.Error())
		}
	}()
	_, err = tx.Exec(ctx, `DELETE FROM user_domains WHERE domain_id = $1 AND user_id = $2`, id, uid)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		DELETE FROM domains
		WHERE id = $1
		AND NOT EXISTS (
			SELECT 1 FROM user_domains ud
			WHERE ud.domain_id = $1
		)`, id)
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, domains []Domain) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			logging.DefaultLogger().Error("failed to rollback tx", "error", err.Error())
		}
	}()

	for _, d := range domains {
		var errStr *string = nil
		if d.Certificate.Error != nil {
			str := d.Certificate.Error.Error()
			errStr = &str
		}
		_, err := tx.Exec(ctx, `
		UPDATE domains
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
			d.Certificate.DNSNames,
			d.Certificate.IP,
			d.Certificate.IssuedBy,
			d.Certificate.Status,
			d.Certificate.ExpiresAt,
			d.Certificate.CheckedAt,
			d.Certificate.Latency,
			d.Certificate.Signature,
			errStr,
			d.ID,
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

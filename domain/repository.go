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
	"github.com/lionpuro/neverexpire/model"
)

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(dbpool *pgxpool.Pool) *Repository {
	return &Repository{DB: dbpool}
}

func (r *Repository) ByID(ctx context.Context, userID string, id int) (model.Domain, error) {
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
		d.signature
	FROM domains d
	INNER JOIN user_domains ud
		ON d.id = ud.domain_id
	WHERE d.id = $1 AND ud.user_id = $2`, id, userID)
	var result model.Domain
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
	)
	if err != nil {
		return model.Domain{}, err
	}

	return result, nil
}

func (r *Repository) All(ctx context.Context) ([]model.Domain, error) {
	order := fmt.Sprintf(
		"array[%d, %d, %d]",
		model.CertificateStatusUnknown,
		model.CertificateStatusOffline,
		model.CertificateStatusInvalid,
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
			signature
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

	var domains []model.Domain
	for rows.Next() {
		var d model.Domain
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
		)
		if err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}

	return domains, nil
}

func (r *Repository) Notifiable(ctx context.Context) ([]model.DomainWithUser, error) {
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

	var domains []model.DomainWithUser
	for rows.Next() {
		var record model.DomainWithUser
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
			&record.User.ID,
			&record.User.Email,
			&record.Settings.WebhookURL,
			&record.Settings.RemindBefore,
		)
		if err != nil {
			return nil, err
		}
		domains = append(domains, record)
	}

	return domains, nil
}

func (r *Repository) AllByUser(ctx context.Context, userID string) ([]model.Domain, error) {
	order := fmt.Sprintf(
		"array[%d, %d, %d]",
		model.CertificateStatusUnknown,
		model.CertificateStatusOffline,
		model.CertificateStatusInvalid,
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
			d.signature
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

	var domains []model.Domain
	for rows.Next() {
		var d model.Domain
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
		)
		if err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}

	return domains, nil
}

func (r *Repository) Create(uid string, domains []model.Domain) error {
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
			signature
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (domain_name) DO UPDATE SET
			dns_names  = EXCLUDED.dns_names,
			ip_address = EXCLUDED.ip_address,
			issued_by  = EXCLUDED.issued_by,
			status     = EXCLUDED.status,
			expires_at = EXCLUDED.expires_at,
			checked_at = EXCLUDED.checked_at,
			latency    = EXCLUDED.latency,
			signature  = EXCLUDED.signature
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

func (r *Repository) Update(ctx context.Context, domains []model.Domain) error {
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
			updated_at = (now() at time zone 'utc')
		WHERE id = $9
		`,
			d.Certificate.DNSNames,
			d.Certificate.IP,
			d.Certificate.IssuedBy,
			d.Certificate.Status,
			d.Certificate.ExpiresAt,
			d.Certificate.CheckedAt,
			d.Certificate.Latency,
			d.Certificate.Signature,
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

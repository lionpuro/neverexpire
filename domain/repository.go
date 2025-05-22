package domain

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lionpuro/trackcerts/db"
	"github.com/lionpuro/trackcerts/model"
)

type Repository interface {
	ByID(ctx context.Context, userID string, id int) (model.Domain, error)
	All(ctx context.Context) ([]model.Domain, error)
	AllByUser(ctx context.Context, userID string) ([]model.Domain, error)
	Create(d model.Domain) error
	Update(d model.Domain) (model.Domain, error)
	UpdateMultiple(ctx context.Context, domains []model.Domain) error
	Delete(userID string, id int) error
}

type DomainRepository struct {
	DB *pgxpool.Pool
}

func (r *DomainRepository) ByID(ctx context.Context, userID string, id int) (model.Domain, error) {
	row := r.DB.QueryRow(ctx, `
	SELECT
		id,
		user_id,
		domain_name,
		dns_names,
		ip_address,
		issued_by,
		status,
		expires_at,
		checked_at,
		latency
	FROM domains
	WHERE id = $1 AND user_id = $2`, id, userID)
	var result model.Domain
	err := row.Scan(
		&result.ID,
		&result.UserID,
		&result.DomainName,
		&result.Certificate.DNSNames,
		&result.Certificate.IP,
		&result.Certificate.Issuer,
		&result.Certificate.Status,
		&result.Certificate.Expires,
		&result.Certificate.CheckedAt,
		&result.Certificate.Latency,
	)
	if err != nil {
		return model.Domain{}, err
	}

	return result, nil
}

func (r *DomainRepository) All(ctx context.Context) ([]model.Domain, error) {
	q := `
	SELECT
		id,
		user_id,
		domain_name,
		dns_names,
		ip_address,
		issued_by,
		status,
		expires_at,
		checked_at,
		latency
	FROM domains
	ORDER BY
		array_position(array['offline', 'invalid', 'expiring', 'healthy'], status),
		expires_at`
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
			&d.UserID,
			&d.DomainName,
			&d.Certificate.DNSNames,
			&d.Certificate.IP,
			&d.Certificate.Issuer,
			&d.Certificate.Status,
			&d.Certificate.Expires,
			&d.Certificate.CheckedAt,
			&d.Certificate.Latency,
		)
		if err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}

	return domains, nil
}

func (r *DomainRepository) AllByUser(ctx context.Context, userID string) ([]model.Domain, error) {
	q := `
	SELECT
		id,
		user_id,
		domain_name,
		dns_names,
		ip_address,
		issued_by,
		status,
		expires_at,
		checked_at,
		latency
	FROM domains WHERE user_id = $1
	ORDER BY
		array_position(array['offline', 'invalid', 'expiring', 'healthy'], status),
		expires_at`
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
			&d.UserID,
			&d.DomainName,
			&d.Certificate.DNSNames,
			&d.Certificate.IP,
			&d.Certificate.Issuer,
			&d.Certificate.Status,
			&d.Certificate.Expires,
			&d.Certificate.CheckedAt,
			&d.Certificate.Latency,
		)
		if err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}

	return domains, nil
}

func (r *DomainRepository) Create(d model.Domain) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	_, err := r.DB.Exec(ctx, `
	INSERT INTO domains (
		user_id,
		domain_name,
		dns_names,
		ip_address,
		issued_by,
		status,
		expires_at,
		checked_at,
		latency
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		d.UserID,
		d.DomainName,
		d.Certificate.DNSNames,
		d.Certificate.IP,
		d.Certificate.Issuer,
		d.Certificate.Status,
		d.Certificate.Expires,
		d.Certificate.CheckedAt,
		d.Certificate.Latency,
	)
	return err
}

func (r *DomainRepository) Update(d model.Domain) (model.Domain, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	row := r.DB.QueryRow(ctx, `
	UPDATE domains
	SET
		dns_names = $3,
		ip_address = $4,
		issued_by = $5,
		status = $6,
		expires_at = $7,
		checked_at = $8,
		latency = $9,
		updated_at = (now() at time zone 'utc')
	WHERE id = $1 AND user_id = $2
	RETURNING
		id,
		user_id,
		domain_name,
		dns_names,
		ip_address,
		issued_by,
		status,
		expires_at,
		checked_at,
		latency
	`,
		d.ID,
		d.UserID,
		d.Certificate.DNSNames,
		d.Certificate.IP,
		d.Certificate.Issuer,
		d.Certificate.Status,
		d.Certificate.Expires,
		d.Certificate.CheckedAt,
		d.Certificate.Latency,
	)
	var result model.Domain
	err := row.Scan(
		&result.ID,
		&result.UserID,
		&result.DomainName,
		&result.Certificate.DNSNames,
		&result.Certificate.IP,
		&result.Certificate.Issuer,
		&result.Certificate.Status,
		&result.Certificate.Expires,
		&result.Certificate.CheckedAt,
		&result.Certificate.Latency,
	)
	if err != nil {
		return model.Domain{}, err
	}

	return result, nil
}

func (r *DomainRepository) Delete(userID string, id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	_, err := r.DB.Exec(ctx, `DELETE FROM domains WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	return nil
}

func (r *DomainRepository) UpdateMultiple(ctx context.Context, domains []model.Domain) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

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
			updated_at = (now() at time zone 'utc')
		WHERE id = $8
		`,
			d.Certificate.DNSNames,
			d.Certificate.IP,
			d.Certificate.Issuer,
			d.Certificate.Status,
			d.Certificate.Expires,
			d.Certificate.CheckedAt,
			d.Certificate.Latency,
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

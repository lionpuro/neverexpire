package domain

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lionpuro/trackcerts/db"
	"github.com/lionpuro/trackcerts/model"
)

type Repository interface {
	ByID(ctx context.Context, userID string, id int) (model.Domain, error)
	All(ctx context.Context, userID string) ([]model.Domain, error)
	Create(d model.Domain) error
	Update(d model.Domain) (model.Domain, error)
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
		&result.CertificateInfo.DNSNames,
		&result.CertificateInfo.IP,
		&result.CertificateInfo.Issuer,
		&result.CertificateInfo.Status,
		&result.CertificateInfo.Expires,
		&result.CertificateInfo.CheckedAt,
		&result.CertificateInfo.Latency,
	)
	if err != nil {
		return model.Domain{}, err
	}

	return result, nil
}

func (r *DomainRepository) All(ctx context.Context, userID string) ([]model.Domain, error) {
	rows, err := r.DB.Query(ctx, `
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
		expires_at
	`, userID)
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
			&d.CertificateInfo.DNSNames,
			&d.CertificateInfo.IP,
			&d.CertificateInfo.Issuer,
			&d.CertificateInfo.Status,
			&d.CertificateInfo.Expires,
			&d.CertificateInfo.CheckedAt,
			&d.CertificateInfo.Latency,
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
		d.DNSNames,
		d.IP,
		d.Issuer,
		d.Status,
		d.Expires,
		d.CheckedAt,
		d.Latency,
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
		d.DNSNames,
		d.IP,
		d.Issuer,
		d.Status,
		d.Expires,
		d.CheckedAt,
		d.Latency,
	)
	var result model.Domain
	err := row.Scan(
		&result.ID,
		&result.UserID,
		&result.DomainName,
		&result.CertificateInfo.DNSNames,
		&result.CertificateInfo.IP,
		&result.CertificateInfo.Issuer,
		&result.CertificateInfo.Status,
		&result.CertificateInfo.Expires,
		&result.CertificateInfo.CheckedAt,
		&result.CertificateInfo.Latency,
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

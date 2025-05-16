package db

import (
	"context"

	"github.com/lionpuro/trackcerts/model"
)

func (s *Service) DomainByID(ctx context.Context, id, userID string) (model.Domain, error) {
	row := s.DB.QueryRow(ctx, `
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

func (s *Service) Domains(ctx context.Context, userID string) ([]model.Domain, error) {
	rows, err := s.DB.Query(ctx, `
	SELECT
		id,
		user_id,
		domain_name,
		dns_names,
		issued_by,
		status,
		expires_at,
		checked_at
	FROM domains WHERE user_id = $1
	ORDER By expires_at
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
			&d.DNSNames,
			&d.Issuer,
			&d.Status,
			&d.Expires,
			&d.CheckedAt,
		)
		if err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}

	return domains, nil
}

func (s *Service) CreateDomain(d model.Domain) error {
	ctx, cancel := newContext()
	defer cancel()
	_, err := s.DB.Exec(ctx, `
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

func (s *Service) UpdateDomainInfo(id int, userID string, info model.CertificateInfo) error {
	ctx, cancel := newContext()
	defer cancel()
	_, err := s.DB.Exec(ctx, `
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
	WHERE id = $1 AND user_id = $2`,
		id,
		userID,
		info.DNSNames,
		info.IP,
		info.Issuer,
		info.Status,
		info.Expires,
		info.CheckedAt,
		info.Latency,
	)

	return err
}

func (s *Service) DeleteDomain(id, userID string) error {
	ctx, cancel := newContext()
	defer cancel()
	_, err := s.DB.Exec(ctx, `DELETE FROM domains WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	return nil
}

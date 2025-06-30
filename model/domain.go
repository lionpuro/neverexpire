package model

import "time"

type Domain struct {
	ID          int    `db:"id"`
	DomainName  string `db:"domain_name"`
	Certificate CertificateInfo
}

type CertificateInfo struct {
	DNSNames  string            `db:"dns_names"`
	IP        string            `db:"ip_address"`
	IssuedBy  string            `db:"issued_by"`
	Expires   *time.Time        `db:"expires_at"`
	Status    CertificateStatus `db:"status"`
	CheckedAt time.Time         `db:"checked_at"`
	Latency   int               `db:"latency"`
	Signature string            `db:"signature"`
	Error     error             `db:"-"`
}

type DomainWithUser struct {
	Domain   Domain
	User     User
	Settings Settings
}

func (c CertificateInfo) DaysLeft() int {
	if c.Expires == nil {
		return 0
	}
	now := time.Now().UTC()
	if c.Expires.Before(now) {
		return 0
	}
	diff := c.Expires.Sub(now)
	return int(diff.Hours() / 24)
}

type CertificateStatus int

const (
	CertificateStatusUnknown CertificateStatus = iota
	CertificateStatusOffline
	CertificateStatusInvalid
	CertificateStatusHealthy
)

func (s CertificateStatus) String() string {
	switch s {
	case CertificateStatusOffline:
		return "offline"
	case CertificateStatusInvalid:
		return "invalid"
	case CertificateStatusHealthy:
		return "healthy"
	default:
		return "unknown"
	}
}

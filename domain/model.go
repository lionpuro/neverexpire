package domain

import (
	"time"

	"github.com/lionpuro/neverexpire/user"
)

type Domain struct {
	ID          int    `db:"id"`
	DomainName  string `db:"domain_name"`
	Certificate CertificateInfo
}

type CertificateInfo struct {
	DNSNames  string            `db:"dns_names"`
	IP        string            `db:"ip_address"`
	IssuedBy  string            `db:"issued_by"`
	ExpiresAt *time.Time        `db:"expires_at"`
	Status    CertificateStatus `db:"status"`
	CheckedAt time.Time         `db:"checked_at"`
	Latency   int               `db:"latency"`
	Signature string            `db:"signature"`
	Error     error             `db:"-"`
}

type DomainWithUser struct {
	Domain   Domain
	User     user.User
	Settings user.Settings
}

func (c CertificateInfo) TimeLeft() time.Duration {
	exp := c.ExpiresAt
	now := time.Now().UTC()
	if exp == nil || exp.Before(now) {
		return 0
	}
	return exp.Sub(now)
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

package hosts

import (
	"time"

	"github.com/lionpuro/neverexpire/users"
)

type Host struct {
	ID          int    `db:"id"`
	HostName    string `db:"hostname"`
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

type HostWithUser struct {
	Host     Host
	User     users.User
	Settings users.Settings
}

type APIModel struct {
	HostName  string     `json:"hostname"`
	Issuer    *string    `json:"issuer"`
	ExpiresAt *time.Time `json:"expires_at"`
	CheckedAt time.Time  `json:"checked_at"`
	Error     *string    `json:"error"`
}

func ToAPIModel(h Host) APIModel {
	var errMsg *string
	if err := h.Certificate.Error; err != nil {
		msg := err.Error()
		errMsg = &msg
	}
	result := APIModel{
		HostName:  h.HostName,
		Issuer:    &h.Certificate.IssuedBy,
		ExpiresAt: h.Certificate.ExpiresAt,
		CheckedAt: h.Certificate.CheckedAt,
		Error:     errMsg,
	}
	if iss := h.Certificate.IssuedBy; iss == "n/a" || iss == "" {
		result.Issuer = nil
	}
	return result
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

package model

import "time"

type Domain struct {
	ID          int    `db:"id"`
	UserID      string `db:"user_id"`
	DomainName  string `db:"domain_name"`
	Certificate CertificateInfo
}

type CertificateInfo struct {
	DNSNames  string     `db:"dns_names"`
	IP        string     `db:"ip_address"`
	IssuedBy  string     `db:"issued_by"`
	Expires   *time.Time `db:"expires_at"`
	Status    string     `db:"status"`
	CheckedAt time.Time  `db:"checked_at"`
	Latency   int        `db:"latency"`
	Signature string     `db:"signature"`
	Error     error      `db:"-"`
}

type DomainWithSettings struct {
	Domain   Domain
	Settings Settings
}

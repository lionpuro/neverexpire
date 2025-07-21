package testutils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/lionpuro/neverexpire/hosts"
	"github.com/lionpuro/neverexpire/users"
)

func RandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func NewTestUser() (users.User, error) {
	id, err := RandomString(24)
	if err != nil {
		return users.User{}, err
	}
	email := id[:8] + "@example.com"
	user := users.User{ID: id, Email: email}
	return user, nil
}

func NewTestHost(exp *time.Time, status hosts.CertificateStatus, hostErr error) (hosts.Host, error) {
	str, err := RandomString(8)
	if err != nil {
		return hosts.Host{}, err
	}
	name := str + ".example.com"
	host := hosts.Host{
		Hostname: name,
		Certificate: hosts.CertificateInfo{
			DNSNames:  name,
			IP:        "",
			IssuedBy:  "Example Certs",
			ExpiresAt: exp,
			Status:    status,
			Latency:   1,
			Signature: str,
			Error:     hostErr,
		},
	}
	return host, nil
}

func NewTestHosts() ([]hosts.Host, error) {
	input := []struct {
		expires *time.Time
		status  hosts.CertificateStatus
		err     error
	}{
		// healthy
		{
			expires: addTime(50 * 24 * time.Hour),
			status:  hosts.CertificateStatusHealthy,
		},
		{
			expires: addTime(10 * 24 * time.Hour),
			status:  hosts.CertificateStatusHealthy,
		},
		{
			expires: addTime(5 * 24 * time.Hour),
			status:  hosts.CertificateStatusHealthy,
		},
		{
			expires: addTime(20 * time.Hour),
			status:  hosts.CertificateStatusHealthy,
		},
		// invalid
		{
			expires: addTime(62 * 24 * time.Hour),
			status:  hosts.CertificateStatusInvalid,
			err:     hosts.ErrCertInvalid,
		},
		{
			expires: nil,
			status:  hosts.CertificateStatusInvalid,
			err:     hosts.ErrCertInvalid,
		},
		// offline
		{
			expires: nil,
			status:  hosts.CertificateStatusOffline,
		},
	}
	var result []hosts.Host
	for _, in := range input {
		h, err := NewTestHost(in.expires, in.status, in.err)
		if err != nil {
			return nil, fmt.Errorf("failed to create fake host data: %v", err)
		}
		result = append(result, h)
	}
	return result, nil
}

func addTime(d time.Duration) *time.Time {
	t := time.Now().UTC().Add(d)
	return &t
}

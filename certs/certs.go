package certs

import (
	"context"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"
	"time"

	"github.com/lionpuro/neverexpire/logging"
	"github.com/lionpuro/neverexpire/model"
)

const (
	StatusHealthy  = "healthy"
	StatusExpiring = "expiring"
	StatusInvalid  = "invalid"
	StatusOffline  = "offline"
)

func FetchCert(ctx context.Context, domain string) (*model.CertificateInfo, error) {
	errch := make(chan error, 1)
	result := make(chan model.CertificateInfo, 1)
	go func() {
		start := time.Now().UTC()
		conn, err := tls.Dial("tcp", fmt.Sprintf("%s:443", domain), &tls.Config{})
		if err != nil {
			if strings.Contains(err.Error(), "tls: failed to verify") {
				result <- model.CertificateInfo{
					Status:    StatusInvalid,
					IssuedBy:  "n/a",
					CheckedAt: start,
					Error:     err,
				}
				return
			}
			if strings.Contains(err.Error(), "no such host") || strings.Contains(err.Error(), "Temporary failure in name resolution") {
				result <- model.CertificateInfo{
					Status:    StatusOffline,
					IssuedBy:  "n/a",
					CheckedAt: start,
					Error:     err,
				}
				return
			}
			errch <- err
			return
		}
		defer func() {
			if err := conn.Close(); err != nil {
				logging.DefaultLogger().Error("error closing connection", "error", err.Error())
			}
		}()

		state := conn.ConnectionState()
		cert := state.PeerCertificates[0]
		result <- model.CertificateInfo{
			DNSNames:  strings.Join(cert.DNSNames, ", "),
			IP:        conn.RemoteAddr().String(),
			Expires:   &cert.NotAfter,
			IssuedBy:  cert.Issuer.Organization[0],
			CheckedAt: start,
			Status:    StatusString(cert.NotAfter),
			Latency:   int(time.Since(start).Milliseconds()),
			Signature: fingerprint(cert),
		}
	}()

	select {
	case err := <-errch:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-result:
		return &result, nil
	}
}

func fingerprint(cert *x509.Certificate) string {
	fingerprint := sha1.Sum(cert.Raw)
	return fmt.Sprintf("%x", fingerprint)
}

func StatusString(expires time.Time) string {
	switch {
	case expires.Before(time.Now().UTC()):
		return StatusInvalid
	case expires.Before(time.Now().UTC().AddDate(0, 0, 14)):
		return StatusExpiring
	default:
		return StatusHealthy
	}
}

func DaysLeft(expires time.Time) int {
	now := time.Now().UTC()
	if expires.Before(now) {
		return 0
	}
	diff := expires.Sub(now)
	return int(diff.Hours() / 24)
}

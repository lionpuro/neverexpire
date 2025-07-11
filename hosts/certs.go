package hosts

import (
	"context"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lionpuro/neverexpire/logging"
)

func FetchCert(ctx context.Context, hostname string) (*CertificateInfo, error) {
	errch := make(chan error, 1)
	result := make(chan CertificateInfo, 1)
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	go func() {
		start := time.Now().UTC()
		conn, err := tls.Dial("tcp", fmt.Sprintf("%s:443", hostname), nil)
		if err != nil {
			status := errorStatus(err)
			result <- CertificateInfo{
				Status:    status,
				IssuedBy:  "n/a",
				CheckedAt: start,
				Error:     mapError(err),
			}
			return
		}
		defer func() {
			if err := conn.Close(); err != nil {
				logging.DefaultLogger().Error("error closing connection", "error", err.Error())
			}
		}()
		state := conn.ConnectionState()
		cert := state.PeerCertificates[0]
		status := CertificateStatusInvalid
		if cert.NotAfter.After(time.Now().UTC()) {
			status = CertificateStatusHealthy
		}
		result <- CertificateInfo{
			DNSNames:  strings.Join(cert.DNSNames, ", "),
			IP:        conn.RemoteAddr().String(),
			ExpiresAt: &cert.NotAfter,
			IssuedBy:  cert.Issuer.Organization[0],
			CheckedAt: start,
			Status:    status,
			Latency:   int(time.Since(start).Milliseconds()),
			Signature: fingerprint(cert),
		}
	}()

	select {
	case err := <-errch:
		return nil, err
	case <-ctx.Done():
		if err := ctx.Err(); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				cert := &CertificateInfo{
					Status:    CertificateStatusOffline,
					IssuedBy:  "n/a",
					CheckedAt: time.Now().UTC(),
					Error:     ErrConnTimedout,
				}
				return cert, nil
			}
		}
		return nil, ctx.Err()
	case result := <-result:
		return &result, nil
	}
}

func fingerprint(cert *x509.Certificate) string {
	fingerprint := sha1.Sum(cert.Raw)
	return fmt.Sprintf("%x", fingerprint)
}

func errorStatus(err error) CertificateStatus {
	switch {
	case strings.Contains(err.Error(), "tls: failed to verify"):
		return CertificateStatusInvalid
	case
		strings.Contains(err.Error(), "connection refused"),
		strings.Contains(err.Error(), "no such host"),
		strings.Contains(err.Error(), "Temporary failure in name resolution"):
		return CertificateStatusOffline
	}
	return CertificateStatusUnknown
}

func mapError(err error) Error {
	contains := func(s string) bool {
		return strings.Contains(err.Error(), s)
	}
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return ErrConnTimedout
	case contains("tls: failed to verify"):
		return ErrCertInvalid
	case contains("connection refused"):
		return ErrConnRefused
	case
		contains("no such host"),
		contains("dial tcp: lookup"),
		contains("failure in name resolution"):
		return ErrConn
	default:
		return ErrUnknown
	}
}

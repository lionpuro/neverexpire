package domain

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

func FetchCert(ctx context.Context, domain string) (*model.CertificateInfo, error) {
	errch := make(chan error, 1)
	result := make(chan model.CertificateInfo, 1)
	go func() {
		start := time.Now().UTC()
		conn, err := tls.Dial("tcp", fmt.Sprintf("%s:443", domain), &tls.Config{})
		if err != nil {
			status := errorStatus(err)
			if status != model.CertificateStatusUnknown {
				result <- model.CertificateInfo{
					Status:    status,
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
		status := model.CertificateStatusInvalid
		if cert.NotAfter.After(time.Now().UTC()) {
			status = model.CertificateStatusHealthy
		}
		result <- model.CertificateInfo{
			DNSNames:  strings.Join(cert.DNSNames, ", "),
			IP:        conn.RemoteAddr().String(),
			Expires:   &cert.NotAfter,
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
		return nil, ctx.Err()
	case result := <-result:
		return &result, nil
	}
}

func fingerprint(cert *x509.Certificate) string {
	fingerprint := sha1.Sum(cert.Raw)
	return fmt.Sprintf("%x", fingerprint)
}

func errorStatus(err error) model.CertificateStatus {
	switch {
	case strings.Contains(err.Error(), "tls: failed to verify"):
		return model.CertificateStatusInvalid
	case
		strings.Contains(err.Error(), "connection refused"),
		strings.Contains(err.Error(), "no such host"),
		strings.Contains(err.Error(), "Temporary failure in name resolution"):
		return model.CertificateStatusOffline
	}
	return model.CertificateStatusUnknown
}

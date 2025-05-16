package certs

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/lionpuro/trackcerts/model"
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
					Issuer:    "n/a",
					CheckedAt: start,
					Error:     err,
				}
				return
			}
			errch <- err
			return
		}
		defer conn.Close()

		state := conn.ConnectionState()
		cert := state.PeerCertificates[0]
		result <- model.CertificateInfo{
			DNSNames:  strings.Join(cert.DNSNames, ", "),
			IP:        conn.RemoteAddr().String(),
			Expires:   cert.NotAfter,
			Issuer:    cert.Issuer.Organization[0],
			CheckedAt: start,
			Status:    StatusString(cert.NotAfter),
			Latency:   int(time.Since(start).Milliseconds()),
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

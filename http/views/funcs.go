package views

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/lionpuro/neverexpire/logging"
	"github.com/lionpuro/neverexpire/model"
)

func funcMap() template.FuncMap {
	return template.FuncMap{
		"datef":          datef,
		"sprintf":        fmt.Sprintf,
		"cn":             cn,
		"statusClass":    statusClass,
		"statusText":     statusText,
		"withAttributes": withAttributes,
		"split":          split,
	}
}

func cn(classnames ...string) string {
	var classes []string
	for _, s := range classnames {
		if strings.TrimSpace(s) != "" {
			classes = append(classes, strings.TrimSpace(s))
		}
	}
	return strings.Join(classes, " ")
}

func datef(t time.Time, layout string) string {
	if t.IsZero() {
		return "n/a"
	}
	return t.Format(layout)
}

func statusClass(cert model.CertificateInfo) string {
	switch cert.Status {
	case model.CertificateStatusOffline, model.CertificateStatusUnknown:
		return "text-base-900 bg-[#cacaca]"
	case model.CertificateStatusInvalid:
		return "text-danger-dark bg-danger-light"
	}
	if cert.Expires == nil {
		return ""
	}
	if cert.Expires.Before(time.Now().UTC().AddDate(0, 0, 14)) {
		return "text-warning-dark bg-warning-light"
	}
	return "text-healthy-dark bg-healthy-light"
}

func statusText(cert model.CertificateInfo) string {
	switch cert.Status {
	case model.CertificateStatusUnknown:
		return "-"
	case
		model.CertificateStatusOffline,
		model.CertificateStatusInvalid:
		return cert.Status.String()
	}
	if cert.Expires == nil {
		return "-"
	}
	days := cert.DaysLeft()
	if days == 0 {
		now := time.Now().UTC()
		diff := cert.Expires.Sub(now)
		hours := int(diff.Minutes() / 60)
		return fmt.Sprintf("%d hours", hours)
	}
	return fmt.Sprintf("%d days", days)
}

func split(s, sep string) []string {
	return strings.Split(s, sep)
}

// Use with caution
func withAttributes(kv ...string) map[string]string {
	if len(kv)%2 != 0 {
		logging.DefaultLogger().Error(fmt.Sprintf("missing value for attribute %s", kv[len(kv)-1]))
		return map[string]string{}
	}
	result := make(map[string]string)
	for i := 0; i < len(kv); i += 2 {
		result[kv[i]] = kv[i+1]
	}
	return result
}

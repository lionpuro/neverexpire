package views

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/lionpuro/neverexpire/hosts"
)

func funcMap() template.FuncMap {
	return template.FuncMap{
		"datef":       datef,
		"cn":          cn,
		"ccn":         ccn,
		"statusClass": statusClass,
		"statusText":  statusText,
		"split":       split,
		"kv":          kv,
		"args":        args,
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

func statusClass(cert hosts.CertificateInfo) string {
	switch cert.Status {
	case hosts.CertificateStatusOffline, hosts.CertificateStatusUnknown:
		return "text-base-900 bg-[#cacaca]"
	case hosts.CertificateStatusInvalid:
		return "text-danger-dark bg-danger-light"
	}
	if cert.ExpiresAt == nil {
		return ""
	}
	if cert.ExpiresAt.Before(time.Now().UTC().AddDate(0, 0, 14)) {
		return "text-warning-dark bg-warning-light"
	}
	return "text-healthy-dark bg-healthy-light"
}

func statusText(cert hosts.CertificateInfo) string {
	switch cert.Status {
	case hosts.CertificateStatusUnknown:
		return "-"
	case
		hosts.CertificateStatusOffline,
		hosts.CertificateStatusInvalid:
		return cert.Status.String()
	}
	if cert.ExpiresAt == nil {
		return "-"
	}
	left := cert.TimeLeft()
	days := int(left.Hours() / 24)
	if days == 0 {
		hours := int(left.Minutes() / 60)
		return fmt.Sprintf("%d hours", hours)
	}
	return fmt.Sprintf("%d days", days)
}

func split(s, sep string) []string {
	return strings.Split(s, sep)
}

func ccn(condition bool, cn string) string {
	if condition {
		return cn
	}
	return ""
}

func kv(key string, val any) map[string]any {
	return map[string]any{key: val}
}

func args(kvs ...map[string]any) map[string]any {
	result := make(map[string]any)
	for _, kv := range kvs {
		for k, v := range kv {
			result[k] = v
		}
	}
	return result
}

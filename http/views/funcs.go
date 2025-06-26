package views

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/lionpuro/neverexpire/domain"
	"github.com/lionpuro/neverexpire/logging"
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

func statusClass(status string) string {
	switch status {
	case domain.StatusOffline:
		return "text-base-900 bg-[#cacaca]"
	case domain.StatusInvalid:
		return "text-danger-dark bg-danger-light"
	case domain.StatusExpiring:
		return "text-warning-dark bg-warning-light"
	default:
		return "text-healthy-dark bg-healthy-light"
	}
}

func statusText(status string, expires *time.Time) string {
	if expires == nil {
		if status == domain.StatusInvalid {
			return domain.StatusInvalid
		}
		return domain.StatusOffline
	}
	switch status {
	case domain.StatusOffline:
		return status
	case domain.StatusInvalid:
		return status
	}
	days := domain.DaysLeft(*expires)
	if days == 0 {
		now := time.Now().UTC()
		diff := expires.Sub(now)
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

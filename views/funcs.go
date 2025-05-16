package views

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lionpuro/trackcerts/certs"
)

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
	if t == (time.Time{}) {
		return "n/a"
	}
	return t.Format(layout)
}

func statusClass(status string) string {
	switch status {
	case certs.StatusOffline:
		return "text-base-900 bg-[#cacaca]"
	case certs.StatusInvalid:
		return "text-danger-dark bg-danger-light"
	case certs.StatusExpiring:
		return "text-warning-dark bg-warning-light"
	default:
		return "text-healthy-dark bg-healthy-light"
	}
}

func statusText(status string, expires time.Time) string {
	switch status {
	case certs.StatusOffline:
		return status
	case certs.StatusInvalid:
		return "expired"
	}
	return fmt.Sprintf("%d days", certs.DaysLeft(expires))
}

// Use with caution
func withAttributes(kv ...string) map[string]string {
	if len(kv)%2 != 0 {
		log.Printf("missing value for attribute %s", kv[len(kv)-1])
		return map[string]string{}
	}
	result := make(map[string]string)
	for i := 0; i < len(kv); i += 2 {
		result[kv[i]] = kv[i+1]
	}
	return result
}

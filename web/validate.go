package web

import (
	"fmt"
	"strings"

	"github.com/lionpuro/neverexpire/notifications"
)

func parseWebhook(provider, url string) (*notifications.WebhookProvider, string, error) {
	p, ok := notifications.NewWebhookProvider(provider)
	if !ok {
		return nil, "", fmt.Errorf("invalid webhook provider")
	}
	u := strings.TrimSpace(url)
	if len(u) == 0 {
		return nil, "", fmt.Errorf("invalid webhook url")
	}
	if ok := p.ValidateURL(u); !ok {
		return nil, "", fmt.Errorf("invalid webhook url")
	}
	return &p, u, nil
}

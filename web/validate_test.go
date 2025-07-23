package web

import (
	"testing"

	"github.com/lionpuro/neverexpire/notifications"
)

func TestParseWebhook(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		url      string
		valid    bool
	}{
		{
			name:  "Empty URL",
			url:   "",
			valid: false,
		},
		{
			name:     "Empty URL with Discord provider",
			provider: string(notifications.DiscordProvider),
			url:      "",
			valid:    false,
		},
		{
			name:     "Empty URL with Discord provider",
			provider: string(notifications.DiscordProvider),
			url:      "",
			valid:    false,
		},
		{
			name:     "Valid Discord Webhook",
			provider: string(notifications.DiscordProvider),
			url:      "https://discord.com/api/webhooks/1376156197287362632/rm1wgu-I7mdon75z4eQbHG6KD2ane37YdRSnGIQt59z6xIYAswEDCiDV0gsnfYrddSa1",
			valid:    true,
		},
		{
			name:     "Invalid Discord URL",
			provider: string(notifications.DiscordProvider),
			url:      "https://discord.com/api/webhooks/1234",
			valid:    false,
		},
		{
			name:     "Valid Discord URL with invalid provider",
			provider: "discrod",
			url:      "https://discord.com/api/webhooks/1376156197287362632/rm1wgu-I7mdon75z4eQbHG6KD2ane37YdRSnGIQt59z6xIYAswEDCiDV0gsnfYrddSa1",
			valid:    false,
		},
		{
			name:     "Valid Slack URL",
			provider: string(notifications.SlackProvider),
			url:      "https://hooks.slack.com/services/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",
			valid:    true,
		},
		{
			name:     "Invalid Slack URL",
			provider: string(notifications.SlackProvider),
			url:      "https://hooks.slack.com/services/T7MDON75Z4E",
			valid:    false,
		},
	}
	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			_, _, err := parseWebhook(ts.provider, ts.url)
			if !ts.valid && err == nil {
				t.Error("expected error and got none")
			} else if ts.valid && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

package handlers

import "testing"

func TestParseWebhookURL(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		valid bool
	}{
		{
			name:  "Empty URL",
			url:   "",
			valid: false,
		},
		{
			name:  "Valid Discord URL",
			url:   "https://discord.com/api/webhooks/1376156197287362632/rm1wgu-I7mdon75z4eQbHG6KD2ane37YdRSnGIQt59z6xIYAswEDCiDV0gsnfYrddSa1",
			valid: true,
		},
		{
			name:  "Invalid Discord URL",
			url:   "https://discord.com/api/webhooks/1234",
			valid: false,
		},
		{
			name:  "Valid Slack URL",
			url:   "https://hooks.slack.com/services/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",
			valid: true,
		},
		{
			name:  "Invalid Slack URL",
			url:   "https://hooks.slack.com/services/T7MDON75Z4E",
			valid: false,
		},
	}
	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			_, err := parseWebhookURL(ts.url)
			if !ts.valid && err == nil {
				t.Error("expected error and got none")
			} else if ts.valid && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

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

func TestParseHostname(t *testing.T) {
	var longStr string
	for range 10 {
		longStr = longStr + "abcdefghijklmnopqrstuvwxyz"
	}
	tests := []struct {
		name           string
		hostname       string
		expectedResult string
		expectErr      bool
	}{
		{
			name:           "Empty string",
			hostname:       "",
			expectedResult: "",
			expectErr:      true,
		},
		{
			name:           "No TLD",
			hostname:       "localhost",
			expectedResult: "localhost",
			expectErr:      false,
		},
		{
			name:           "No TLD with port",
			hostname:       "localhost:80",
			expectedResult: "localhost",
			expectErr:      false,
		},
		{
			name:           "Valid domain",
			hostname:       "example.com",
			expectedResult: "example.com",
			expectErr:      false,
		},
		{
			name:           "Valid subdomain",
			hostname:       "www.example.com",
			expectedResult: "www.example.com",
			expectErr:      false,
		},
		{
			name:           "Valid domain with protocol",
			hostname:       "https://example.com",
			expectedResult: "example.com",
			expectErr:      false,
		},
		{
			name:           "Valid domain with leading and trailing whitespaces",
			hostname:       " example.com ",
			expectedResult: "example.com",
			expectErr:      false,
		},
		{
			name:           "Invalid domain (starts with '-')",
			hostname:       "-example.com",
			expectedResult: "",
			expectErr:      true,
		},
		{
			name:           "Invalid protocol",
			hostname:       "htp://example.com",
			expectedResult: "",
			expectErr:      true,
		},
		{
			name:           "No host",
			hostname:       "https://",
			expectedResult: "",
			expectErr:      true,
		},
		{
			name:           "Too long",
			hostname:       longStr,
			expectedResult: "",
			expectErr:      true,
		},
	}

	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			result, err := parseHostname(ts.hostname)
			if ts.expectErr && err == nil {
				t.Error("expected error and got none")
			} else if !ts.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			} else if result != ts.expectedResult {
				t.Errorf("incorrect result: expected %s, got %s", ts.expectedResult, result)
			}
		})
	}
}

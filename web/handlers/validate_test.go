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

func TestParseDomain(t *testing.T) {
	var longStr string
	for range 10 {
		longStr = longStr + "abcdefghijklmnopqrstuvwxyz"
	}
	tests := []struct {
		name           string
		domain         string
		expectedResult string
		expectErr      bool
	}{
		{
			name:           "Empty string",
			domain:         "",
			expectedResult: "",
			expectErr:      true,
		},
		{
			name:           "No TLD",
			domain:         "localhost",
			expectedResult: "localhost",
			expectErr:      false,
		},
		{
			name:           "No TLD with port",
			domain:         "localhost:80",
			expectedResult: "localhost",
			expectErr:      false,
		},
		{
			name:           "Valid domain",
			domain:         "example.com",
			expectedResult: "example.com",
			expectErr:      false,
		},
		{
			name:           "Valid subdomain",
			domain:         "www.example.com",
			expectedResult: "www.example.com",
			expectErr:      false,
		},
		{
			name:           "Valid domain with protocol",
			domain:         "https://example.com",
			expectedResult: "example.com",
			expectErr:      false,
		},
		{
			name:           "Valid domain with leading and trailing whitespaces",
			domain:         " example.com ",
			expectedResult: "example.com",
			expectErr:      false,
		},
		{
			name:           "Invalid domain (starts with '-')",
			domain:         "-example.com",
			expectedResult: "",
			expectErr:      true,
		},
		{
			name:           "Invalid protocol",
			domain:         "htp://example.com",
			expectedResult: "",
			expectErr:      true,
		},
		{
			name:           "No host",
			domain:         "https://",
			expectedResult: "",
			expectErr:      true,
		},
		{
			name:           "Too long",
			domain:         longStr,
			expectedResult: "",
			expectErr:      true,
		},
	}

	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			result, err := parseDomain(ts.domain)
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

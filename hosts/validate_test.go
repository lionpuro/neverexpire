package hosts_test

import (
	"testing"

	"github.com/lionpuro/neverexpire/hosts"
)

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
			result, err := hosts.ParseHostname(ts.hostname)
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

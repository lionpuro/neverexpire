package http

import "testing"

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

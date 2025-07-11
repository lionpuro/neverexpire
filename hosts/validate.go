package hosts

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
)

func ParseHostname(input string) (string, error) {
	if len(input) > 200 {
		return "", fmt.Errorf("hostname too long")
	}
	s := strings.TrimSpace(input)
	if s == "" {
		return "", fmt.Errorf("hostname can't be empty")
	}
	split := strings.Split(s, "://")
	if len(split) > 1 {
		if !slices.Contains([]string{"https", "http"}, split[0]) {
			return "", fmt.Errorf("invalid protocol")
		}
		input = "https://" + split[1]
	}
	if len(split) == 1 {
		input = "https://" + split[0]
	}

	u, err := url.Parse(input)
	if err != nil {
		return "", fmt.Errorf("invalid url: %v", err)
	}
	dn := u.Hostname()
	if dn == "" {
		return "", fmt.Errorf("invalid hostname")
	}
	for _, s := range strings.Split(dn, ".") {
		if len(s) == 0 {
			return "", fmt.Errorf("invalid hostname")
		}
		if !isAlphanumeric(rune(s[0])) || !isAlphanumeric(rune(s[len(s)-1])) {
			return "", fmt.Errorf("illegal character in hostname")
		}
	}

	return strings.TrimPrefix(dn, "https://"), nil
}

func isAlphanumeric(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

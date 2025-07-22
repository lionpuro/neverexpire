package web

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	discordPattern = regexp.MustCompile(`^https:\/\/discord\.com\/api\/webhooks\/[0-9]{18,19}\/[a-zA-Z0-9_-]+$`)
	slackPattern   = regexp.MustCompile(`^https:\/\/hooks\.slack\.com\/services\/[a-zA-Z0-9]+\/[a-zA-Z0-9]+\/[a-zA-Z0-9]+$`)
)

func parseWebhookURL(input string) (string, error) {
	str := strings.TrimSpace(input)
	if len(str) == 0 {
		return "", fmt.Errorf("webhook url can't be empty")
	}
	switch {
	case strings.HasPrefix(str, "https://discord.com"):
		if !discordPattern.MatchString(str) {
			return "", fmt.Errorf("invalid webhook url")
		}
		return str, nil
	case strings.HasPrefix(str, "https://hooks.slack.com"):
		if !slackPattern.MatchString(str) {
			return "", fmt.Errorf("invalid webhook url")
		}
		return str, nil
	}
	return "", fmt.Errorf("invalid webhook url")
}

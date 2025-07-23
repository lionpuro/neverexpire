package notifications

import (
	"regexp"
)

var webhookRegexp = map[WebhookProvider]*regexp.Regexp{
	DiscordProvider: regexp.MustCompile(
		`^https:\/\/discord\.com\/api\/webhooks\/[0-9]{18,19}\/[a-zA-Z0-9_-]+$`,
	),
	SlackProvider: regexp.MustCompile(
		`^https:\/\/hooks\.slack\.com\/services\/[a-zA-Z0-9]+\/[a-zA-Z0-9]+\/[a-zA-Z0-9]+$`,
	),
}

const (
	DiscordProvider WebhookProvider = "DISCORD"
	SlackProvider   WebhookProvider = "SLACK"
)

type WebhookProvider string

func (p WebhookProvider) String() string {
	return string(p)
}

func (p WebhookProvider) ValidateURL(url string) bool {
	pattern, ok := webhookRegexp[p]
	if !ok {
		return false
	}
	return pattern.MatchString(url)
}

func NewWebhookProvider(input string) (WebhookProvider, bool) {
	switch input {
	case string(DiscordProvider):
		return DiscordProvider, true
	case string(SlackProvider):
		return SlackProvider, true
	default:
		return WebhookProvider(""), false
	}
}

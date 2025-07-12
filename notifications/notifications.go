package notifications

import (
	"net/http"
	"time"

	"github.com/lionpuro/neverexpire/logging"
)

const (
	ThresholdDay    = 24 * 60 * 60
	Threshold2Days  = ThresholdDay * 2
	ThresholdWeek   = ThresholdDay * 7
	Threshold2Weeks = ThresholdDay * 7 * 2
)

const (
	testMessage = "Hello! Your notification webhook for neverexpire.xyz is set up correctly."
)

func SendTestNotification(url string) error {
	c := &http.Client{
		Timeout: 10 * time.Second,
	}
	return sendNotification(logging.DefaultLogger(), c, url, testMessage)
}

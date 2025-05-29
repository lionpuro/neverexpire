package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	TestMessage = "Hello!\nIf you're seeing this, the notification webhook has been set up correctly."
)

type Notifier struct {
	client *http.Client
}

func NewNotifier() *Notifier {
	return &Notifier{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (n *Notifier) Send(url, msg string) error {
	buf, err := json.Marshal(map[string]string{
		"content": msg,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	code := res.StatusCode
	if code < 200 || code > 299 {
		return fmt.Errorf("response status: %s", res.Status)
	}
	return nil
}

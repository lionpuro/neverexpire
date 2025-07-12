package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lionpuro/neverexpire/hosts"
	"github.com/lionpuro/neverexpire/logging"
)

var avatarURL string = os.Getenv("WEBHOOK_AVATAR_URL")

type Worker struct {
	interval      time.Duration
	client        *http.Client
	notifications *Service
	hosts         *hosts.Service
	log           logging.Logger
}

func NewWorker(
	interval time.Duration,
	ns *Service,
	hs *hosts.Service,
	logger logging.Logger,
) *Worker {
	return &Worker{
		interval: interval,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		notifications: ns,
		hosts:         hs,
		log:           logger,
	}
}

func (w *Worker) Start(ctx context.Context) {
	t := time.NewTicker(w.interval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if err := w.processNotifications(ctx); err != nil {
				w.log.Error("failed to process notifications", "error", err.Error())
			}
		case <-ctx.Done():
			return
		}
	}
}

func (w *Worker) send(notif Notification) error {
	return sendNotification(w.log, w.client, notif.Endpoint, notif.Body)
}

func (w *Worker) notify(notif Notification) error {
	if err := w.send(notif); err != nil {
		attempts := notif.Attempts + 1
		input := NotificationUpdate{
			Attempts: &attempts,
		}
		if err := w.notifications.Update(context.Background(), notif.ID, input); err != nil {
			return err
		}
		return fmt.Errorf("failed to send notification: %v", err)
	}
	ts := time.Now().UTC()
	input := NotificationUpdate{
		DeliveredAt: &ts,
	}
	err := w.notifications.Update(context.Background(), notif.ID, input)
	return err
}

func (w *Worker) processNotifications(ctx context.Context) error {
	hosts, err := w.hosts.Expiring(ctx)
	if err != nil {
		return err
	}
	if err := w.notifications.CreateReminders(ctx, hosts); err != nil {
		return err
	}
	notifs, err := w.notifications.AllDue(ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, notif := range notifs {
		wg.Add(1)
		go func(no Notification) {
			defer wg.Done()
			if err := w.notify(notif); err != nil {
				w.log.Error("failed to notify user", "error", err.Error())
			}
		}(notif)
	}
	go func() {
		wg.Wait()
	}()

	return nil
}

func sendNotification(logger logging.Logger, client *http.Client, url, msg string) error {
	body := map[string]string{}
	if strings.Contains(url, "discord") {
		body["content"] = msg
		if avatarURL != "" {
			body["avatar_url"] = avatarURL
		}
	}
	if strings.Contains(url, "slack") {
		body["text"] = msg
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			logger.Error("error closing webhook response body", "error", err.Error())
		}
	}()
	code := res.StatusCode
	if code < 200 || code > 299 {
		return fmt.Errorf("response status: %s", res.Status)
	}
	return nil
}

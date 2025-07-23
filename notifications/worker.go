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
			if err := w.NotifyExpiring(ctx); err != nil {
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
	notif.Attempts++
	if err := w.send(notif); err != nil {
		if err := w.notifications.Upsert(context.Background(), notif); err != nil {
			return err
		}
		return fmt.Errorf("failed to send notification: %v", err)
	}
	t := time.Now().UTC()
	notif.DeliveredAt = &t
	return w.notifications.Upsert(context.Background(), notif)
}

func (w *Worker) NotifyExpiring(ctx context.Context) error {
	records, err := w.hosts.Expiring(ctx)
	if err != nil {
		return err
	}
	var notifs []Notification
	for _, rec := range records {
		n := newReminder(rec)
		if n != nil {
			notifs = append(notifs, *n)
		}
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

func newReminder(record hosts.NotifiableHost) *Notification {
	exp := record.Host.Certificate.ExpiresAt
	if exp == nil {
		return nil
	}
	msg := formatReminderMsg(record.Host)
	diff := time.Duration(record.Threshold) * time.Second
	n := &Notification{
		Endpoint:     record.WebhookURL,
		UserID:       record.UserID,
		HostID:       record.Host.ID,
		Type:         NotificationTypeExpiration,
		Body:         msg,
		Due:          record.Host.Certificate.ExpiresAt.Add(-diff),
		DeliveredAt:  nil,
		Attempts:     record.Attempts,
		DeletedAfter: *exp,
	}
	return n
}

func formatReminderMsg(d hosts.Host) string {
	hours := int(d.Certificate.TimeLeft().Hours())
	count := hours / 24
	unit := "days"
	switch {
	case hours < 24:
		count = hours
		unit = "hours"
		if count == 1 {
			unit = "hour"
		}
	default:
		if count == 1 {
			unit = "day"
		}
	}
	msg := fmt.Sprintf(
		"TLS certificate for %s is expiring in %d %s (at %s UTC)",
		d.Hostname,
		count,
		unit,
		d.Certificate.ExpiresAt.Format(time.DateTime),
	)
	return msg
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

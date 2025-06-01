package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/lionpuro/trackcerts/domain"
	"github.com/lionpuro/trackcerts/model"
)

const (
	testMessage = "Hello!\nIf you're seeing this, the notification webhook has been set up correctly."
)

type Notifier struct {
	client        *http.Client
	notifications *Service
	domains       *domain.Service
}

func NewNotifier(ns *Service, ds *domain.Service) *Notifier {
	return &Notifier{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		notifications: ns,
		domains:       ds,
	}
}

func (n *Notifier) Start(ctx context.Context) {
	t := time.NewTicker(60 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if err := n.processNotifications(ctx); err != nil {
				log.Printf("process notifications: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (n *Notifier) send(notif model.Notification) error {
	return sendNotification(n.client, notif.Endpoint, notif.Body)
}

func sendNotification(client *http.Client, url, msg string) error {
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
	res, err := client.Do(req)
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

func (n *Notifier) notify(notif model.Notification) error {
	if err := n.send(notif); err != nil {
		attempts := notif.Attempts + 1
		input := model.NotificationUpdate{
			Attempts: &attempts,
		}
		if err := n.notifications.Update(context.Background(), notif.ID, input); err != nil {
			return err
		}
		return err
	}
	ts := time.Now().UTC()
	input := model.NotificationUpdate{
		DeliveredAt: &ts,
	}
	err := n.notifications.Update(context.Background(), notif.ID, input)
	return err
}

func (n *Notifier) processNotifications(ctx context.Context) error {
	domains, err := n.domains.Notifiable(ctx)
	if err != nil {
		return err
	}
	if err := n.notifications.CreateReminders(ctx, domains); err != nil {
		return err
	}
	notifs, err := n.notifications.AllDue(ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, notif := range notifs {
		wg.Add(1)
		go func(no model.Notification) {
			defer wg.Done()
			if err := n.notify(notif); err != nil {
				log.Printf("notify: %v", err)
			}
		}(notif)
	}
	go func() {
		wg.Wait()
	}()

	return nil
}

func SendTestNotification(url string) error {
	c := &http.Client{
		Timeout: 10 * time.Second,
	}
	return sendNotification(c, url, testMessage)
}

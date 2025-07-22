package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/lionpuro/neverexpire/hosts"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, n NotificationInput) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.repo.Create(ctx, n)
}

func (s *Service) Update(ctx context.Context, id int, n NotificationUpdate) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.repo.Update(ctx, id, n)
}

func (s *Service) AllDue(ctx context.Context) ([]Notification, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.repo.AllDue(ctx)
}

func (s *Service) CreateReminders(ctx context.Context, hosts []hosts.HostWithUser) error {
	if len(hosts) == 0 {
		return nil
	}
	var eg errgroup.Group
	for _, h := range hosts {
		eg.Go(func() error {
			now := time.Now().UTC()
			if h.Host.Certificate.ExpiresAt == nil {
				return nil
			}
			if !h.Host.Certificate.ExpiresAt.Before(now) {
				if err := s.createReminder(ctx, h); err != nil {
					return err
				}
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func (s *Service) createReminder(ctx context.Context, record hosts.HostWithUser) error {
	exp := record.Host.Certificate.ExpiresAt
	if exp == nil {
		return nil
	}
	msg := formatReminder(record.Host)
	diff := time.Duration(record.Settings.ReminderThreshold) * time.Second
	input := NotificationInput{
		UserID:       record.User.ID,
		HostID:       record.Host.ID,
		Type:         NotificationTypeExpiration,
		Body:         msg,
		Due:          record.Host.Certificate.ExpiresAt.Add(-diff),
		DeliveredAt:  nil,
		Attempts:     0,
		DeletedAfter: *exp,
	}
	return s.repo.Create(ctx, input)
}

func formatReminder(d hosts.Host) string {
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

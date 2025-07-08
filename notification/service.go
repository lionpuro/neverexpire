package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/lionpuro/neverexpire/model"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, n model.NotificationInput) error {
	return s.repo.Create(ctx, n)
}

func (s *Service) Update(ctx context.Context, id int, n model.NotificationUpdate) error {
	return s.repo.Update(ctx, id, n)
}

func (s *Service) AllDue(ctx context.Context) ([]model.Notification, error) {
	return s.repo.AllDue(ctx)
}

func (s *Service) CreateReminders(ctx context.Context, domains []model.DomainWithUser) error {
	for _, d := range domains {
		now := time.Now().UTC()
		if !d.Domain.Certificate.ExpiresAt.Before(now) {
			if err := s.createReminder(ctx, d); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) createReminder(ctx context.Context, record model.DomainWithUser) error {
	exp := record.Domain.Certificate.ExpiresAt
	if exp == nil {
		return nil
	}
	msg := formatReminder(record.Domain)
	diff := time.Duration(record.Settings.RemindBefore) * time.Second
	input := model.NotificationInput{
		UserID:       record.User.ID,
		DomainID:     record.Domain.ID,
		Type:         model.NotificationTypeExpiration,
		Body:         msg,
		Due:          record.Domain.Certificate.ExpiresAt.Add(-diff),
		DeliveredAt:  nil,
		Attempts:     0,
		DeletedAfter: *exp,
	}
	return s.repo.Create(ctx, input)
}

func formatReminder(d model.Domain) string {
	hours := int(d.Certificate.TimeLeft().Hours())
	count := hours / 24
	unit := "days"
	if hours < 24 {
		count = hours
		unit = "hours"
	}
	msg := fmt.Sprintf(
		"TLS certificate for %s is expiring in %d %s (at %s UTC)",
		d.DomainName,
		count,
		unit,
		d.Certificate.ExpiresAt.Format(time.DateTime),
	)
	return msg
}

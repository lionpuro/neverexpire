package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/lionpuro/neverexpire/domain"
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
		if !d.Domain.Certificate.Expires.Before(now) {
			if err := s.createReminder(ctx, d); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) createReminder(ctx context.Context, record model.DomainWithUser) error {
	exp := record.Domain.Certificate.Expires
	if exp == nil {
		return nil
	}
	days := domain.DaysLeft(*exp)
	body := fmt.Sprintf("SSL certificate for %s is expiring in %d days!", record.Domain.DomainName, days)
	diff := time.Duration(record.Settings.RemindBefore) * time.Second
	input := model.NotificationInput{
		UserID:       record.User.ID,
		DomainID:     record.Domain.ID,
		Type:         model.NotificationTypeExpiration,
		Body:         body,
		Due:          record.Domain.Certificate.Expires.Add(-diff),
		DeliveredAt:  nil,
		Attempts:     0,
		DeletedAfter: *exp,
	}
	return s.repo.Create(ctx, input)
}

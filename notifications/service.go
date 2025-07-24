package notifications

import (
	"context"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) AllByUser(ctx context.Context, uid string) ([]AppNotification, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.repo.AllByUser(ctx, uid)
}

func (s *Service) Update(uid string, input []NotificationUpdate) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.repo.Update(ctx, uid, input)
}

func (s *Service) Upsert(ctx context.Context, n Notification) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.repo.Upsert(ctx, n)
}

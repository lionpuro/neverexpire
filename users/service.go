package users

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

func (s *Service) ByID(ctx context.Context, id string) (User, error) {
	return s.repo.ByID(ctx, id)
}

func (s *Service) Create(id, email string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return s.repo.Create(ctx, id, email)
}

func (s *Service) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return s.repo.Delete(ctx, id)
}

func (s *Service) Settings(ctx context.Context, userID string) (Settings, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	sett, err := s.repo.Settings(ctx, userID)
	if err != nil {
		return Settings{}, err
	}
	return sett, nil
}

func (s *Service) SaveSettings(userID string, settings SettingsInput) (Settings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return s.repo.SaveSettings(ctx, userID, settings)
}

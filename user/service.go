package user

import (
	"context"
	"time"

	"github.com/lionpuro/trackcerts/model"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ByID(ctx context.Context, id string) (model.User, error) {
	return s.repo.ByID(ctx, id)
}

func (s *Service) Create(id, email string) error {
	return s.repo.Create(id, email)
}

func (s *Service) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *Service) Settings(ctx context.Context, userID string) (model.Settings, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	sett, err := s.repo.Settings(ctx, userID)
	if err != nil {
		return model.Settings{}, err
	}
	return sett, nil
}

func (s *Service) SaveSettings(userID string, settings model.SettingsInput) (model.Settings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return s.repo.SaveSettings(ctx, userID, settings)
}

package domain

import (
	"context"
	"time"

	"github.com/lionpuro/trackcerts/certs"
	"github.com/lionpuro/trackcerts/model"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ByID(ctx context.Context, id int, userID string) (model.Domain, error) {
	return s.repo.ByID(ctx, userID, id)
}

func (s *Service) All(ctx context.Context, userID string) ([]model.Domain, error) {
	return s.repo.AllByUser(ctx, userID)
}

func (s *Service) Create(user model.User, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	info, err := certs.FetchCert(ctx, name)
	if err != nil {
		return err
	}
	domain := model.Domain{
		UserID:      user.ID,
		DomainName:  name,
		Certificate: *info,
	}
	return s.repo.Create(domain)
}

func (s *Service) Update(d model.Domain) (model.Domain, error) {
	return s.repo.Update(d)
}

func (s *Service) Delete(userID string, id int) error {
	return s.repo.Delete(userID, id)
}

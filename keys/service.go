package keys

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

func (s *Service) ByUser(ctx context.Context, uid string) ([]AccessKey, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	keys, err := s.repo.ByUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (s *Service) ByID(ctx context.Context, id string) (AccessKey, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	key, err := s.repo.ByID(ctx, id)
	if err != nil {
		return AccessKey{}, err
	}
	return key, nil
}

func (s *Service) Create(uid string) (string, *AccessKey, error) {
	raw, err := GenerateAccessKey()
	if err != nil {
		return "", nil, err
	}
	key, err := NewAccessKey(raw, uid)
	if err != nil {
		return "", nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := s.repo.Create(ctx, *key); err != nil {
		return "", nil, err
	}
	return raw, key, nil
}

func (s *Service) Delete(id, uid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return s.repo.Delete(ctx, id, uid)
}

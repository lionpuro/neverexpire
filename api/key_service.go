package api

import (
	"context"
	"time"
)

type KeyService struct {
	repo *KeyRepository
}

func NewKeyService(repo *KeyRepository) *KeyService {
	return &KeyService{repo: repo}
}

func (s *KeyService) ByUser(ctx context.Context, uid string) ([]Key, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	keys, err := s.repo.ByUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (s *KeyService) ByID(ctx context.Context, id string) (Key, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	key, err := s.repo.ByID(ctx, id)
	if err != nil {
		return Key{}, err
	}
	return key, nil
}

func (s *KeyService) Create(uid string) (string, *Key, error) {
	raw, err := GenerateKey()
	if err != nil {
		return "", nil, err
	}
	key, err := NewKey(raw, uid)
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

func (s *KeyService) Delete(id, uid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return s.repo.Delete(ctx, id, uid)
}

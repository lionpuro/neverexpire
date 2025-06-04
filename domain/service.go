package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lionpuro/trackcerts/certs"
	"github.com/lionpuro/trackcerts/model"
	"golang.org/x/sync/errgroup"
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

func (s *Service) Notifiable(ctx context.Context) ([]model.DomainWithSettings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return s.repo.Notifiable(ctx)
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

func (s *Service) CreateMultiple(user model.User, names []string) error {
	eg := errgroup.Group{}
	for _, name := range names {
		func(n string) {
			eg.Go(func() error {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				info, err := certs.FetchCert(ctx, name)
				if err != nil {
					if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "Temporary failure in name resolution") {
						return fmt.Errorf("can't connect to %s", n)
					}
					return fmt.Errorf("fetch cert: %v", err)
				}
				domain := model.Domain{
					UserID:      user.ID,
					DomainName:  n,
					Certificate: *info,
				}
				if err := s.repo.Create(domain); err != nil {
					str := `duplicate key value violates unique constraint "uq_domains_user_id_domain_name"`
					if strings.Contains(err.Error(), str) {
						return fmt.Errorf("already tracking %s", n)
					}
					return err
				}
				return nil
			})
		}(name)
	}
	return eg.Wait()
}

func (s *Service) Update(d model.Domain) (model.Domain, error) {
	return s.repo.Update(d)
}

func (s *Service) Delete(userID string, id int) error {
	return s.repo.Delete(userID, id)
}

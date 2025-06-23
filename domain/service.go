package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lionpuro/neverexpire/model"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ByID(ctx context.Context, id int, userID string) (model.Domain, error) {
	return s.repo.ByID(ctx, userID, id)
}

func (s *Service) AllByUser(ctx context.Context, userID string) ([]model.Domain, error) {
	return s.repo.AllByUser(ctx, userID)
}

func (s *Service) All(ctx context.Context) ([]model.Domain, error) {
	return s.repo.All(ctx)
}

func (s *Service) Notifiable(ctx context.Context) ([]model.DomainWithSettings, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	return s.repo.Notifiable(ctx)
}

func (s *Service) Create(user model.User, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	info, err := FetchCert(ctx, name)
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
	domainch := make(chan model.Domain, len(names))
	domains := make([]model.Domain, 0)
	eg, ctx := errgroup.WithContext(context.Background())
	for _, name := range names {
		eg.Go(func() error {
			fetchCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			info, err := FetchCert(fetchCtx, name)
			if err != nil {
				if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "Temporary failure in name resolution") {
					return fmt.Errorf("can't connect to %s", name)
				}
				return fmt.Errorf("fetch cert: %v", err)
			}
			domain := model.Domain{
				UserID:      user.ID,
				DomainName:  name,
				Certificate: *info,
			}
			select {
			case domainch <- domain:
			case <-ctx.Done():
				return context.Canceled
			default:
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	close(domainch)
	for d := range domainch {
		domains = append(domains, d)
	}
	return s.repo.CreateMultiple(domains)
}

func (s *Service) Update(d model.Domain) (model.Domain, error) {
	return s.repo.Update(d)
}

func (s *Service) UpdateMultiple(ctx context.Context, domains []model.Domain) error {
	return s.repo.UpdateMultiple(ctx, domains)
}

func (s *Service) Delete(userID string, id int) error {
	return s.repo.Delete(userID, id)
}

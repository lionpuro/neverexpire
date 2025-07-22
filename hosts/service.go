package hosts

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ByID(ctx context.Context, id int, userID string) (Host, error) {
	return s.repo.ByID(ctx, userID, id)
}

func (s *Service) ByName(ctx context.Context, name, userID string) (Host, error) {
	return s.repo.ByName(ctx, userID, name)
}

func (s *Service) AllByUser(ctx context.Context, userID string) ([]Host, error) {
	return s.repo.AllByUser(ctx, userID)
}

func (s *Service) All(ctx context.Context) ([]Host, error) {
	return s.repo.All(ctx)
}

func (s *Service) Expiring(ctx context.Context) ([]NotifiableHost, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	return s.repo.Expiring(ctx)
}

func (s *Service) Create(uid string, names []string) error {
	hostch := make(chan Host, len(names))
	hosts := make([]Host, 0)
	eg, ctx := errgroup.WithContext(context.Background())
	for _, name := range names {
		eg.Go(func() error {
			info, err := FetchCert(context.Background(), name)
			if err != nil {
				if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "Temporary failure in name resolution") {
					return fmt.Errorf("can't connect to %s", name)
				}
				if !errors.Is(err, context.DeadlineExceeded) {
					return fmt.Errorf("fetch cert: %v", err)
				}
				info = &CertificateInfo{
					Status:    CertificateStatusOffline,
					IssuedBy:  "n/a",
					CheckedAt: time.Now().UTC(),
					Error:     err,
				}
			}
			host := Host{
				Hostname:    name,
				Certificate: *info,
			}
			select {
			case hostch <- host:
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
	close(hostch)
	for h := range hostch {
		hosts = append(hosts, h)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return s.repo.Create(ctx, uid, hosts)
}

func (s *Service) Update(ctx context.Context, hosts []Host) error {
	return s.repo.Update(ctx, hosts)
}

func (s *Service) Delete(userID string, id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return s.repo.Delete(ctx, userID, id)
}

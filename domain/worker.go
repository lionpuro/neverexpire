package domain

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/lionpuro/neverexpire/logging"
)

type Worker struct {
	interval time.Duration
	domains  *Service
	quit     chan struct{}
	log      logging.Logger
}

func NewWorker(interval time.Duration, ds *Service, logger logging.Logger) *Worker {
	return &Worker{
		interval: interval,
		domains:  ds,
		quit:     make(chan struct{}),
		log:      logger,
	}
}

func (w *Worker) Start() {
	t := time.NewTicker(w.interval)
	if err := w.poll(); err != nil {
		log.Fatal(err)
		return
	}
	for {
		select {
		case t := <-t.C:
			start := time.Now()
			w.log.Info(fmt.Sprintf("start polling at %v", t))
			if err := w.poll(); err != nil {
				w.log.Error("error polling domains", "error", err.Error())
			}
			w.log.Info(fmt.Sprintf("finish polling in %v", time.Since(start)))
		case <-w.quit:
			w.log.Info("stopping monitor...")
			return
		}
	}
}

func (w *Worker) poll() error {
	domains, err := w.domains.All(context.Background())
	if err != nil {
		return err
	}
	workers := make(chan struct{}, 15)
	wg := sync.WaitGroup{}
	results := make(chan Domain, len(domains))

	for _, dom := range domains {
		wg.Add(1)
		go func(d Domain) {
			workers <- struct{}{}
			defer func() {
				<-workers
				wg.Done()
			}()
			cert, err := FetchCert(context.Background(), d.DomainName)
			if err != nil {
				cert = &CertificateInfo{
					Status:    CertificateStatusOffline,
					IssuedBy:  "n/a",
					CheckedAt: time.Now().UTC(),
					Error:     err,
				}
			}
			domain := d
			domain.Certificate = *cert
			results <- domain
		}(dom)
	}

	wg.Wait()
	close(results)
	return w.updateData(results)
}

func (w *Worker) updateData(domainch chan Domain) error {
	domains := make([]Domain, len(domainch))
	i := 0
	for d := range domainch {
		domains[i] = d
		i++
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return w.domains.Update(ctx, domains)
}

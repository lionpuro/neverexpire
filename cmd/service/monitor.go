package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/lionpuro/neverexpire/domain"
	"github.com/lionpuro/neverexpire/logging"
	"github.com/lionpuro/neverexpire/model"
	"github.com/lionpuro/neverexpire/notification"
)

type Monitor struct {
	interval      time.Duration
	domains       *domain.Service
	notifications *notification.Service
	quit          chan struct{}
	log           logging.Logger
}

func NewMonitor(interval time.Duration, ds *domain.Service, ns *notification.Service, logger logging.Logger) *Monitor {
	return &Monitor{
		interval:      interval,
		domains:       ds,
		notifications: ns,
		quit:          make(chan struct{}),
		log:           logger,
	}
}

func (m *Monitor) Start() {
	t := time.NewTicker(m.interval)
	if err := m.poll(); err != nil {
		log.Fatal(err)
		return
	}
	for {
		select {
		case t := <-t.C:
			start := time.Now()
			m.log.Info(fmt.Sprintf("start polling at %v", t))
			if err := m.poll(); err != nil {
				m.log.Error("error polling domains", "error", err.Error())
			}
			m.log.Info(fmt.Sprintf("finish polling in %v", time.Since(start)))
		case <-m.quit:
			m.log.Info("stopping monitor...")
			return
		}
	}
}

func (m *Monitor) poll() error {
	domains, err := m.domains.All(context.Background())
	if err != nil {
		return err
	}
	workers := make(chan struct{}, 15)
	wg := sync.WaitGroup{}
	results := make(chan model.Domain, len(domains))

	for _, dom := range domains {
		wg.Add(1)
		go func(d model.Domain) {
			workers <- struct{}{}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer func() {
				<-workers
				wg.Done()
				cancel()
			}()
			cert, err := domain.FetchCert(ctx, d.DomainName)
			if err != nil {
				cert = &model.CertificateInfo{
					Status:    domain.StatusOffline,
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
	return m.updateData(results)
}

func (m *Monitor) updateData(domainch chan model.Domain) error {
	domains := make([]model.Domain, len(domainch))
	i := 0
	for d := range domainch {
		domains[i] = d
		i++
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return m.domains.Update(ctx, domains)
}

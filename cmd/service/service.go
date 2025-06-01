package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/lionpuro/trackcerts/certs"
	"github.com/lionpuro/trackcerts/db"
	"github.com/lionpuro/trackcerts/domain"
	"github.com/lionpuro/trackcerts/model"
	"github.com/lionpuro/trackcerts/notification"
)

func main() {
	conn := fmt.Sprintf(
		"postgres://%s:%s@localhost:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST_PORT"),
		os.Getenv("POSTGRES_DB"),
	)

	pool, err := db.NewPool(conn)
	if err != nil {
		log.Fatal(err)
		return
	}

	dr := domain.NewRepository(pool)
	ds := domain.NewService(dr)
	ns := notification.NewService(notification.NewRepository(pool))
	monitor := NewMonitor(time.Minute*30, dr, ns)
	notifier := notification.NewNotifier(ns, ds)

	fmt.Println("Starting notification service...")
	go notifier.Start(context.Background())

	fmt.Println("Starting domain monitoring service...")
	monitor.Start()
}

type Monitor struct {
	interval      time.Duration
	domains       domain.Repository
	notifications *notification.Service
	quit          chan struct{}
}

func NewMonitor(interval time.Duration, dr domain.Repository, ns *notification.Service) *Monitor {
	return &Monitor{
		interval:      interval,
		domains:       dr,
		notifications: ns,
		quit:          make(chan struct{}),
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
			log.Printf("start polling at %v", t)
			if err := m.poll(); err != nil {
				log.Printf("poll service: %v", err)
			}
			log.Printf("finish polling in %v", time.Since(start))
		case <-m.quit:
			log.Printf("stopping monitor...")
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
			cert, err := certs.FetchCert(ctx, d.DomainName)
			if err != nil {
				log.Printf("fetch certificate: %v", err)
				return
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
	return m.domains.UpdateMultiple(ctx, domains)
}

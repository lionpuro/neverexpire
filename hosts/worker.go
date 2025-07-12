package hosts

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
	hosts    *Service
	quit     chan struct{}
	log      logging.Logger
}

func NewWorker(interval time.Duration, hs *Service, logger logging.Logger) *Worker {
	return &Worker{
		interval: interval,
		hosts:    hs,
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
				w.log.Error("error polling hosts", "error", err.Error())
			}
			w.log.Info(fmt.Sprintf("finish polling in %v", time.Since(start)))
		case <-w.quit:
			w.log.Info("stopping monitor...")
			return
		}
	}
}

func (w *Worker) poll() error {
	hosts, err := w.hosts.All(context.Background())
	if err != nil {
		return err
	}
	workers := make(chan struct{}, 15)
	wg := sync.WaitGroup{}
	results := make(chan Host, len(hosts))

	for _, hst := range hosts {
		wg.Add(1)
		go func(h Host) {
			workers <- struct{}{}
			defer func() {
				<-workers
				wg.Done()
			}()
			cert, err := FetchCert(context.Background(), h.HostName)
			if err != nil {
				cert = &CertificateInfo{
					Status:    CertificateStatusOffline,
					IssuedBy:  "n/a",
					CheckedAt: time.Now().UTC(),
					Error:     err,
				}
			}
			host := h
			host.Certificate = *cert
			results <- host
		}(hst)
	}

	wg.Wait()
	close(results)
	return w.updateData(results)
}

func (w *Worker) updateData(hostch chan Host) error {
	hosts := make([]Host, len(hostch))
	i := 0
	for h := range hostch {
		hosts[i] = h
		i++
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return w.hosts.Update(ctx, hosts)
}

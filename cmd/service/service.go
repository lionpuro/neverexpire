package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lionpuro/neverexpire/config"
	"github.com/lionpuro/neverexpire/db"
	"github.com/lionpuro/neverexpire/domain"
	"github.com/lionpuro/neverexpire/notification"
)

func main() {
	conf := config.FromEnv()
	pool, err := db.NewPool(conf.PostgresURL)
	if err != nil {
		log.Fatal(err)
		return
	}

	ds := domain.NewService(domain.NewRepository(pool))
	ns := notification.NewService(notification.NewRepository(pool))
	monitor := NewMonitor(time.Minute*30, ds, ns)
	notifier := notification.NewNotifier(ns, ds)

	fmt.Println("Starting notification service...")
	go notifier.Start(context.Background())

	fmt.Println("Starting domain monitoring service...")
	monitor.Start()
}

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lionpuro/neverexpire/config"
	"github.com/lionpuro/neverexpire/db"
	"github.com/lionpuro/neverexpire/domain"
	"github.com/lionpuro/neverexpire/logging"
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
	logger := logging.NewLogger()
	updater := domain.NewWorker(30*time.Minute, ds, logger)
	notifier := notification.NewWorker(60*time.Second, ns, ds, logger)

	fmt.Println("Starting notification service...")
	go notifier.Start(context.Background())

	fmt.Println("Starting domain monitoring service...")
	updater.Start()
}

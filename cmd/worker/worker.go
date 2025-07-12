package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lionpuro/neverexpire/config"
	"github.com/lionpuro/neverexpire/db"
	"github.com/lionpuro/neverexpire/hosts"
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

	hs := hosts.NewService(hosts.NewRepository(pool))
	ns := notification.NewService(notification.NewRepository(pool))
	logger := logging.NewLogger()
	updater := hosts.NewWorker(30*time.Minute, hs, logger)
	notifier := notification.NewWorker(60*time.Second, ns, hs, logger)

	fmt.Println("Starting notification service...")
	go notifier.Start(context.Background())

	fmt.Println("Starting monitoring service...")
	updater.Start()
}

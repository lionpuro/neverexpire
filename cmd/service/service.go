package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/lionpuro/neverexpire/config"
	"github.com/lionpuro/neverexpire/db"
	"github.com/lionpuro/neverexpire/domain"
	"github.com/lionpuro/neverexpire/notification"
)

func main() {
	if os.Getenv("APP_ENV") != "production" {
		if err := config.LoadEnvFile(".env"); err != nil {
			log.Fatalf("load env file: %v", err)
		}
	}
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

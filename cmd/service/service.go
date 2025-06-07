package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/lionpuro/neverexpire/db"
	"github.com/lionpuro/neverexpire/domain"
	"github.com/lionpuro/neverexpire/notification"
)

func main() {
	conn := db.ConnString(
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_HOST_PORT"),
		os.Getenv("POSTGRES_DB"),
	)
	pool, err := db.NewPool(conn)
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

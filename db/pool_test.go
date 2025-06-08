package db_test

import (
	"context"
	"testing"

	"github.com/lionpuro/neverexpire/config"
	"github.com/lionpuro/neverexpire/db"
)

func TestNewPool(t *testing.T) {
	conf, err := config.FromEnvFile("../.env.test")
	if err != nil {
		t.Fatalf("failed to load .env.test: %v", err)
	}
	t.Run("create and ping db pool", func(t *testing.T) {
		pool, err := db.NewPool(conf.PostgresURL)
		if err != nil {
			t.Fatalf("new pool: %v", err)
		}
		if err := pool.Ping(context.Background()); err != nil {
			t.Errorf("ping pool: %v", err)
		}
	})
}

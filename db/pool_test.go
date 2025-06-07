package db_test

import (
	"context"
	"testing"

	"github.com/lionpuro/neverexpire/db"
)

func TestNewPool(t *testing.T) {
	t.Run("create and ping db pool", func(t *testing.T) {
		conn := db.ConnString(
			"postgres",
			"password",
			"localhost",
			"5433",
			"testing",
		)
		pool, err := db.NewPool(conn)
		if err != nil {
			t.Fatalf("new pool: %v", err)
		}
		if err := pool.Ping(context.Background()); err != nil {
			t.Errorf("ping pool: %v", err)
		}
	})
}

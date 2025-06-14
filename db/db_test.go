package db_test

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lionpuro/neverexpire/db"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	migrations, err := os.ReadDir("./migrations")
	if err != nil {
		log.Printf("failed to read migrations dir: %v", err)
		return
	}
	scripts := []string{}
	for _, f := range migrations {
		scripts = append(scripts, filepath.Join("./migrations", f.Name()))
	}
	container, err := postgres.Run(ctx,
		"postgres:17.4",
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.WithDatabase("testing"),
		postgres.WithInitScripts(scripts...),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Printf("failed to terminate container: %v", err)
		}
		cancel()
	}()
	if err != nil {
		log.Printf("failed to start container: %v", err)
		return
	}
	addr, err := container.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %v", err)
		return
	}
	pool, err = db.NewPool(addr)
	if err != nil {
		log.Printf("failed to initialize pool: %v", err)
		return
	}
	os.Exit(m.Run())
}

func TestPingPool(t *testing.T) {
	t.Run("ping db pool", func(t *testing.T) {
		if err := pool.Ping(context.Background()); err != nil {
			t.Errorf("ping pool: %v", err)
		}
	})
}

package testutils

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Initializes a postgres container and a connection pool
func NewPostgresConn() (conn *pgxpool.Pool, cleanup func() error, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	var container *postgres.PostgresContainer
	clean := func() error {
		if container != nil {
			if err := testcontainers.TerminateContainer(container); err != nil {
				return fmt.Errorf("failed to terminate container: %v", err)
			}
		}
		cancel()
		return nil
	}
	migrations, err := os.ReadDir("../db/migrations")
	if err != nil {
		log.Printf("failed to read migrations dir: %v", err)
		return
	}
	scripts := []string{}
	for _, f := range migrations {
		if strings.HasSuffix(f.Name(), "up.sql") {
			scripts = append(scripts, filepath.Join("../db/migrations", f.Name()))
		}
	}
	container, err = postgres.Run(ctx,
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
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start container: %v", err)
	}
	addr, err := container.ConnectionString(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get connection string: %v", err)
	}
	pool, err := pgxpool.New(ctx, addr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize pool: %v", err)
	}
	return pool, clean, nil
}

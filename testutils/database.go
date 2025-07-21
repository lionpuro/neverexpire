package testutils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lionpuro/neverexpire/logging"
	"github.com/lionpuro/neverexpire/users"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Initializes a postgres container and a connection pool
func NewDatabase() (conn *pgxpool.Pool, cleanup func() error, err error) {
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
	// populate with test data
	if err := insertTestData(pool); err != nil {
		return nil, nil, fmt.Errorf("failed to populate db with test data: %v", err)
	}
	return pool, clean, nil
}

func insertTestData(conn *pgxpool.Pool) error {
	ctx := context.Background()
	var usrs []users.User
	for range 2 {
		user, err := NewTestUser()
		if err != nil {
			return err
		}
		usrs = append(usrs, user)
	}
	hosts, err := NewTestHosts()
	if err != nil {
		return err
	}

	for _, user := range usrs {
		tx, err := conn.Begin(ctx)
		defer func() {
			if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
				logging.DefaultLogger().Error("failed to rollback tx", "error", err.Error())
			}
		}()
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO users (id, email) VALUES ($1, $2)`, user.ID, user.Email)
		if err != nil {
			return err
		}

		for _, h := range hosts {
			var id int
			var errStr *string = nil
			if h.Certificate.Error != nil {
				str := h.Certificate.Error.Error()
				errStr = &str
			}
			sql := `
			INSERT INTO hosts (
				hostname,
				dns_names,
				ip_address,
				issued_by,
				status,
				expires_at,
				checked_at,
				latency,
				signature,
				error_message
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (hostname) DO UPDATE SET
				dns_names      = EXCLUDED.dns_names,
				ip_address     = EXCLUDED.ip_address,
				issued_by      = EXCLUDED.issued_by,
				status         = EXCLUDED.status,
				expires_at     = EXCLUDED.expires_at,
				checked_at     = EXCLUDED.checked_at,
				latency        = EXCLUDED.latency,
				signature      = EXCLUDED.signature,
				error_message  = EXCLUDED.error_message
			RETURNING id`
			err := tx.QueryRow(ctx, sql,
				h.Hostname,
				h.Certificate.DNSNames,
				h.Certificate.IP,
				h.Certificate.IssuedBy,
				h.Certificate.Status,
				h.Certificate.ExpiresAt,
				h.Certificate.CheckedAt,
				h.Certificate.Latency,
				h.Certificate.Signature,
				errStr,
			).Scan(&id)
			if err != nil {
				return err
			}
			_, err = tx.Exec(ctx, `INSERT INTO user_hosts (host_id, user_id) VALUES ($1, $2)`, id, user.ID)
			if err != nil {
				return err
			}
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}
	return nil
}

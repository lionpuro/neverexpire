package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	Timeout = time.Second * 5
)

func NewPool(dbConn string) (*pgxpool.Pool, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbConn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return pool, nil
}

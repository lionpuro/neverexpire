package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	DB *pgxpool.Pool
}

func NewService(dbConn string) (*Service, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbConn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	s := &Service{
		DB: pool,
	}
	return s, nil
}

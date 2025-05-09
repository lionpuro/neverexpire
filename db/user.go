package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/lionpuro/trackcert/model"
)

func (s *Service) UserByID(ctx context.Context, id string) (model.User, error) {
	rows, err := s.DB.Query(ctx, `SELECT id, email FROM users WHERE id = $1`, id)
	if err != nil {
		return model.User{}, err
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.User])
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (s *Service) CreateUser(ctx context.Context, id, email string) error {
	_, err := s.DB.Exec(ctx, `
		INSERT INTO users (id, email) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, id, email)
	if err != nil {
		return fmt.Errorf("create user: %v", err)
	}

	return nil
}

package user

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lionpuro/trackcerts/db"
	"github.com/lionpuro/trackcerts/model"
)

type Repository interface {
	ByID(ctx context.Context, id string) (model.User, error)
	Create(id, email string) error
	Delete(id string) error
}

type UserRepository struct {
	DB *pgxpool.Pool
}

func NewRepository(dbpool *pgxpool.Pool) *UserRepository {
	return &UserRepository{DB: dbpool}
}

func (r *UserRepository) ByID(ctx context.Context, id string) (model.User, error) {
	rows, err := r.DB.Query(ctx, `SELECT id, email FROM users WHERE id = $1`, id)
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

func (r *UserRepository) Create(id, email string) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	_, err := r.DB.Exec(ctx, `
		INSERT INTO users (id, email) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, id, email)
	return err
}

func (r *UserRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	_, err := r.DB.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

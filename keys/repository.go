package keys

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ByUser(ctx context.Context, uid string) ([]AccessKey, error) {
	q := `SELECT id, hash, user_id, created_at FROM api_keys WHERE user_id = $1`
	rows, err := r.db.Query(ctx, q, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	keys, err := pgx.CollectRows(rows, pgx.RowToStructByName[AccessKey])
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *Repository) ByID(ctx context.Context, id string) (AccessKey, error) {
	q := `SELECT id, hash, user_id, created_at FROM api_keys WHERE id = $1`
	rows, err := r.db.Query(ctx, q, id)
	if err != nil {
		return AccessKey{}, err
	}
	defer rows.Close()
	key, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[AccessKey])
	if err != nil {
		return AccessKey{}, err
	}
	return key, nil
}

func (r *Repository) Create(ctx context.Context, key AccessKey) error {
	q := `
		INSERT INTO api_keys (id, hash, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, hash, user_id, created_at
	`
	_, err := r.db.Exec(ctx, q, key.ID, key.Hash, key.UserID)
	return err
}

func (r *Repository) Delete(ctx context.Context, id, uid string) error {
	q := `DELETE FROM api_keys WHERE id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, q, id, uid)
	return err
}

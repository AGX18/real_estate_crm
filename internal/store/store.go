package store

import (
	"context"

	db "github.com/AGX18/real_estate_crm/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (s *Store) Queries() db.Querier {
	return s.queries
}

func (s *Store) WithTx(ctx context.Context, fn func(q db.Querier) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)
	if err := fn(qtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

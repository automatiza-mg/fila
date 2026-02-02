package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrNotFound é o erro retornado para uma query que não retorna resultado.
	ErrNotFound = errors.New("record not found")
)

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Store struct {
	db DBTX
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{
		db: pool,
	}
}

// WithTx retorna uma nova instância da [Store] usando a [pgx.Tx] para fazer chamadas
// ao banco de dados.
func (s *Store) WithTx(tx pgx.Tx) *Store {
	return &Store{db: tx}
}

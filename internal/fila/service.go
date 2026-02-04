package fila

import (
	"context"

	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SeiService interface {
	ListarUnidades(ctx context.Context) (*sei.ListarUnidadesResponse, error)
}

type Service struct {
	pool  *pgxpool.Pool
	store *database.Store
	sei   SeiService
	cache cache.Cache
}

func NewService(pool *pgxpool.Pool, sei SeiService, cache cache.Cache) *Service {
	return &Service{
		pool:  pool,
		store: database.New(pool),
		sei:   sei,
		cache: cache,
	}
}

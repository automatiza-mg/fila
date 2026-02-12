package tasks

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

const QueueProcessos = "processos"

type RiverOptFunc func(cfg *river.Config)

func WithTestOnly() RiverOptFunc {
	return func(cfg *river.Config) {
		cfg.TestOnly = true
	}
}

func NewQueue(ctx context.Context, pool *pgxpool.Pool, opts ...RiverOptFunc) (*river.Client[pgx.Tx], error) {
	driver := riverpgxv5.New(pool)

	migrator, err := rivermigrate.New(driver, nil)
	if err != nil {
		return nil, err
	}
	if _, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, nil); err != nil {
		return nil, err
	}

	var cfg river.Config
	for _, opt := range opts {
		opt(&cfg)
	}

	return river.NewClient(driver, &cfg)
}

func NewWorker(ctx context.Context, pool *pgxpool.Pool, workers *river.Workers) (*river.Client[pgx.Tx], error) {
	driver := riverpgxv5.New(pool)

	migrator, err := rivermigrate.New(driver, nil)
	if err != nil {
		return nil, err
	}
	if _, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, nil); err != nil {
		return nil, err
	}

	return river.NewClient(driver, &river.Config{
		Workers: workers,
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
			QueueProcessos:     {MaxWorkers: 2},
		},
	})
}

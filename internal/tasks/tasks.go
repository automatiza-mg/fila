package tasks

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

func NewQueue(ctx context.Context, pool *pgxpool.Pool) (*river.Client[pgx.Tx], error) {
	driver := riverpgxv5.New(pool)

	migrator, err := rivermigrate.New(driver, nil)
	if err != nil {
		return nil, err
	}
	if _, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, nil); err != nil {
		return nil, err
	}

	return river.NewClient(driver, &river.Config{})
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

	cfg := &river.Config{}
	if workers != nil {
		cfg.Workers = workers
		cfg.Queues = map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		}
	}

	return river.NewClient(driver, cfg)
}

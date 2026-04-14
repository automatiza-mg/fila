package fila

import (
	"context"
	"fmt"

	"github.com/automatiza-mg/fila/internal/tasks"
)

// EnqueueRecalcularScores enfileira um job para recalcular os scores de todos
// os processos de aposentadoria.
func (s *Service) EnqueueRecalcularScores(ctx context.Context) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = s.queue.InsertTx(ctx, tx, tasks.RecalcularScoresArgs{}, nil)
	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	return tx.Commit(ctx)
}

package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/automatiza-mg/fila/internal/aposentadoria"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

// RecalcularScoresArgs são os argumentos para o job de recálculo de scores.
type RecalcularScoresArgs struct{}

func (args RecalcularScoresArgs) Kind() string {
	return "fila:recalcular-scores"
}

func (args RecalcularScoresArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: river.QueueDefault,
		UniqueOpts: river.UniqueOpts{
			ByPeriod: 5 * time.Minute,
		},
	}
}

// RecalcularScoresWorker processa o recálculo de scores de todos os processos
// de aposentadoria.
type RecalcularScoresWorker struct {
	pool  *pgxpool.Pool
	store *database.Store
	river.WorkerDefaults[RecalcularScoresArgs]
}

func (w *RecalcularScoresWorker) Work(ctx context.Context, job *river.Job[RecalcularScoresArgs]) error {
	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	store := w.store.WithTx(tx)

	paa, err := store.ListAllProcessoAposentadoria(ctx)
	if err != nil {
		return fmt.Errorf("failed to list processos: %w", err)
	}

	updated := 0
	for _, pa := range paa {
		novo := aposentadoria.CalculateScore(
			pa.DataNascimentoRequerente,
			pa.Invalidez,
			pa.Judicial,
			pa.Prioridade,
		)

		if novo == pa.Score {
			continue
		}

		pa.Score = novo
		if err := store.UpdateProcessoAposentadoria(ctx, pa); err != nil {
			return fmt.Errorf("failed to update processo %d: %w", pa.ID, err)
		}
		updated++
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	slog.Info("scores recalculados",
		slog.Int("total", len(paa)),
		slog.Int("atualizados", updated),
	)

	return nil
}

// NewRecalcularScoresWorker cria uma nova instância de [RecalcularScoresWorker].
func NewRecalcularScoresWorker(pool *pgxpool.Pool) *RecalcularScoresWorker {
	return &RecalcularScoresWorker{
		pool:  pool,
		store: database.New(pool),
	}
}

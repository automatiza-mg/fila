package tasks

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/processos"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
)

var _ processos.AnalyzeEnqueuer = (*ProcessoEnqueuer)(nil)

type ProcessoEnqueuer struct {
	Client *river.Client[pgx.Tx]
}

func (p *ProcessoEnqueuer) EnqueueAnalyzeTx(ctx context.Context, tx pgx.Tx, procID uuid.UUID) (bool, error) {
	res, err := p.Client.InsertTx(ctx, tx, AnalyzeProcessoArgs{
		ProcessoID: procID,
	}, nil)
	if err != nil {
		return false, err
	}

	inserted := !res.UniqueSkippedAsDuplicate
	return inserted, nil
}

type AnalyzeProcessoArgs struct {
	ProcessoID uuid.UUID `json:"processo_id"`
}

func (args AnalyzeProcessoArgs) Kind() string {
	return "analyze:processo"
}

func (args AnalyzeProcessoArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: QueueProcessos,
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: time.Hour,
		},
	}
}

type AnalyzeProcessoWorker struct {
	logger    *slog.Logger
	processos *processos.Service
	river.WorkerDefaults[AnalyzeProcessoArgs]
}

func NewAnalyzeProcessoWorker(logger *slog.Logger, service *processos.Service) *AnalyzeProcessoWorker {
	return &AnalyzeProcessoWorker{
		logger:    logger.With(slog.String("worker", "processos")),
		processos: service,
	}
}

func (w *AnalyzeProcessoWorker) Work(ctx context.Context, job *river.Job[AnalyzeProcessoArgs]) error {
	t := time.Now()
	processoID := job.Args.ProcessoID
	logger := w.logger.With(
		slog.String("processo_id", processoID.String()),
		slog.Int64("job_id", job.ID),
		slog.String("queue", job.Queue),
	)

	logger.Info("Iniciando análise de processo")

	err := w.processos.Analyze(ctx, job.Args.ProcessoID)
	if err != nil {
		logger.Error(
			"Análise de processo falhou",
			slog.Any("err", err),
			slog.Duration("duration", time.Since(t)),
		)

		switch {
		case errors.Is(err, database.ErrNotFound), errors.Is(err, pgx.ErrNoRows):
			return river.JobCancel(err)
		default:
			return err
		}
	}

	logger.Info(
		"Análise de processo concluída",
		slog.Duration("duration", time.Since(t)),
	)
	return nil
}

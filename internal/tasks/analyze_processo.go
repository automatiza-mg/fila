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
	w.logger.Info("Iniciando análise de processo", slog.String("processo_id", processoID.String()))

	err := w.processos.Analyze(ctx, job.Args.ProcessoID)
	if err != nil {
		w.logger.Error(
			"Análise de processo falhou",
			slog.String("processo_id", processoID.String()),
			slog.Any("err", err),
			slog.Duration("duration", time.Since(t)),
		)

		switch {
		case errors.Is(err, database.ErrNotFound):
			return river.JobCancel(err)
		case errors.Is(err, pgx.ErrNoRows):
			return river.JobCancel(err)
		default:
			return err
		}
	}

	w.logger.Info(
		"Análise de processo concluída",
		slog.String("processo_id", processoID.String()),
		slog.Duration("duration", time.Since(t)),
	)

	return nil
}

package tasks

import (
	"context"
	"errors"
	"log"
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
	processos *processos.Service
	river.WorkerDefaults[AnalyzeProcessoArgs]
}

func NewAnalyzeProcessoWorker(service *processos.Service) *AnalyzeProcessoWorker {
	return &AnalyzeProcessoWorker{
		processos: service,
	}
}

func (w *AnalyzeProcessoWorker) Work(ctx context.Context, job *river.Job[AnalyzeProcessoArgs]) error {
	log.Printf("Iniciando análise do processo %s", job.Args.ProcessoID)
	err := w.processos.Analyze(ctx, job.Args.ProcessoID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			return river.JobCancel(err)
		case errors.Is(err, pgx.ErrNoRows):
			return river.JobCancel(err)
		default:
			return err
		}
	}
	return nil
}

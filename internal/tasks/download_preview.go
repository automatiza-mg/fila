package tasks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/automatiza-mg/fila/internal/blob"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/docintel"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

type DownloadPreviewArgs struct {
	ProcessoID uuid.UUID `json:"processo_id"`
}

func (args DownloadPreviewArgs) Kind() string {
	return "processo:preview"
}

func (args DownloadPreviewArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       QueueProcessos,
		MaxAttempts: 5,
	}
}

type DownloadPreviewWorker struct {
	store    *database.Store
	arquivos *ArquivoProcessor
	sei      *sei.Client
	river.WorkerDefaults[DownloadPreviewArgs]
}

func NewDownloadPreviewWorker(pool *pgxpool.Pool, storage blob.Storage, sei *sei.Client, cv *docintel.AzureDocIntel) *DownloadPreviewWorker {
	store := database.New(pool)
	arquivos := NewArquivoProcessor(store, storage, cv)

	return &DownloadPreviewWorker{
		store:    store,
		arquivos: arquivos,
		sei:      sei,
	}
}

func (w *DownloadPreviewWorker) Work(ctx context.Context, job *river.Job[DownloadPreviewArgs]) error {
	p, err := w.store.GetProcesso(ctx, job.Args.ProcessoID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return river.JobCancel(err)
		}
		return fmt.Errorf("failed to get processo: %w", err)
	}

	body, err := w.sei.DownloadProcedimento(ctx, p.LinkAcesso)
	if err != nil {
		if errors.Is(err, sei.ErrProcessoVazio) {
			return river.JobCancel(err)
		}
		return fmt.Errorf("failed to download preview: %w", err)
	}
	defer body.Close()

	arq, err := w.arquivos.Process(ctx, body, "application/pdf")
	if err != nil {
		return fmt.Errorf("failed to process preview arquivo: %w", err)
	}

	err = w.store.UpdateProcessoPreviewHash(ctx, p.ID, arq.Hash)
	if err != nil {
		return fmt.Errorf("failed to update processo preview_hash: %w", err)
	}

	return nil
}

func (w *DownloadPreviewWorker) Timeout(job *river.Job[DownloadPreviewArgs]) time.Duration {
	return 30 * time.Second
}

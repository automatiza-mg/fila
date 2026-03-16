package processos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/pipeline"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TextExtractor interface {
	ExtractText(ctx context.Context, r io.Reader, contentType string) (string, error)
}

type AnalyzeEnqueuer interface {
	EnqueueAnalyzeTx(ctx context.Context, tx pgx.Tx, procID uuid.UUID) (bool, error)
}

type DocumentoFetcher interface {
	FetchDocumentos(ctx context.Context, nums []string) ([]DocumentoSei, error)
}

type SeiClient interface {
	ConsultarProcedimento(ctx context.Context, protocolo string) (*sei.ConsultarProcedimentoResponse, error)
	ListarDocumentos(ctx context.Context, linkAcesso string) ([]sei.LinhaDocumento, error)
}

type Service struct {
	pool    *pgxpool.Pool
	store   *database.Store
	sei     SeiClient
	cache   cache.Cache
	queue   AnalyzeEnqueuer
	fetcher DocumentoFetcher
	hooks   []AnalyzeHook
}

type ServiceOpts struct {
	Pool    *pgxpool.Pool
	Sei     SeiClient
	Cache   cache.Cache
	Queue   AnalyzeEnqueuer
	Fetcher DocumentoFetcher
}

func New(opts *ServiceOpts) *Service {
	return &Service{
		pool:    opts.Pool,
		sei:     opts.Sei,
		store:   database.New(opts.Pool),
		cache:   opts.Cache,
		queue:   opts.Queue,
		fetcher: opts.Fetcher,
	}
}

// Analyze busca e processa os documentos de um processo SEI.
func (s *Service) Analyze(ctx context.Context, procID uuid.UUID) error {
	p, err := s.store.GetProcesso(ctx, procID)
	if err != nil {
		return err
	}

	// Pipeline de pré-transação: lista, filtra e busca documentos do SEI.
	state := &pipeline.State{
		ProcessoID: p.ID,
		LinkAcesso: p.LinkAcesso,
		Status:     p.StatusProcessamento,
	}

	preTx := pipeline.New(
		pipeline.ListDocumentos(&seiListerAdapter{sei: s.sei}),
		pipeline.FiltrarNovos(&storeCheckerAdapter{store: s.store}),
		pipeline.BuscarDocumentos(&fetcherAdapter{fetcher: s.fetcher}),
	)

	if err := preTx.Run(ctx, state); err != nil {
		return fmt.Errorf("pre-tx pipeline: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)
	txUpdater := &storeTxUpdater{store: store}
	txPersister := &storeTxPersister{store: store}

	txPipeline := pipeline.New(
		pipeline.AtualizarStatus("PROCESSANDO", txUpdater),
		pipeline.PersistirDocumentos(txPersister),
		pipeline.AtualizarStatus("SUCESSO", txUpdater),
	)

	if err := txPipeline.Run(ctx, state); err != nil {
		return fmt.Errorf("tx pipeline: %w", err)
	}

	// Recarrega o processo com o status atualizado pela pipeline transacional.
	p, err = store.GetProcesso(ctx, procID)
	if err != nil {
		return err
	}

	dd, err := s.listDocumentos(ctx, store, p.ID)
	if err != nil {
		return err
	}

	err = s.notifyAnalyzeCompleteTx(ctx, tx, mapProcesso(p), dd)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// seiListerAdapter adapta SeiClient para pipeline.DocumentLister.
type seiListerAdapter struct {
	sei SeiClient
}

func (a *seiListerAdapter) ListDocumentos(ctx context.Context, linkAcesso string) ([]string, error) {
	docs, err := a.sei.ListarDocumentos(ctx, linkAcesso)
	if err != nil {
		return nil, err
	}
	nums := make([]string, len(docs))
	for i, d := range docs {
		nums[i] = d.Numero
	}
	return nums, nil
}

// storeCheckerAdapter adapta database.Store para pipeline.DocumentChecker.
type storeCheckerAdapter struct {
	store *database.Store
}

func (a *storeCheckerAdapter) ExisteDocumento(ctx context.Context, numero string) (bool, error) {
	_, err := a.store.GetDocumentoByNumero(ctx, numero)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, database.ErrNotFound) {
		return false, nil
	}
	return false, err
}

// fetcherAdapter adapta DocumentoFetcher para pipeline.DocumentFetcher.
type fetcherAdapter struct {
	fetcher DocumentoFetcher
}

func (a *fetcherAdapter) FetchDocumentos(ctx context.Context, nums []string) ([]pipeline.DocBuscado, error) {
	docs, err := a.fetcher.FetchDocumentos(ctx, nums)
	if err != nil {
		return nil, err
	}
	out := make([]pipeline.DocBuscado, len(docs))
	for i, d := range docs {
		metadados, err := json.Marshal(d.APIData)
		if err != nil {
			return nil, err
		}
		out[i] = pipeline.DocBuscado{
			Numero:       d.Numero,
			Conteudo:     d.Conteudo,
			ContentType:  d.ContentType,
			Tipo:         d.Tipo(),
			Unidade:      d.APIData.UnidadeElaboradora.Sigla,
			LinkAcesso:   d.APIData.LinkAcesso,
			MetadadosAPI: metadados,
		}
	}
	return out, nil
}

// storeTxUpdater adapta database.Store (transacional) para pipeline.StatusUpdater.
type storeTxUpdater struct {
	store *database.Store
}

func (a *storeTxUpdater) AtualizarStatus(ctx context.Context, state *pipeline.State) error {
	p, err := a.store.GetProcesso(ctx, state.ProcessoID)
	if err != nil {
		return err
	}
	p.StatusProcessamento = state.Status
	return a.store.UpdateProcesso(ctx, p)
}

// storeTxPersister adapta database.Store (transacional) para pipeline.DocumentPersister.
type storeTxPersister struct {
	store *database.Store
}

func (a *storeTxPersister) PersistirDocumentos(ctx context.Context, state *pipeline.State) error {
	for _, doc := range state.DocumentosBuscados {
		d := &database.Documento{
			Numero:       doc.Numero,
			ProcessoID:   state.ProcessoID,
			Tipo:         doc.Tipo,
			Unidade:      doc.Unidade,
			ContentType:  doc.ContentType,
			OCR:          doc.Conteudo,
			LinkAcesso:   doc.LinkAcesso,
			MetadadosAPI: doc.MetadadosAPI,
		}
		if err := a.store.SaveDocumento(ctx, d); err != nil {
			return err
		}
	}
	return nil
}

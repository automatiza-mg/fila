package processos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/database"
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

	docs, err := s.sei.ListarDocumentos(ctx, p.LinkAcesso)
	if err != nil {
		return err
	}

	seiDocs, err := s.prepareSeiDocs(ctx, docs)
	if err != nil {
		return fmt.Errorf("failed to prepare sei docs: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

	p.StatusProcessamento = "PROCESSANDO"
	err = store.UpdateProcesso(ctx, p)
	if err != nil {
		return err
	}

	err = s.processDocsTx(ctx, store, p, seiDocs)
	if err != nil {
		return fmt.Errorf("failed to process docs: %w", err)
	}

	// Update status to SUCESSO before fetching documents and notifying hooks
	p.StatusProcessamento = "SUCESSO"
	err = store.UpdateProcesso(ctx, p)
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

func (s *Service) prepareSeiDocs(ctx context.Context, docs []sei.LinhaDocumento) ([]DocumentoSei, error) {
	nums := make([]string, 0, len(docs))
	for _, doc := range docs {
		_, err := s.store.GetDocumentoByNumero(ctx, doc.Numero)
		if err == nil {
			continue
		}
		if !errors.Is(err, database.ErrNotFound) {
			return nil, err
		}
		nums = append(nums, doc.Numero)
	}

	// Não há nada para buscar.
	if len(nums) == 0 {
		return []DocumentoSei{}, nil
	}

	return s.fetcher.FetchDocumentos(ctx, nums)
}

func (s *Service) processDocsTx(ctx context.Context, store *database.Store, p *database.Processo, seiDocs []DocumentoSei) error {
	for _, seiDoc := range seiDocs {
		metadados, err := json.Marshal(seiDoc.APIData)
		if err != nil {
			return err
		}

		d := &database.Documento{
			Numero:       seiDoc.Numero,
			ProcessoID:   p.ID,
			Tipo:         seiDoc.Tipo(),
			Unidade:      seiDoc.APIData.UnidadeElaboradora.Sigla,
			ContentType:  seiDoc.ContentType,
			OCR:          seiDoc.Conteudo,
			LinkAcesso:   seiDoc.APIData.LinkAcesso,
			MetadadosAPI: metadados,
		}

		err = store.SaveDocumento(ctx, d)
		if err != nil {
			return err
		}
	}

	return nil
}

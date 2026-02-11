package processos

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/automatiza-mg/fila/internal/aposentadoria"
	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DocumentAnalyzer interface {
	AnalisarAposentadoria(ctx context.Context, docs []*Documento) (*aposentadoria.Analise, error)
}

type TextExtractor interface {
	ExtractText(ctx context.Context, r io.Reader, contentType string) (string, error)
}

type AnalyzeEnqueuer interface {
	EnqueueAnalyzeTx(ctx context.Context, tx pgx.Tx, procID uuid.UUID) (bool, error)
}

type DocumentoFetcher interface {
	FetchDocumentos(ctx context.Context, nums []string) ([]DocumentoSei, error)
}

type Service struct {
	pool     *pgxpool.Pool
	store    *database.Store
	sei      *sei.Client
	cache    cache.Cache
	analyzer DocumentAnalyzer
	queue    AnalyzeEnqueuer
	fetcher  DocumentoFetcher
}

type ServiceOpts struct {
	Pool     *pgxpool.Pool
	Sei      *sei.Client
	Cache    cache.Cache
	Analyzer DocumentAnalyzer
	Queue    AnalyzeEnqueuer
	Fetcher  DocumentoFetcher
}

func New(opts *ServiceOpts) *Service {
	return &Service{
		pool:     opts.Pool,
		sei:      opts.Sei,
		store:    database.New(opts.Pool),
		cache:    opts.Cache,
		analyzer: opts.Analyzer,
		queue:    opts.Queue,
		fetcher:  opts.Fetcher,
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

	p.StatusProcessamento = "PROCESSANDO"
	err = s.store.UpdateProcesso(ctx, p)
	if err != nil {
		return err
	}

	err = s.processDocs(ctx, p, docs)
	if err != nil {
		return fmt.Errorf("failed to process docs: %w", err)
	}

	dd, err := s.ListDocumentos(ctx, p.ID)
	if err != nil {
		return err
	}

	// Analise de IA
	apos, err := s.analyzer.AnalisarAposentadoria(ctx, dd)
	if err != nil {
		return err
	}

	metadados, err := json.Marshal(apos)
	if err != nil {
		return err
	}

	p.MetadadosIA = metadados
	p.AnalisadoEm = sql.Null[time.Time]{
		V:     time.Now(),
		Valid: true,
	}
	p.Aposentadoria = sql.Null[bool]{
		Valid: true,
	}

	// Processo é de aposentadoria
	if apos.Aposentadoria {
		p.Aposentadoria.V = true

		dataNascimento, err := time.Parse(time.DateOnly, apos.DataNascimento)
		if err != nil {
			return err
		}

		dataRequerimento, err := time.Parse(time.DateOnly, apos.DataRequerimento)
		if err != nil {
			return err
		}

		err = s.store.SaveProcessoAposentadoria(ctx, &database.ProcessoAposentadoria{
			ProcessoID:               p.ID,
			CPFRequerente:            apos.CPF,
			Invalidez:                apos.Invalidez,
			Judicial:                 apos.Judicial,
			DataNascimentoRequerente: dataNascimento,
			DataRequerimento:         dataRequerimento,
			Status:                   database.StatusProcessoAnalisePendente,
		})
		if err != nil {
			return err
		}
	}

	p.StatusProcessamento = "SUCESSO"
	err = s.store.UpdateProcesso(ctx, p)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) processDocs(ctx context.Context, p *database.Processo, docs []sei.LinhaDocumento) error {
	nums := make([]string, 0, len(docs))
	for _, doc := range docs {
		_, err := s.store.GetDocumentoByNumero(ctx, doc.Numero)
		if err == nil {
			continue
		}
		if !errors.Is(err, database.ErrNotFound) {
			return err
		}
		nums = append(nums, doc.Numero)
	}

	// Não há nada para buscar.
	if len(nums) == 0 {
		return nil
	}

	seiDocs, err := s.fetcher.FetchDocumentos(ctx, nums)
	if err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

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

	return tx.Commit(ctx)
}

// TriggerReanalysis tenta inserir uma nova análise do processo.
func (s *Service) TriggerReanalysis(ctx context.Context, procID uuid.UUID) error {
	p, err := s.store.GetProcesso(ctx, procID)
	if err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	inserted, err := s.queue.EnqueueAnalyzeTx(ctx, tx, p.ID)
	if err != nil {
		return err
	}

	if inserted {
		p.StatusProcessamento = "PENDENTE"
		err = s.store.WithTx(tx).UpdateProcesso(ctx, p)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

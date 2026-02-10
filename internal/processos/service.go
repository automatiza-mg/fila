package processos

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/automatiza-mg/fila/internal/aposentadoria"
	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/automatiza-mg/fila/internal/soap"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
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

type Service struct {
	pool     *pgxpool.Pool
	store    *database.Store
	sei      *sei.Client
	cache    cache.Cache
	ocr      TextExtractor
	analyzer DocumentAnalyzer
	queue    AnalyzeEnqueuer
}

type ServiceOpts struct {
	Pool     *pgxpool.Pool
	Sei      *sei.Client
	Cache    cache.Cache
	OCR      TextExtractor
	Analyzer DocumentAnalyzer
	Queue    AnalyzeEnqueuer
}

func New(opts *ServiceOpts) *Service {
	return &Service{
		pool:     opts.Pool,
		store:    database.New(opts.Pool),
		sei:      opts.Sei,
		cache:    opts.Cache,
		ocr:      opts.OCR,
		analyzer: opts.Analyzer,
		queue:    opts.Queue,
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

		nasc, err := time.Parse(time.DateOnly, apos.DataNascimento)
		if err != nil {
			return err
		}

		requeri, err := time.Parse(time.DateOnly, apos.DataRequerimento)
		if err != nil {
			return err
		}

		err = s.store.SaveProcessoAposentadoria(ctx, &database.ProcessoAposentadoria{
			ProcessoID:               p.ID,
			CPFRequerente:            apos.CPF,
			Invalidez:                apos.Invalidez,
			Judicial:                 apos.Judicial,
			DataNascimentoRequerente: nasc,
			DataRequerimento:         requeri,
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
	g := new(errgroup.Group)
	g.SetLimit(5)

	dd := make([]*database.Documento, len(docs))

	for i, doc := range docs {
		g.Go(func() error {
			_, err := s.store.GetDocumentoByNumero(ctx, doc.Numero)
			if err == nil {
				return nil
			}
			if !errors.Is(err, database.ErrNotFound) {
				return err
			}

			resp, err := s.sei.ConsultarDocumento(ctx, doc.Numero)
			if err != nil {
				var soapError *soap.Error
				switch {
				case errors.As(err, &soapError):
					return nil
				default:
					return err
				}
			}

			metadados, err := json.Marshal(resp.Parametros)
			if err != nil {
				return err
			}

			res, err := http.Get(resp.Parametros.LinkAcesso)
			if err != nil {
				return err
			}
			defer res.Body.Close()

			contentType := res.Header.Get("Content-Type")
			text, err := s.ocr.ExtractText(ctx, res.Body, contentType)
			if err != nil {
				return err
			}

			tipo := resp.Parametros.Serie.Nome
			if resp.Parametros.Numero != "" {
				tipo += " " + resp.Parametros.Numero
			}

			dd[i] = &database.Documento{
				Numero:       doc.Numero,
				ProcessoID:   p.ID,
				Tipo:         tipo,
				OCR:          text,
				Unidade:      resp.Parametros.UnidadeElaboradora.Sigla,
				LinkAcesso:   resp.Parametros.LinkAcesso,
				ContentType:  contentType,
				MetadadosAPI: metadados,
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)
	for _, d := range dd {
		if d == nil {
			continue
		}

		err := store.SaveDocumento(ctx, d)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
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

package processos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/automatiza-mg/fila/internal/soap"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

type TextExtractor interface {
	// TODO: Implementar ExtractTextFromURL.
	ExtractText(ctx context.Context, r io.Reader, contentType string) (string, error)
}

type Service struct {
	pool  *pgxpool.Pool
	store *database.Store
	sei   *sei.Client
	cache cache.Cache
	ocr   TextExtractor
}

type ServiceOpts struct {
	Pool  *pgxpool.Pool
	Sei   *sei.Client
	Cache cache.Cache
	OCR   TextExtractor
}

func New(opts *ServiceOpts) *Service {
	return &Service{
		pool:  opts.Pool,
		store: database.New(opts.Pool),
		sei:   opts.Sei,
		cache: opts.Cache,
		ocr:   opts.OCR,
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

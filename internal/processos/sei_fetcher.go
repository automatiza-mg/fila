package processos

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/automatiza-mg/fila/internal/docintel"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/automatiza-mg/fila/internal/soap"
	"golang.org/x/sync/errgroup"
)

type SeiFetcher struct {
	sei *sei.Client
	ocr *docintel.AzureDocIntel
}

func NewSeiFetcher(sei *sei.Client, ocr *docintel.AzureDocIntel) *SeiFetcher {
	return &SeiFetcher{
		sei: sei,
		ocr: ocr,
	}
}

type DocumentoSei struct {
	Numero      string
	Conteudo    string
	ContentType string
	APIData     sei.RetornoConsultaDocumento
}

// Tipo retorna o tipo do documento formatado.
func (d DocumentoSei) Tipo() string {
	if d.APIData.Numero == "" {
		return d.APIData.Serie.Nome
	}
	return fmt.Sprintf("%s %s", d.APIData.Serie.Nome, d.APIData.Numero)
}

// FetchDocumentos consulta uma lista de números de documentos no SEI e extrai
// o texto dos documentos usando OCR (Azure Doc Intel).
func (s *SeiFetcher) FetchDocumentos(ctx context.Context, nums []string) ([]DocumentoSei, error) {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(5)

	docs := make([]*DocumentoSei, len(nums))

	for i, num := range nums {
		g.Go(func() error {
			resp, err := s.sei.ConsultarDocumento(ctx, num)
			if err != nil {
				var soapError *soap.Error
				switch {
				case errors.As(err, &soapError):
					// Documento pode ter sido cancelado, etc.
					return nil
				default:
					return err
				}
			}

			res, err := http.Get(resp.Parametros.LinkAcesso)
			if err != nil {
				return err
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				return fmt.Errorf("failed to download documento %s: %d", num, res.StatusCode)
			}

			contentType := res.Header.Get("Content-Type")
			text, err := s.ocr.ExtractText(ctx, res.Body, contentType)
			if err != nil {
				return err
			}

			tipo := resp.Parametros.Serie.Nome
			if resp.Parametros.Numero != "" {
				tipo += " " + resp.Parametros.Numero
			}

			docs[i] = &DocumentoSei{
				Numero:      num,
				Conteudo:    text,
				ContentType: contentType,
				APIData:     resp.Parametros,
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Remove os documentos nulos.
	out := make([]DocumentoSei, 0, len(docs))
	for _, doc := range docs {
		if doc != nil {
			out = append(out, *doc)
		}
	}

	return out, nil
}

package processos

import (
	"context"
	"encoding/json"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/google/uuid"
)

type Assinatura struct {
	Nome string `json:"nome"`
	CPF  string `json:"cpf"`
}

type Documento struct {
	ID              int64        `json:"id"`
	Numero          string       `json:"numero"`
	Tipo            string       `json:"tipo"`
	Conteudo        string       `json:"conteudo"`
	LinkAcesso      string       `json:"-"`
	ContentType     string       `json:"content_type"`
	Data            string       `json:"data"`
	UnidadeGeradora string       `json:"unidade_geradora"`
	Assinaturas     []Assinatura `json:"assinaturas"`
}

func mapDocumento(d *database.Documento, a *database.Arquivo) (*Documento, error) {
	doc := Documento{
		ID:              d.ID,
		Numero:          d.Numero,
		Tipo:            d.Tipo,
		Conteudo:        a.Conteudo,
		LinkAcesso:      d.LinkAcesso,
		ContentType:     a.ContentType,
		UnidadeGeradora: d.Unidade,
	}

	var resp sei.RetornoConsultaDocumento
	err := json.Unmarshal(d.MetadadosAPI, &resp)
	if err != nil {
		return nil, err
	}

	doc.Data = resp.Data
	doc.Assinaturas = make([]Assinatura, len(resp.Assinaturas.Itens))
	for i, a := range resp.Assinaturas.Itens {
		doc.Assinaturas[i] = Assinatura{
			Nome: a.Nome,
			CPF:  a.Sigla,
		}
	}

	return &doc, nil
}

// ListDocumentos retorna a lista de documentos de um processo SEI.
func (s *Service) ListDocumentos(ctx context.Context, processoID uuid.UUID) ([]*Documento, error) {
	return s.listDocumentos(ctx, s.store, processoID)
}

func (s *Service) listDocumentos(ctx context.Context, store *database.Store, processoID uuid.UUID) ([]*Documento, error) {
	dd, err := store.ListDocumentos(ctx, processoID)
	if err != nil {
		return nil, err
	}

	arquivoMap, err := s.loadArquivosMap(ctx, store, dd)
	if err != nil {
		return nil, err
	}

	docs := make([]*Documento, len(dd))
	for i, d := range dd {
		arq := arquivoMap[d.ArquivoHash]
		doc, err := mapDocumento(d, arq)
		if err != nil {
			return nil, err
		}
		docs[i] = doc
	}

	return docs, nil
}

func (s *Service) loadArquivosMap(ctx context.Context, store *database.Store, dd []*database.Documento) (map[string]*database.Arquivo, error) {
	hashes := make([]string, 0, len(dd))
	for _, d := range dd {
		hashes = append(hashes, d.ArquivoHash)
	}
	return store.GetArquivosMap(ctx, hashes)
}

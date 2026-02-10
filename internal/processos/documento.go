package processos

import (
	"context"
	"encoding/json"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/sei"
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
	LinkAcesso      string       `json:"link_acesso"`
	UnidadeGeradora string       `json:"unidade_geradora"`
	Assinaturas     []Assinatura `json:"assinaturas"`
}

func mapDocumento(d *database.Documento) (*Documento, error) {
	doc := Documento{
		ID:              d.ID,
		Numero:          d.Numero,
		Tipo:            d.Tipo,
		Conteudo:        d.OCR,
		LinkAcesso:      d.LinkAcesso,
		UnidadeGeradora: d.Unidade,
	}

	var resp sei.RetornoConsultaDocumento
	err := json.Unmarshal(d.MetadadosAPI, &resp)
	if err != nil {
		return nil, err
	}

	doc.Assinaturas = make([]Assinatura, len(resp.Assinaturas.Itens))
	for i, a := range resp.Assinaturas.Itens {
		doc.Assinaturas[i] = Assinatura{
			Nome: a.Nome,
			CPF:  a.Sigla,
		}
	}

	return &doc, nil
}

func (s *Service) GetDocumentoByNumero(ctx context.Context, numero string) (*Documento, error) {
	d, err := s.store.GetDocumentoByNumero(ctx, numero)
	if err != nil {
		return nil, err
	}
	return mapDocumento(d)
}

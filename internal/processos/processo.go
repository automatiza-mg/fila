package processos

import (
	"context"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/google/uuid"
)

type Processo struct {
	ID     uuid.UUID `json:"id"`
	Numero string    `json:"numero"`
	Status string    `json:"status"`

	Documentos []*Documento `json:"documentos"`
}

func (s *Service) buildProcesso(ctx context.Context, p *database.Processo) (*Processo, error) {
	proc := &Processo{
		ID:     p.ID,
		Numero: p.Numero,
		Status: p.StatusProcessamento,
	}

	dd, err := s.store.ListDocumentos(ctx, p.ID)
	if err != nil {
		return nil, err
	}

	proc.Documentos = make([]*Documento, len(dd))
	for i, d := range dd {
		doc, err := mapDocumento(d)
		if err != nil {
			return nil, err
		}
		proc.Documentos[i] = doc
	}

	return proc, nil
}

func (s *Service) CreateProcesso(ctx context.Context, numeroProcesso string) (*Processo, error) {
	resp, err := s.sei.ConsultarProcedimento(ctx, numeroProcesso)
	if err != nil {
		return nil, err
	}

	unidade := resp.Parametros.AndamentoGeracao.Unidade
	linkAcesso := resp.Parametros.LinkAcesso

	p := &database.Processo{
		Numero:              numeroProcesso,
		StatusProcessamento: "PENDENTE",
		LinkAcesso:          linkAcesso,
		SeiUnidadeID:        unidade.IdUnidade,
		SeiUnidadeSigla:     unidade.Sigla,
	}
	err = s.store.SaveProcesso(ctx, p)
	if err != nil {
		return nil, err
	}

	return s.buildProcesso(ctx, p)
}

// GetProcessoByNumero retorna os dados de um processo pelo numero (protocolo).
func (s *Service) GetProcessoByNumero(ctx context.Context, numeroProcesso string) (*Processo, error) {
	p, err := s.store.GetProcessoByNumero(ctx, numeroProcesso)
	if err != nil {
		return nil, err
	}
	return s.buildProcesso(ctx, p)
}

// GetProcesso retorna os dados de um processo pelo ID.
func (s *Service) GetProcesso(ctx context.Context, processoID uuid.UUID) (*Processo, error) {
	p, err := s.store.GetProcesso(ctx, processoID)
	if err != nil {
		return nil, err
	}
	return s.buildProcesso(ctx, p)
}

package processos

import (
	"context"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/google/uuid"
)

type Processo struct {
	ID              uuid.UUID `json:"id"`
	Numero          string    `json:"numero"`
	Status          string    `json:"status"`
	LinkAcesso      string    `json:"link_acesso"`
	SeiUnidadeID    string    `json:"sei_unidade_id"`
	SeiUnidadeSigla string    `json:"sei_unidade_sigla"`
}

func mapProcesso(p *database.Processo) *Processo {
	return &Processo{
		ID:              p.ID,
		Numero:          p.Numero,
		Status:          p.StatusProcessamento,
		LinkAcesso:      p.LinkAcesso,
		SeiUnidadeID:    p.SeiUnidadeID,
		SeiUnidadeSigla: p.SeiUnidadeSigla,
	}
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

	return mapProcesso(p), nil
}

// GetProcessoByNumero retorna os dados de um processo pelo numero (protocolo).
func (s *Service) GetProcessoByNumero(ctx context.Context, numeroProcesso string) (*Processo, error) {
	p, err := s.store.GetProcessoByNumero(ctx, numeroProcesso)
	if err != nil {
		return nil, err
	}
	return mapProcesso(p), nil
}

// GetProcesso retorna os dados de um processo pelo ID.
func (s *Service) GetProcesso(ctx context.Context, processoID uuid.UUID) (*Processo, error) {
	p, err := s.store.GetProcesso(ctx, processoID)
	if err != nil {
		return nil, err
	}
	return mapProcesso(p), nil
}

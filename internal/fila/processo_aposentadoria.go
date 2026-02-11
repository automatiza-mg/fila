package fila

import (
	"context"
	"encoding/json"
	"time"

	"github.com/automatiza-mg/fila/internal/aposentadoria"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/google/uuid"
)

// Processo é um processo de aposentadoria processado pelo sistema.
type Processo struct {
	ID                       int64                  `json:"id"`
	ProcessoID               uuid.UUID              `json:"processo_id"`
	Numero                   string                 `json:"numero"`
	DataRequerimento         time.Time              `json:"data_requerimento"`
	CPFRequerente            string                 `json:"cpf_requerente"`
	DataNascimentoRequerente time.Time              `json:"data_nascimento_requerente"`
	Invalidez                bool                   `json:"invalidez"`
	Judicial                 bool                   `json:"judicial"`
	Prioridade               bool                   `json:"prioridade"`
	Score                    int                    `json:"score"`
	Status                   string                 `json:"status"`
	AnalistaID               *int64                 `json:"analista_id"`
	AnaliseIA                *aposentadoria.Analise `json:"analise_ia"`
	CriadoEm                 time.Time              `json:"criado_em"`
	AtualizadoEm             time.Time              `json:"atualizado_em"`
}

func mapProcesso(pa *database.ProcessoAposentadoria, numero string) (*Processo, error) {
	var analiseIA *aposentadoria.Analise
	if pa.MetadadosIA != nil {
		analiseIA = &aposentadoria.Analise{}
		err := json.Unmarshal(pa.MetadadosIA, &analiseIA)
		if err != nil {
			return nil, err
		}
	}

	return &Processo{
		ID:                       pa.ID,
		ProcessoID:               pa.ProcessoID,
		Numero:                   numero,
		DataRequerimento:         pa.DataRequerimento,
		CPFRequerente:            pa.CPFRequerente,
		DataNascimentoRequerente: pa.DataNascimentoRequerente,
		Invalidez:                pa.Invalidez,
		Judicial:                 pa.Judicial,
		Prioridade:               pa.Prioridade,
		Score:                    pa.Score,
		Status:                   string(pa.Status),
		AnalistaID:               database.Ptr(pa.AnalistaID),
		AnaliseIA:                analiseIA,
		CriadoEm:                 pa.CriadoEm,
		AtualizadoEm:             pa.AtualizadoEm,
	}, nil
}

// GetProcesso retorna um processo de aposentadoria pelo ID.
func (s *Service) GetProcesso(ctx context.Context, id int64) (*Processo, error) {
	pa, err := s.store.GetProcessoAposentadoria(ctx, id)
	if err != nil {
		return nil, err
	}

	p, err := s.store.GetProcesso(ctx, pa.ProcessoID)
	if err != nil {
		return nil, err
	}

	return mapProcesso(pa, p.Numero)
}

// GetProcessoByNumero retorna um processo de aposentadoria pelo
// número do processo SEI.
func (s *Service) GetProcessoByNumero(ctx context.Context, numero string) (*Processo, error) {
	pa, err := s.store.GetProcessoAposentadoriaByNumero(ctx, numero)
	if err != nil {
		return nil, err
	}

	return mapProcesso(pa, numero)
}

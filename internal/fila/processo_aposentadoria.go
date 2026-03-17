package fila

import (
	"context"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/pagination"
	"github.com/google/uuid"
)

// Processo é um processo de aposentadoria processado pelo sistema.
type ProcessoAposentadoria struct {
	ID                       int64     `json:"id"`
	ProcessoID               uuid.UUID `json:"processo_id"`
	Numero                   string    `json:"numero"`
	DataRequerimento         time.Time `json:"data_requerimento"`
	CPFRequerente            string    `json:"cpf_requerente"`
	DataNascimentoRequerente time.Time `json:"data_nascimento_requerente"`
	Invalidez                bool      `json:"invalidez"`
	Judicial                 bool      `json:"judicial"`
	Prioridade               bool      `json:"prioridade"`
	Score                    int       `json:"score"`
	Status                   string    `json:"status"`
	Analista                 *string   `json:"analista"`
	AnalistaID               *int64    `json:"analista_id"`
	CriadoEm                 time.Time `json:"criado_em"`
	AtualizadoEm             time.Time `json:"atualizado_em"`
}

func mapProcesso(pa *database.ProcessoAposentadoria, numero string, analista *string) *ProcessoAposentadoria {
	return &ProcessoAposentadoria{
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
		CriadoEm:                 pa.CriadoEm,
		AtualizadoEm:             pa.AtualizadoEm,
	}
}

// GetProcessoAposentadoria retorna um processo de aposentadoria pelo ID.
func (s *Service) GetProcessoAposentadoria(ctx context.Context, id int64) (*ProcessoAposentadoria, error) {
	pa, err := s.store.GetProcessoAposentadoria(ctx, id)
	if err != nil {
		return nil, err
	}

	numero, err := s.store.GetNumeroProcessoAposentadoria(ctx, pa.ID)
	if err != nil {
		return nil, err
	}

	var analista *string
	if pa.AnalistaID.Valid {
		nome, err := s.store.GetNomeAnalista(ctx, pa.AnalistaID.V)
		if err != nil {
			return nil, err
		}
		analista = &nome
	}

	return mapProcesso(pa, numero, analista), nil
}

// GetProcessoAposentadoriaByNumero retorna um processo de aposentadoria pelo
// número do processo SEI.
func (s *Service) GetProcessoAposentadoriaByNumero(ctx context.Context, numero string) (*ProcessoAposentadoria, error) {
	pa, err := s.store.GetProcessoAposentadoriaByNumero(ctx, numero)
	if err != nil {
		return nil, err
	}

	var analista *string
	if pa.AnalistaID.Valid {
		nome, err := s.store.GetNomeAnalista(ctx, pa.AnalistaID.V)
		if err != nil {
			return nil, err
		}
		analista = &nome
	}

	return mapProcesso(pa, numero, analista), nil
}

type ListProcessoAposentadoriaParams struct {
	Numero string
	Status string
	Page   int
	Limit  int
}

// ListProcesso retorna a lista paginada dos processos de aposentadoria com seus numeros.
func (s *Service) ListProcesso(ctx context.Context, params ListProcessoAposentadoriaParams) (*pagination.Result[*ProcessoAposentadoria], error) {
	offset := pagination.Offset(params.Page, params.Limit)

	paa, totalCount, err := s.store.ListProcessoAposentadoria(ctx, database.ListProcessoAposentadoriaParams{
		Numero: params.Numero,
		Status: params.Status,
		Limit:  params.Limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	if len(paa) == 0 {
		return pagination.NewResult[*ProcessoAposentadoria](nil, params.Page, 0, params.Limit), nil
	}

	processos := make([]*ProcessoAposentadoria, 0, len(paa))

	// Busca os números dos processos.
	for _, pa := range paa {
		numero, err := s.store.GetNumeroProcessoAposentadoria(ctx, pa.ID)
		if err != nil {
			return nil, err
		}

		var analista *string
		if pa.AnalistaID.Valid {
			nome, err := s.store.GetNomeAnalista(ctx, pa.AnalistaID.V)
			if err != nil {
				return nil, err
			}
			analista = &nome
		}

		processos = append(processos, mapProcesso(pa, numero, analista))
	}

	return pagination.NewResult(processos, params.Page, totalCount, params.Limit), nil
}

// GetProcessoAtribuido retorna o processo de aposentadoria atribuído a um analista.
// Retorna database.ErrNotFound se o analista não tiver um processo EM_ANALISE.
func (s *Service) GetProcessoAtribuido(ctx context.Context, analistaID int64) (*ProcessoAposentadoria, error) {
	pa, err := s.store.GetProcessoAtribuido(ctx, analistaID)
	if err != nil {
		return nil, err
	}

	p, err := s.store.GetProcesso(ctx, pa.ProcessoID)
	if err != nil {
		return nil, err
	}

	var analista *string
	if pa.AnalistaID.Valid {
		nome, err := s.store.GetNomeAnalista(ctx, pa.AnalistaID.V)
		if err != nil {
			return nil, err
		}
		analista = &nome
	}

	return mapProcesso(pa, p.Numero, analista), nil
}

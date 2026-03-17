package fila

import (
	"context"
	"fmt"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/pagination"
)

type SolicitacaoPrioridade struct {
	ID                      int64     `json:"id"`
	NumeroProcesso          string    `json:"numero_processo"`
	ProcessoAposentadoriaID int64     `json:"processo_aposentadoria_id"`
	Justificativa           string    `json:"justificativa"`
	Status                  string    `json:"status"`
	CriadoEm                time.Time `json:"criado_em"`
	AtualizadoEm            time.Time `json:"atualizado_em"`
}

func mapSolicitacaoPrioridade(sp *database.SolicitacaoPrioridade, numero string) *SolicitacaoPrioridade {
	return &SolicitacaoPrioridade{
		ID:                      sp.ID,
		NumeroProcesso:          numero,
		ProcessoAposentadoriaID: sp.ProcessoAposentadoriaID,
		Justificativa:           sp.Justificativa,
		Status:                  sp.Status,
		CriadoEm:                sp.CriadoEm,
		AtualizadoEm:            sp.AtualizadoEm,
	}
}

type SolicitarPrioridadeParams struct {
	ProcessoAposentadoriaID int64
	UsuarioID               int64
	Justificativa           string
}

func (s *Service) CreateSolicitacaoPrioridade(ctx context.Context, params SolicitarPrioridadeParams) (*SolicitacaoPrioridade, error) {
	sp := &database.SolicitacaoPrioridade{
		ProcessoAposentadoriaID: params.ProcessoAposentadoriaID,
		Justificativa:           params.Justificativa,
		Status:                  "pendente",
		UsuarioID:               params.UsuarioID,
	}

	err := s.store.SaveSolicitacaoPrioridade(ctx, sp)
	if err != nil {
		return nil, err
	}

	numero, err := s.store.GetNumeroProcessoAposentadoria(ctx, sp.ProcessoAposentadoriaID)
	if err != nil {
		return nil, err
	}

	return mapSolicitacaoPrioridade(sp, numero), nil
}

func (s *Service) GetSolicitacaoPrioridade(ctx context.Context, spID int64) (*SolicitacaoPrioridade, error) {
	sp, err := s.store.GetSolicitacaoPrioridade(ctx, spID)
	if err != nil {
		return nil, err
	}

	numero, err := s.store.GetNumeroProcessoAposentadoria(ctx, sp.ProcessoAposentadoriaID)
	if err != nil {
		return nil, err
	}

	return mapSolicitacaoPrioridade(sp, numero), nil
}

type ListSolicitacoesPrioridadeParams struct {
	ProcessoAposentadoriaID int64
	Page                    int
	Limit                   int
}

func (s *Service) ListSolicitacoesPrioridade(ctx context.Context, params ListSolicitacoesPrioridadeParams) (*pagination.Result[*SolicitacaoPrioridade], error) {
	ssp, totalCount, err := s.store.ListSolicitacoesPrioridade(ctx, database.ListSolicitacoesPrioridadeParams{
		ProcessoAposentadoriaID: params.ProcessoAposentadoriaID,
		Limit:                   params.Limit,
		Offset:                  pagination.Offset(params.Page, params.Limit),
	})
	if err != nil {
		return nil, err
	}

	solicitacoes := make([]*SolicitacaoPrioridade, 0, len(ssp))
	for _, sp := range ssp {
		numero, err := s.store.GetNumeroProcessoAposentadoria(ctx, sp.ProcessoAposentadoriaID)
		if err != nil {
			return nil, err
		}
		solicitacoes = append(solicitacoes, mapSolicitacaoPrioridade(sp, numero))
	}

	return pagination.NewResult(solicitacoes, params.Page, totalCount, params.Limit), nil
}

func (s *Service) AprovarSolicitacaoPrioridade(ctx context.Context, spID int64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

	sp, err := store.GetSolicitacaoPrioridade(ctx, spID)
	if err != nil {
		return fmt.Errorf("failed to get solicitacao: %w", err)
	}

	pa, err := store.GetProcessoAposentadoria(ctx, sp.ProcessoAposentadoriaID)
	if err != nil {
		return fmt.Errorf("failed to get processo: %w", err)
	}

	// Atualiza o processo de aposentadoria.
	pa.Prioridade = true
	err = store.UpdateProcessoAposentadoria(ctx, pa)
	if err != nil {
		return err
	}

	// Atualiza a solicitação.
	sp.Status = "aprovado"
	err = store.UpdateSolicitacaoPrioridade(ctx, sp)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Service) NegarSolicitacaoPrioridade(ctx context.Context, spID int64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

	sp, err := store.GetSolicitacaoPrioridade(ctx, spID)
	if err != nil {
		return err
	}

	pa, err := store.GetProcessoAposentadoria(ctx, sp.ProcessoAposentadoriaID)
	if err != nil {
		return err
	}

	// Atualiza o processo de aposentadoria.
	pa.Prioridade = false
	err = store.UpdateProcessoAposentadoria(ctx, pa)
	if err != nil {
		return err
	}

	// Atualiza a solicitação.
	sp.Status = "negado"
	err = store.UpdateSolicitacaoPrioridade(ctx, sp)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

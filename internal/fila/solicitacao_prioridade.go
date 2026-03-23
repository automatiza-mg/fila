package fila

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/automatiza-mg/fila/internal/auth"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/mail"
	"github.com/automatiza-mg/fila/internal/pagination"
	"github.com/automatiza-mg/fila/internal/tasks"
	"github.com/jackc/pgx/v5"
)

const (
	// O valor que deve ser adicionado / subtraído do score nos casos de prioridade.
	prioScore = 6
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
	SolicitacaoURL          func(numero string) string
}

// CreateSolicitacaoPrioridade cria uma solicitação de priorização de um
// processo a ser analisada por um usuário com papel de SUBSECRETARIO.
func (s *Service) CreateSolicitacaoPrioridade(ctx context.Context, params SolicitarPrioridadeParams) (*SolicitacaoPrioridade, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

	sp := &database.SolicitacaoPrioridade{
		ProcessoAposentadoriaID: params.ProcessoAposentadoriaID,
		Justificativa:           strings.TrimSpace(params.Justificativa),
		Status:                  "pendente",
		UsuarioID:               params.UsuarioID,
	}

	err = store.SaveSolicitacaoPrioridade(ctx, sp)
	if err != nil {
		return nil, err
	}

	numero, err := store.GetNumeroProcessoAposentadoria(ctx, sp.ProcessoAposentadoriaID)
	if err != nil {
		return nil, err
	}

	url := ""
	if params.SolicitacaoURL != nil {
		url = params.SolicitacaoURL(numero)
	}

	if err := s.notifyPrioridadeCreated(ctx, tx, numero, sp.Justificativa, url); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return mapSolicitacaoPrioridade(sp, numero), nil
}

// Envia notificação ao(s) subsecretário(s) cadastrados no sistema de uma nova
// solicitação de prioridade.
func (s *Service) notifyPrioridadeCreated(ctx context.Context, tx pgx.Tx, numero, justificativa, solicitacaoURL string) error {
	store := s.store.WithTx(tx)

	subs, err := store.ListEmailsByPapel(ctx, auth.PapelSubsecretario)
	if err != nil {
		return err
	}
	if len(subs) == 0 {
		return nil
	}

	email, err := mail.NewPrioridadeEmail(subs, mail.PrioridadeEmailParams{
		NumeroProcesso: numero,
		Justificativa:  justificativa,
		SolicitacaoURL: solicitacaoURL,
	})
	if err != nil {
		return err
	}

	_, err = s.queue.InsertTx(ctx, tx, tasks.SendEmailArgs{
		Email: email,
	}, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetSolicitacaoPrioridade retorna os dados básicos de uma solicitação de
// prioridade.
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
	Status                  string
	Numero                  string
	Page                    int
	Limit                   int
}

// ListSolicitacoesPrioridade retorna a lista paginada de solicitações de
// prioridade.
func (s *Service) ListSolicitacoesPrioridade(ctx context.Context, params ListSolicitacoesPrioridadeParams) (*pagination.Result[*SolicitacaoPrioridade], error) {
	ssp, totalCount, err := s.store.ListSolicitacoesPrioridade(ctx, database.ListSolicitacoesPrioridadeParams{
		ProcessoAposentadoriaID: params.ProcessoAposentadoriaID,
		Status:                  params.Status,
		Numero:                  params.Numero,
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

// AprovarSolicitacaoPrioridade marca um processo como prioritário a partir
// de uma solicitação criada por um gestor.
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

	if !pa.Prioridade {
		pa.Score += prioScore
	}

	pa.Prioridade = true
	err = store.UpdateProcessoAposentadoria(ctx, pa)
	if err != nil {
		return err
	}

	sp.Status = "aprovado"
	err = store.UpdateSolicitacaoPrioridade(ctx, sp)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// NegarSolicitacaoPrioridade marca um processo como não prioritário a partir
// de uma solicitação criada por um gestor.
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

	if pa.Prioridade {
		pa.Score -= prioScore
	}
	pa.Prioridade = false
	err = store.UpdateProcessoAposentadoria(ctx, pa)
	if err != nil {
		return err
	}

	sp.Status = "negado"
	err = store.UpdateSolicitacaoPrioridade(ctx, sp)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

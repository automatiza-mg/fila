package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	StatusPrioridadePendente = "pendente"
	StatusPrioridadeAprovado = "aprovado"
	StatusPrioridadeNegado   = "negado"
)

type SolicitacaoPrioridade struct {
	ID                      int64
	ProcessoAposentadoriaID int64
	Justificativa           string
	Status                  string // pendente, aprovado, negado
	UsuarioID               int64
	CriadoEm                time.Time
	AtualizadoEm            time.Time
}

func (s *Store) SaveSolicitacaoPrioridade(ctx context.Context, sp *SolicitacaoPrioridade) error {
	q := `
	INSERT INTO solicitacoes_prioridade (processo_aposentadoria_id, justificativa, status, usuario_id)
	VALUES ($1, $2, $3, $4)
	RETURNING id, criado_em, atualizado_em`
	args := []any{sp.ProcessoAposentadoriaID, sp.Justificativa, sp.Status, sp.UsuarioID}

	err := s.db.QueryRow(ctx, q, args...).Scan(&sp.ID, &sp.CriadoEm, &sp.AtualizadoEm)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetSolicitacaoPrioridade(ctx context.Context, spID int64) (*SolicitacaoPrioridade, error) {
	q := `
	SELECT 
		id,
		processo_aposentadoria_id,
		justificativa,
		status,
		usuario_id,
		criado_em,
		atualizado_em
	FROM solicitacoes_prioridade
	WHERE id = $1`

	var sp SolicitacaoPrioridade
	err := s.db.QueryRow(ctx, q, spID).Scan(
		&sp.ID,
		&sp.ProcessoAposentadoriaID,
		&sp.Justificativa,
		&sp.Status,
		&sp.UsuarioID,
		&sp.CriadoEm,
		&sp.AtualizadoEm,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &sp, nil
}

type ListSolicitacoesPrioridadeParams struct {
	ProcessoAposentadoriaID int64
	Status                  string
	Limit                   int
	Offset                  int
}

func (s *Store) ListSolicitacoesPrioridade(ctx context.Context, params ListSolicitacoesPrioridadeParams) ([]*SolicitacaoPrioridade, int, error) {
	q := `
	SELECT 
		id,
		processo_aposentadoria_id,
		justificativa,
		status,
		usuario_id,
		criado_em,
		atualizado_em,
		COUNT(*) OVER()
	FROM solicitacoes_prioridade
	WHERE (processo_aposentadoria_id = $1 OR $1 = 0)
	AND (status = $2 OR $2 = '')
	ORDER BY id
	LIMIT $3 OFFSET $4`

	totalCount := 0
	ssp := make([]*SolicitacaoPrioridade, 0)

	args := []any{
		params.ProcessoAposentadoriaID,
		params.Status,
		params.Limit,
		params.Offset,
	}

	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var sp SolicitacaoPrioridade
		err := rows.Scan(
			&sp.ID,
			&sp.ProcessoAposentadoriaID,
			&sp.Justificativa,
			&sp.Status,
			&sp.UsuarioID,
			&sp.CriadoEm,
			&sp.AtualizadoEm,
			&totalCount,
		)
		if err != nil {
			return nil, 0, err
		}
		ssp = append(ssp, &sp)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return ssp, totalCount, nil
}

func (s *Store) UpdateSolicitacaoPrioridade(ctx context.Context, sp *SolicitacaoPrioridade) error {
	q := `
	UPDATE solicitacoes_prioridade SET
		status = $2,
		justificativa = $3,
		atualizado_em = CURRENT_TIMESTAMP
	WHERE id = $1
	RETURNING atualizado_em`
	args := []any{sp.ID, sp.Status, sp.Justificativa}

	err := s.db.QueryRow(ctx, q, args...).Scan(&sp.AtualizadoEm)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteSolicitacaoPrioridade(ctx context.Context, spID int64) error {
	q := `DELETE FROM solicitacoes_prioridade WHERE id = $1`
	_, err := s.db.Exec(ctx, q, spID)
	if err != nil {
		return err
	}
	return nil
}

package database

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Processo struct {
	ID                  uuid.UUID
	Numero              string
	Gatilho             string
	StatusProcessamento string
	LinkAcesso          string
	SeiUnidadeID        string
	SeiUnidadeSigla     string
	CriadoEm            time.Time
	AtualizadoEm        time.Time
}

func (s *Store) SaveProcesso(ctx context.Context, p *Processo) error {
	q := `
	INSERT INTO processos (numero, gatilho, status_processamento, link_acesso, sei_unidade_id, sei_unidade_sigla)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, criado_em, atualizado_em`
	args := []any{p.Numero, p.Gatilho, p.StatusProcessamento, p.LinkAcesso, p.SeiUnidadeID, p.SeiUnidadeSigla}

	err := s.db.QueryRow(ctx, q, args...).Scan(&p.ID, &p.CriadoEm, &p.AtualizadoEm)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetProcesso(ctx context.Context, id uuid.UUID) (*Processo, error) {
	q := `
	SELECT 
		id, numero, gatilho, status_processamento, link_acesso,
		sei_unidade_id, sei_unidade_sigla, criado_em, atualizado_em
	FROM processos
	WHERE id = $1`

	var p Processo
	err := s.db.QueryRow(ctx, q, id).Scan(
		&p.ID,
		&p.Numero,
		&p.Gatilho,
		&p.StatusProcessamento,
		&p.LinkAcesso,
		&p.SeiUnidadeID,
		&p.SeiUnidadeSigla,
		&p.CriadoEm,
		&p.AtualizadoEm,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (s *Store) GetProcessoByNumero(ctx context.Context, numero string) (*Processo, error) {
	q := `
	SELECT 
		id, numero, gatilho, status_processamento, link_acesso,
		sei_unidade_id, sei_unidade_sigla, criado_em, atualizado_em
	FROM processos
	WHERE numero = $1`

	var p Processo
	err := s.db.QueryRow(ctx, q, numero).Scan(
		&p.ID,
		&p.Numero,
		&p.Gatilho,
		&p.StatusProcessamento,
		&p.LinkAcesso,
		&p.SeiUnidadeID,
		&p.SeiUnidadeSigla,
		&p.CriadoEm,
		&p.AtualizadoEm,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (s *Store) UpdateProcesso(ctx context.Context, p *Processo) error {
	q := `
	UPDATE processos SET
		status_processamento = $2,
		atualizado_em = CURRENT_TIMESTAMP
	WHERE id = $1
	RETURNING atualizado_em`
	args := []any{p.ID, p.StatusProcessamento}

	err := s.db.QueryRow(ctx, q, args...).Scan(&p.AtualizadoEm)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteProcesso(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM processos WHERE id = $1`
	_, err := s.db.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	return nil
}

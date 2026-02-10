package database

import (
	"context"
	"database/sql"
	"encoding/json"
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
	MetadadosIA         json.RawMessage
	AnalisadoEm         sql.Null[time.Time]
	CriadoEm            time.Time
	AtualizadoEm        time.Time
}

func (s *Store) SaveProcesso(ctx context.Context, p *Processo) error {
	q := `
	INSERT INTO processos (numero, gatilho, status_processamento, link_acesso, sei_unidade_id, sei_unidade_sigla)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, metadados_ia, criado_em, atualizado_em`
	args := []any{
		p.Numero,
		p.Gatilho,
		p.StatusProcessamento,
		p.LinkAcesso,
		p.SeiUnidadeID,
		p.SeiUnidadeSigla,
	}

	err := s.db.QueryRow(ctx, q, args...).Scan(
		&p.ID,
		&p.MetadadosIA,
		&p.CriadoEm,
		&p.AtualizadoEm,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetProcesso(ctx context.Context, id uuid.UUID) (*Processo, error) {
	q := `
	SELECT 
		id, numero, gatilho, status_processamento, link_acesso,
		sei_unidade_id, sei_unidade_sigla, metadados_ia, analisado_em, criado_em,
		atualizado_em
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
		&p.MetadadosIA,
		&p.AnalisadoEm,
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
		sei_unidade_id, sei_unidade_sigla, metadados_ia, analisado_em, criado_em,
		atualizado_em
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
		&p.MetadadosIA,
		&p.AnalisadoEm,
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
		metadados_ia = $3,
		analisado_em = $4,
		atualizado_em = CURRENT_TIMESTAMP
	WHERE id = $1
	RETURNING analisado_em, atualizado_em`
	args := []any{p.ID, p.StatusProcessamento, p.MetadadosIA, p.AnalisadoEm}

	err := s.db.QueryRow(ctx, q, args...).Scan(
		&p.AnalisadoEm,
		&p.AtualizadoEm,
	)
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

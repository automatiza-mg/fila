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
	StatusProcessamento string
	LinkAcesso          string
	SeiUnidadeID        string
	SeiUnidadeSigla     string
	MetadadosIA         json.RawMessage
	Aposentadoria       sql.Null[bool]
	AnalisadoEm         sql.Null[time.Time]
	CriadoEm            time.Time
	AtualizadoEm        time.Time
}

func (s *Store) SaveProcesso(ctx context.Context, p *Processo) error {
	q := `
	INSERT INTO processos (numero, status_processamento, link_acesso, sei_unidade_id, sei_unidade_sigla)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, metadados_ia, criado_em, atualizado_em`
	args := []any{
		p.Numero,
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

type ListProcessosParams struct {
	Numero string
	Limit  int
	Offset int
}

func (s *Store) ListProcessos(ctx context.Context, params ListProcessosParams) ([]*Processo, int, error) {
	q := `
	SELECT 
		id, numero, status_processamento, link_acesso, sei_unidade_id,
		sei_unidade_sigla, metadados_ia, aposentadoria, analisado_em, criado_em,
		atualizado_em, COUNT(*) OVER()
	FROM processos
	WHERE (numero LIKE '%' || $1 || '%' OR $1 = '')
	ORDER BY criado_em DESC
	LIMIT $2 OFFSET $3`
	args := []any{params.Numero, params.Limit, params.Offset}

	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	totalCount := 0
	pp := make([]*Processo, 0)

	for rows.Next() {
		var p Processo
		err := rows.Scan(
			&p.ID,
			&p.Numero,
			&p.StatusProcessamento,
			&p.LinkAcesso,
			&p.SeiUnidadeID,
			&p.SeiUnidadeSigla,
			&p.MetadadosIA,
			&p.Aposentadoria,
			&p.AnalisadoEm,
			&p.CriadoEm,
			&p.AtualizadoEm,
			&totalCount,
		)
		if err != nil {
			return nil, 0, err
		}
		pp = append(pp, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return pp, totalCount, nil
}

// GetProcessosMap retorna um mapa de ID -> Processos para os ids informados.
func (s *Store) GetProcessosMap(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*Processo, error) {
	q := `
	SELECT 
		id, numero, status_processamento, link_acesso, sei_unidade_id,
		sei_unidade_sigla, metadados_ia, aposentadoria, analisado_em, criado_em,
		atualizado_em
	FROM processos
	WHERE id = ANY($1)`

	processoMap := make(map[uuid.UUID]*Processo, len(ids))

	rows, err := s.db.Query(ctx, q, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p Processo
		err := rows.Scan(
			&p.ID,
			&p.Numero,
			&p.StatusProcessamento,
			&p.LinkAcesso,
			&p.SeiUnidadeID,
			&p.SeiUnidadeSigla,
			&p.MetadadosIA,
			&p.Aposentadoria,
			&p.AnalisadoEm,
			&p.CriadoEm,
			&p.AtualizadoEm,
		)
		if err != nil {
			return nil, err
		}

		processoMap[p.ID] = &p
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return processoMap, nil
}

func (s *Store) GetProcesso(ctx context.Context, id uuid.UUID) (*Processo, error) {
	q := `
	SELECT 
		id, numero, status_processamento, link_acesso, sei_unidade_id,
		sei_unidade_sigla, metadados_ia, aposentadoria, analisado_em, criado_em,
		atualizado_em
	FROM processos
	WHERE id = $1`

	var p Processo
	err := s.db.QueryRow(ctx, q, id).Scan(
		&p.ID,
		&p.Numero,
		&p.StatusProcessamento,
		&p.LinkAcesso,
		&p.SeiUnidadeID,
		&p.SeiUnidadeSigla,
		&p.MetadadosIA,
		&p.Aposentadoria,
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
		id, numero, status_processamento, link_acesso, sei_unidade_id,
		sei_unidade_sigla, metadados_ia, aposentadoria, analisado_em, criado_em,
		atualizado_em
	FROM processos
	WHERE numero = $1`

	var p Processo
	err := s.db.QueryRow(ctx, q, numero).Scan(
		&p.ID,
		&p.Numero,
		&p.StatusProcessamento,
		&p.LinkAcesso,
		&p.SeiUnidadeID,
		&p.SeiUnidadeSigla,
		&p.MetadadosIA,
		&p.Aposentadoria,
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
		aposentadoria = $4,
		analisado_em = $5,
		atualizado_em = CURRENT_TIMESTAMP
	WHERE id = $1
	RETURNING analisado_em, atualizado_em`
	args := []any{
		p.ID,
		p.StatusProcessamento,
		p.MetadadosIA,
		p.Aposentadoria,
		p.AnalisadoEm,
	}

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

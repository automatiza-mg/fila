package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	StatusProcessoAnalisePendente   StatusProcesso = "ANALISE_PENDENTE"
	StatusProcessoEmAnalise         StatusProcesso = "EM_ANALISE"
	StatusProcessoEmDiligencia      StatusProcesso = "EM_DILIGENCIA"
	StatusProcessoRetornoDiligencia StatusProcesso = "RETORNO_DILIGENCIA"
	StatusProcessoConcluido         StatusProcesso = "CONCLUIDO"
	StatusProcessoLeituraInvalid    StatusProcesso = "LEITURA_INVALIDA"
)

type StatusProcesso string

type ProcessoAposentadoria struct {
	ID                       int64
	ProcessoID               uuid.UUID
	DataRequerimento         time.Time
	CPFRequerente            string
	DataNascimentoRequerente time.Time
	Invalidez                bool
	Judicial                 bool
	Prioridade               bool
	Score                    int
	Status                   StatusProcesso
	AnalistaID               sql.Null[int64]
	UltimoAnalistaID         sql.Null[int64]
	CriadoEm                 time.Time
	AtualizadoEm             time.Time
}

func (s *Store) SaveProcessoAposentadoria(ctx context.Context, pa *ProcessoAposentadoria) error {
	q := `
	INSERT INTO processos_aposentadoria (
		processo_id,
		data_requerimento,
		cpf_requerente,
		data_nascimento_requerente,
		invalidez,
		judicial,
		prioridade,
		score,
		status,
		analista_id,
		ultimo_analista_id
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	ON CONFLICT DO NOTHING
	RETURNING id, data_requerimento, data_nascimento_requerente, criado_em, atualizado_em`
	args := []any{
		pa.ProcessoID,
		pa.DataRequerimento,
		pa.CPFRequerente,
		pa.DataNascimentoRequerente,
		pa.Invalidez,
		pa.Judicial,
		pa.Prioridade,
		pa.Score,
		pa.Status,
		pa.AnalistaID,
		pa.UltimoAnalistaID,
	}

	err := s.db.QueryRow(ctx, q, args...).Scan(
		&pa.ID,
		&pa.DataRequerimento,
		&pa.DataNascimentoRequerente,
		&pa.CriadoEm,
		&pa.AtualizadoEm,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetProcessoAposentadoria(ctx context.Context, id int64) (*ProcessoAposentadoria, error) {
	q := `
	SELECT
		id, processo_id, data_requerimento, cpf_requerente, data_nascimento_requerente,
		invalidez, judicial, prioridade, score, status,
		analista_id, ultimo_analista_id, criado_em, atualizado_em
	FROM processos_aposentadoria
	WHERE id = $1`

	var pa ProcessoAposentadoria
	err := s.db.QueryRow(ctx, q, id).Scan(
		&pa.ID, &pa.ProcessoID, &pa.DataRequerimento, &pa.CPFRequerente, &pa.DataNascimentoRequerente,
		&pa.Invalidez, &pa.Judicial, &pa.Prioridade, &pa.Score, &pa.Status,
		&pa.AnalistaID, &pa.UltimoAnalistaID, &pa.CriadoEm, &pa.AtualizadoEm,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &pa, nil
}

func (s *Store) GetProcessoAposentadoriaByNumero(ctx context.Context, numero string) (*ProcessoAposentadoria, error) {
	q := `
	SELECT
		pa.id, pa.processo_id, pa.data_requerimento, pa.cpf_requerente, pa.data_nascimento_requerente,
		pa.invalidez, pa.judicial, pa.prioridade, pa.score, pa.status,
		pa.analista_id, pa.ultimo_analista_id, pa.criado_em, pa.atualizado_em
	FROM processos_aposentadoria pa
	INNER JOIN processos p ON pa.processo_id = p.id
	WHERE p.numero = $1`

	var pa ProcessoAposentadoria
	err := s.db.QueryRow(ctx, q, numero).Scan(
		&pa.ID, &pa.ProcessoID, &pa.DataRequerimento, &pa.CPFRequerente, &pa.DataNascimentoRequerente,
		&pa.Invalidez, &pa.Judicial, &pa.Prioridade, &pa.Score, &pa.Status,
		&pa.AnalistaID, &pa.UltimoAnalistaID, &pa.CriadoEm, &pa.AtualizadoEm,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &pa, nil
}

func (s *Store) UpdateProcessoAposentadoria(ctx context.Context, pa *ProcessoAposentadoria) error {
	q := `
	UPDATE processos_aposentadoria SET
		analista_id = $2,
		ultimo_analista_id = $3,
		score = $4,
		status = $5,
		atualizado_em = CURRENT_TIMESTAMP
	WHERE id = $1
	RETURNING atualizado_em`
	args := []any{pa.ID, pa.AnalistaID, pa.UltimoAnalistaID, pa.Score, pa.Status}

	err := s.db.QueryRow(ctx, q, args...).Scan(&pa.AtualizadoEm)
	if err != nil {
		return err
	}
	return nil
}

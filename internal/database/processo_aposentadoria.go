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
	ID                       int64           `db:"id"`
	ProcessoID               uuid.UUID       `db:"processo_id"`
	DataRequerimento         time.Time       `db:"data_requerimento"`
	CPFRequerente            string          `db:"cpf_requerente"`
	DataNascimentoRequerente time.Time       `db:"data_nascimento_requerente"`
	Invalidez                bool            `db:"invalidez"`
	Judicial                 bool            `db:"judicial"`
	Prioridade               bool            `db:"prioridade"`
	Score                    int             `db:"score"`
	Status                   StatusProcesso  `db:"status"`
	AnalistaID               sql.Null[int64] `db:"analista_id"`
	UltimoAnalistaID         sql.Null[int64] `db:"ultimo_analista_id"`
	CriadoEm                 time.Time       `db:"criado_em"`
	AtualizadoEm             time.Time       `db:"atualizado_em"`
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

	rows, err := s.db.Query(ctx, q, id)
	if err != nil {
		return nil, err
	}
	pa, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[ProcessoAposentadoria])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return pa, nil
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

	rows, err := s.db.Query(ctx, q, numero)
	if err != nil {
		return nil, err
	}
	pa, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[ProcessoAposentadoria])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return pa, nil
}

func (s *Store) UpdateProcessoAposentadoria(ctx context.Context, pa *ProcessoAposentadoria) error {
	q := `
	UPDATE processos_aposentadoria SET
		analista_id = $2,
		ultimo_analista_id = $3,
		score = $4,
		status = $5,
		prioridade = $6,
		atualizado_em = CURRENT_TIMESTAMP
	WHERE id = $1
	RETURNING atualizado_em`
	args := []any{
		pa.ID,
		pa.AnalistaID,
		pa.UltimoAnalistaID,
		pa.Score,
		pa.Status,
		pa.Prioridade,
	}

	err := s.db.QueryRow(ctx, q, args...).Scan(&pa.AtualizadoEm)
	if err != nil {
		return err
	}
	return nil
}

type ListProcessoAposentadoriaParams struct {
	Numero string
	Status string
	Limit  int
	Offset int
}

// ListProcessoAposentadoria retorna uma lista paginada de processos de aposentadoria.
// Permite filtrar por Numero (do processo) e Status.
func (s *Store) ListProcessoAposentadoria(ctx context.Context, params ListProcessoAposentadoriaParams) ([]*ProcessoAposentadoria, int, error) {
	q := `
	SELECT
		pa.id, pa.processo_id, pa.data_requerimento, pa.cpf_requerente,
		pa.data_nascimento_requerente, pa.invalidez, pa.judicial, pa.prioridade,
		pa.score, pa.status, pa.analista_id, pa.ultimo_analista_id,
		pa.criado_em, pa.atualizado_em, COUNT(*) OVER()
	FROM processos_aposentadoria pa
	INNER JOIN processos p ON pa.processo_id = p.id
	WHERE (LOWER(pa.status::text) = LOWER($1) OR $1 = '')
	  AND (p.numero LIKE '%' || $2 || '%' OR $2 = '')
	ORDER BY pa.criado_em DESC
	LIMIT $3 OFFSET $4`
	args := []any{params.Status, params.Numero, params.Limit, params.Offset}

	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	totalCount := 0
	paa := make([]*ProcessoAposentadoria, 0)

	for rows.Next() {
		var pa ProcessoAposentadoria
		err := rows.Scan(
			&pa.ID, &pa.ProcessoID, &pa.DataRequerimento, &pa.CPFRequerente,
			&pa.DataNascimentoRequerente, &pa.Invalidez, &pa.Judicial, &pa.Prioridade,
			&pa.Score, &pa.Status, &pa.AnalistaID, &pa.UltimoAnalistaID,
			&pa.CriadoEm, &pa.AtualizadoEm, &totalCount,
		)
		if err != nil {
			return nil, 0, err
		}
		paa = append(paa, &pa)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return paa, totalCount, nil
}

func (s *Store) GetProcessoPrioriatario(ctx context.Context, analistaID int64) (*ProcessoAposentadoria, error) {
	q := `
	SELECT 
		pa.id, pa.processo_id, pa.data_requerimento, pa.cpf_requerente,
		pa.data_nascimento_requerente, pa.invalidez, pa.judicial, pa.prioridade,
		pa.score, pa.status, pa.analista_id, pa.ultimo_analista_id,
		pa.criado_em, pa.atualizado_em
	FROM processos_aposentadoria pa
	WHERE pa.status IN ('RETORNO_DILIGENCIA', 'ANALISE_PENDENTE')
	ORDER BY
		CASE
			WHEN pa.status = 'RETORNO_DILIGENCIA' AND pa.ultimo_analista_id = $1 THEN 1
			WHEN pa.status = 'RETORNO_DILIGENCIA' AND pa.ultimo_analista_id IS NULL THEN 2
			WHEN pa.status = 'ANALISE_PENDENTE' THEN 3
			ELSE 4
		END,
		pa.score DESC,
		pa.data_requerimento ASC
	LIMIT 1
	FOR UPDATE SKIP LOCKED`

	var pa ProcessoAposentadoria
	err := s.db.QueryRow(ctx, q, analistaID).Scan(
		&pa.ID, &pa.ProcessoID, &pa.DataRequerimento, &pa.CPFRequerente,
		&pa.DataNascimentoRequerente, &pa.Invalidez, &pa.Judicial, &pa.Prioridade,
		&pa.Score, &pa.Status, &pa.AnalistaID, &pa.UltimoAnalistaID,
		&pa.CriadoEm, &pa.AtualizadoEm,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &pa, nil
}

func (s *Store) GetProcessoAtribuido(ctx context.Context, analistaID int64) (*ProcessoAposentadoria, error) {
	q := `
	SELECT
		pa.id, pa.processo_id, pa.data_requerimento, pa.cpf_requerente,
		pa.data_nascimento_requerente, pa.invalidez, pa.judicial, pa.prioridade,
		pa.score, pa.status, pa.analista_id, pa.ultimo_analista_id,
		pa.criado_em, pa.atualizado_em
	FROM processos_aposentadoria pa
	WHERE pa.status = 'EM_ANALISE'
	AND analista_id = $1`

	var pa ProcessoAposentadoria
	err := s.db.QueryRow(ctx, q, analistaID).Scan(
		&pa.ID, &pa.ProcessoID, &pa.DataRequerimento, &pa.CPFRequerente,
		&pa.DataNascimentoRequerente, &pa.Invalidez, &pa.Judicial, &pa.Prioridade,
		&pa.Score, &pa.Status, &pa.AnalistaID, &pa.UltimoAnalistaID,
		&pa.CriadoEm, &pa.AtualizadoEm,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &pa, nil
}

// ListAllProcessoAposentadoria retorna todos os processos de aposentadoria.
func (s *Store) ListAllProcessoAposentadoria(ctx context.Context) ([]*ProcessoAposentadoria, error) {
	q := `
	SELECT
		id, processo_id, data_requerimento, cpf_requerente, data_nascimento_requerente,
		invalidez, judicial, prioridade, score, status,
		analista_id, ultimo_analista_id, criado_em, atualizado_em
	FROM processos_aposentadoria
	ORDER BY id`

	rows, err := s.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	paa := make([]*ProcessoAposentadoria, 0)
	for rows.Next() {
		var pa ProcessoAposentadoria
		err := rows.Scan(
			&pa.ID, &pa.ProcessoID, &pa.DataRequerimento, &pa.CPFRequerente,
			&pa.DataNascimentoRequerente, &pa.Invalidez, &pa.Judicial, &pa.Prioridade,
			&pa.Score, &pa.Status, &pa.AnalistaID, &pa.UltimoAnalistaID,
			&pa.CriadoEm, &pa.AtualizadoEm,
		)
		if err != nil {
			return nil, err
		}
		paa = append(paa, &pa)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return paa, nil
}

// GetNumeroProcessoAposentadoria retorna no número do processo SEI para um
// determinado processo de aposentadoria
func (s *Store) GetNumeroProcessoAposentadoria(ctx context.Context, paID int64) (string, error) {
	q := `
	SELECT
		p.numero
	FROM processos_aposentadoria pa
	JOIN processos p ON pa.processo_id = p.id
	WHERE pa.id = $1`

	var numero string
	err := s.db.QueryRow(ctx, q, paID).Scan(&numero)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return "", ErrNotFound
		default:
			return "", err
		}
	}
	return numero, nil
}

// HasProcessoAposentadoria reporta se determinado CPF se encontra no nosso
// banco de dados.
func (s *Store) HasProcessoAposentadoria(ctx context.Context, cpf string) (bool, error) {
	q := `SELECT COUNT(*) FROM processos_aposentadoria WHERE cpf_requerente = $1`

	var count int
	err := s.db.QueryRow(ctx, q, cpf).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

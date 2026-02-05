package database

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrAnalistaExists = errors.New("analista already exists")
)

type Analista struct {
	UsuarioID          int64      `json:"-"`
	Orgao              string     `json:"orgao"`
	SEIUnidadeID       string     `json:"sei_unidade_id"`
	SEIUnidadeSigla    string     `json:"sei_unidade_sigla"`
	Afastado           bool       `json:"afastado"`
	UltimaAtribuicaoEm *time.Time `json:"ultima_atribuicao_em"`
}

// SaveAnalista insere os dados de analista vinculado a um usuário no banco de dados.
func (s *Store) SaveAnalista(ctx context.Context, analista *Analista) error {
	q := `
	INSERT INTO analistas (usuario_id, orgao, sei_unidade_id, sei_unidade_sigla, afastado, ultima_atribuicao_em)
	VALUES ($1, $2, $3, $4, $5, $6)`

	args := []any{
		analista.UsuarioID,
		analista.Orgao,
		analista.SEIUnidadeID,
		analista.SEIUnidadeSigla,
		analista.Afastado,
		analista.UltimaAtribuicaoEm,
	}

	_, err := s.db.Exec(ctx, q, args...)
	if err != nil {
		if strings.Contains(err.Error(), "analistas_pkey") {
			return ErrAnalistaExists
		}
		return err
	}
	return nil
}

// GetAnalista retorna os dados de um analista pelo usuarioID. Retorna [ErrNotFound] caso não seja encontrado.
func (s *Store) GetAnalista(ctx context.Context, usuarioID int64) (*Analista, error) {
	q := `
	SELECT 
		usuario_id, orgao, sei_unidade_id, sei_unidade_sigla,
		afastado, ultima_atribuicao_em
	FROM analistas
	WHERE usuario_id = $1`

	var analista Analista
	err := s.db.QueryRow(ctx, q, usuarioID).Scan(
		&analista.UsuarioID,
		&analista.Orgao,
		&analista.SEIUnidadeID,
		&analista.SEIUnidadeSigla,
		&analista.Afastado,
		&analista.UltimaAtribuicaoEm,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &analista, nil
}

func (s *Store) GetAnalistasByUsuarioIDs(ctx context.Context, ids []int64) (map[int64]*Analista, error) {
	q := `
	SELECT
		usuario_id, orgao, sei_unidade_id, sei_unidade_sigla,
		afastado, ultima_atribuicao_em
	FROM analistas
	WHERE usuario_id = ANY($1)`

	rows, err := s.db.Query(ctx, q, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	analistas := make(map[int64]*Analista)
	for rows.Next() {
		var analista Analista
		err := rows.Scan(
			&analista.UsuarioID,
			&analista.Orgao,
			&analista.SEIUnidadeID,
			&analista.SEIUnidadeSigla,
			&analista.Afastado,
			&analista.UltimaAtribuicaoEm,
		)
		if err != nil {
			return nil, err
		}
		analistas[analista.UsuarioID] = &analista
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return analistas, nil
}

func (s *Store) UpdateAnalista(ctx context.Context, analista *Analista) error {
	q := `
	UPDATE analistas SET
		orgao = $2,
		sei_unidade_id = $3,
		sei_unidade_sigla = $4,
		afastado = $5,
		ultima_atribuicao_em = $6
	WHERE usuario_id = $1
	RETURNING ultima_atribuicao_em`

	args := []any{
		analista.UsuarioID,
		analista.Orgao,
		analista.SEIUnidadeID,
		analista.SEIUnidadeSigla,
		analista.Afastado,
		analista.UltimaAtribuicaoEm,
	}

	err := s.db.QueryRow(ctx, q, args...).Scan(&analista.UltimaAtribuicaoEm)
	if err != nil {
		return err
	}
	return nil
}

// DeleteAnalista exclui os dados de analista de um usuário.
func (s *Store) DeleteAnalista(ctx context.Context, usuarioID int64) error {
	q := `DELETE FROM analistas WHERE usuario_id = $1`
	_, err := s.db.Exec(ctx, q, usuarioID)
	if err != nil {
		return err
	}
	return nil
}

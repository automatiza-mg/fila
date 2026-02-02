package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type Analista struct {
	UsuarioID          int64
	Orgao              string
	SEIUnidadeID       string
	Afastado           bool
	UltimaAtribuicaoEm sql.Null[time.Time]
}

// SaveAnalista insere os dados de analista vinculado a um usuário no banco de dados.
func (s *Store) SaveAnalista(ctx context.Context, analista *Analista) error {
	q := `
	INSERT INTO analistas (usuario_id, orgao, sei_unidade_id, afastado, ultima_atribuicao_em)
	VALUES ($1, $2, $3, $4, $5)`

	args := []any{
		analista.UsuarioID,
		analista.Orgao,
		analista.SEIUnidadeID,
		analista.Afastado,
		analista.UltimaAtribuicaoEm,
	}

	_, err := s.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	return nil
}

// GetAnalista retorna os dados de um analista pelo usuarioID. Retorna [ErrNotFound] caso não seja encontrado.
func (s *Store) GetAnalista(ctx context.Context, usuarioID int64) (*Analista, error) {
	q := `
	SELECT 
		usuario_id, orgao, sei_unidade_id, afastado, ultima_atribuicao_em
	FROM analistas
	WHERE usuario_id = $1`

	var analista Analista
	err := s.db.QueryRow(ctx, q, usuarioID).Scan(
		&analista.UsuarioID,
		&analista.Orgao,
		&analista.SEIUnidadeID,
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

func (s *Store) UpdateAnalista(ctx context.Context, analista *Analista) error {
	q := `
	UPDATE analistas SET
		orgao = $2,
		sei_unidade_id = $3,
		afastado = $4,
		ultima_atribuicao_em = $5
	WHERE usuario_id = $1
	RETURNING ultima_atribuicao_em`

	args := []any{
		analista.UsuarioID,
		analista.Orgao,
		analista.SEIUnidadeID,
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

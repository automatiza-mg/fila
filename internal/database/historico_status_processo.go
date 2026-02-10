package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type HistoricoStatusProcesso struct {
	ID                      int64
	ProcessoAposentadoriaID int64
	StatusAnterior          sql.Null[StatusProcesso]
	StatusNovo              StatusProcesso
	UsuarioID               sql.Null[int64]
	Observacao              sql.Null[string]
	AlteradoEm              time.Time
}

func (s *Store) SaveHistoricoStatusProcesso(ctx context.Context, h *HistoricoStatusProcesso) error {
	q := `
	INSERT INTO historico_status_processo (
		processo_aposentadoria_id,
		status_anterior,
		status_novo,
		usuario_id,
		observacao
	)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, alterado_em`
	args := []any{h.ProcessoAposentadoriaID, h.StatusAnterior, h.StatusNovo, h.UsuarioID, h.Observacao}

	err := s.db.QueryRow(ctx, q, args...).Scan(&h.ID, &h.AlteradoEm)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetHistoricoStatusProcesso(ctx context.Context, id int64) (*HistoricoStatusProcesso, error) {
	q := `
	SELECT
		id, processo_aposentadoria_id, status_anterior, status_novo, usuario_id,
		observacao, alterado_em
	FROM historico_status_processo
	WHERE id = $1`

	var h HistoricoStatusProcesso
	err := s.db.QueryRow(ctx, q, id).Scan(
		&h.ID,
		&h.ProcessoAposentadoriaID,
		&h.StatusAnterior,
		&h.StatusNovo,
		&h.UsuarioID,
		&h.Observacao,
		&h.AlteradoEm,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &h, nil
}

func (s *Store) ListHistoricoStatusProcesso(ctx context.Context, paID int64) ([]*HistoricoStatusProcesso, error) {
	q := `
	SELECT
		id, processo_aposentadoria_id, status_anterior, status_novo, usuario_id,
		observacao, alterado_em
	FROM historico_status_processo
	WHERE processo_aposentadoria_id = $1`

	rows, err := s.db.Query(ctx, q, paID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hh := make([]*HistoricoStatusProcesso, 0)
	for rows.Next() {
		var h HistoricoStatusProcesso
		err := rows.Scan(
			&h.ID,
			&h.ProcessoAposentadoriaID,
			&h.StatusAnterior,
			&h.StatusNovo,
			&h.UsuarioID,
			&h.Observacao,
			&h.AlteradoEm,
		)
		if err != nil {
			return nil, err
		}
		hh = append(hh, &h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return hh, nil
}

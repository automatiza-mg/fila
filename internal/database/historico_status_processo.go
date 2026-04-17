package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type HistoricoStatusProcesso struct {
	ID                      int64                    `db:"id"`
	ProcessoAposentadoriaID int64                    `db:"processo_aposentadoria_id"`
	StatusAnterior          sql.Null[StatusProcesso] `db:"status_anterior"`
	StatusNovo              StatusProcesso           `db:"status_novo"`
	UsuarioID               sql.Null[int64]          `db:"usuario_id"`
	Observacao              sql.Null[string]         `db:"observacao"`
	AlteradoEm              time.Time                `db:"alterado_em"`
}

func (h *HistoricoStatusProcesso) SetObservacao(obs string) {
	h.Observacao = sql.Null[string]{
		V:     obs,
		Valid: obs != "",
	}
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

	rows, err := s.db.Query(ctx, q, id)
	if err != nil {
		return nil, err
	}
	h, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[HistoricoStatusProcesso])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return h, nil
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
	return pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[HistoricoStatusProcesso])
}

package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type SolicitacaoDiligencia struct {
	ID                      int64     `db:"id"`
	ProcessoAposentadoriaID int64     `db:"processo_aposentadoria_id"`
	AnalistaID              int64     `db:"analista_id"`
	CriadoEm                time.Time `db:"criado_em"`
}

type ItemDiligencia struct {
	ID                      int64    `db:"id"`
	SolicitacaoDiligenciaID int64    `db:"solicitacao_diligencia_id"`
	Tipo                    string   `db:"tipo"`
	Subcategorias           []string `db:"subcategorias"`
	Detalhe                 string   `db:"detalhe"`
}

// SaveSolicitacaoDiligencia insere uma nova solicitação de diligência.
func (s *Store) SaveSolicitacaoDiligencia(ctx context.Context, sd *SolicitacaoDiligencia) error {
	q := `
	INSERT INTO solicitacoes_diligencia (processo_aposentadoria_id, analista_id)
	VALUES ($1, $2)
	RETURNING id, criado_em`
	args := []any{sd.ProcessoAposentadoriaID, sd.AnalistaID}

	err := s.db.QueryRow(ctx, q, args...).Scan(&sd.ID, &sd.CriadoEm)
	if err != nil {
		return err
	}
	return nil
}

// GetSolicitacaoDiligencia retorna uma solicitação pelo ID. Retorna
// [ErrNotFound] caso não exista.
func (s *Store) GetSolicitacaoDiligencia(ctx context.Context, id int64) (*SolicitacaoDiligencia, error) {
	q := `
	SELECT
		id, processo_aposentadoria_id, analista_id, criado_em
	FROM solicitacoes_diligencia
	WHERE id = $1`

	rows, err := s.db.Query(ctx, q, id)
	if err != nil {
		return nil, err
	}
	sd, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[SolicitacaoDiligencia])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return sd, nil
}

// ListSolicitacoesDiligenciaByProcesso retorna as solicitações de diligência
// de um processo de aposentadoria, ordenadas da mais recente para a mais antiga.
func (s *Store) ListSolicitacoesDiligenciaByProcesso(ctx context.Context, paID int64) ([]*SolicitacaoDiligencia, error) {
	q := `
	SELECT
		id, processo_aposentadoria_id, analista_id, criado_em
	FROM solicitacoes_diligencia
	WHERE processo_aposentadoria_id = $1
	ORDER BY criado_em DESC, id DESC`

	rows, err := s.db.Query(ctx, q, paID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[SolicitacaoDiligencia])
}

// DeleteSolicitacaoDiligencia exclui uma solicitação de diligência.
func (s *Store) DeleteSolicitacaoDiligencia(ctx context.Context, id int64) error {
	q := `DELETE FROM solicitacoes_diligencia WHERE id = $1`
	_, err := s.db.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	return nil
}

// SaveItemDiligencia insere um novo item em uma solicitação de diligência.
func (s *Store) SaveItemDiligencia(ctx context.Context, item *ItemDiligencia) error {
	q := `
	INSERT INTO itens_diligencia (solicitacao_diligencia_id, tipo, subcategorias, detalhe)
	VALUES ($1, $2, $3, $4)
	RETURNING id`
	args := []any{item.SolicitacaoDiligenciaID, item.Tipo, item.Subcategorias, item.Detalhe}

	err := s.db.QueryRow(ctx, q, args...).Scan(&item.ID)
	if err != nil {
		return err
	}
	return nil
}

// ListItensDiligencia retorna os itens de uma solicitação de diligência
// ordenados pelo ID crescente (ordem de inserção).
func (s *Store) ListItensDiligencia(ctx context.Context, solicitacaoID int64) ([]*ItemDiligencia, error) {
	q := `
	SELECT
		id, solicitacao_diligencia_id, tipo, subcategorias, detalhe
	FROM itens_diligencia
	WHERE solicitacao_diligencia_id = $1
	ORDER BY id ASC`

	rows, err := s.db.Query(ctx, q, solicitacaoID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ItemDiligencia])
}

// DeleteItensDiligencia remove todos os itens associados a uma solicitação.
func (s *Store) DeleteItensDiligencia(ctx context.Context, solicitacaoID int64) error {
	q := `DELETE FROM itens_diligencia WHERE solicitacao_diligencia_id = $1`
	_, err := s.db.Exec(ctx, q, solicitacaoID)
	if err != nil {
		return err
	}
	return nil
}

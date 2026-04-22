package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// StatusSolicitacaoDiligencia representa o estado de uma solicitação de diligência.
type StatusSolicitacaoDiligencia string

const (
	// StatusSolicitacaoRascunho indica uma solicitação em construção pelo analista.
	StatusSolicitacaoRascunho StatusSolicitacaoDiligencia = "rascunho"
	// StatusSolicitacaoEnviada indica uma solicitação finalizada e enviada.
	StatusSolicitacaoEnviada StatusSolicitacaoDiligencia = "enviada"
)

// SolicitacaoDiligencia representa um lote de diligências de um analista para
// um processo de aposentadoria. Pode estar em rascunho ou já enviada.
type SolicitacaoDiligencia struct {
	ID                      int64                       `db:"id"`
	ProcessoAposentadoriaID int64                       `db:"processo_aposentadoria_id"`
	AnalistaID              int64                       `db:"analista_id"`
	Status                  StatusSolicitacaoDiligencia `db:"status"`
	CriadoEm                time.Time                   `db:"criado_em"`
	EnviadaEm               sql.Null[time.Time]         `db:"enviada_em"`
}

// ItemDiligencia representa uma diligência individual dentro de uma solicitação.
type ItemDiligencia struct {
	ID                      int64    `db:"id"`
	SolicitacaoDiligenciaID int64    `db:"solicitacao_diligencia_id"`
	Tipo                    string   `db:"tipo"`
	Subcategorias           []string `db:"subcategorias"`
	Detalhe                 string   `db:"detalhe"`
}

// SaveSolicitacaoDiligencia insere uma nova solicitação de diligência. Caso
// Status seja o zero value, o default da coluna (rascunho) é aplicado.
func (s *Store) SaveSolicitacaoDiligencia(ctx context.Context, sd *SolicitacaoDiligencia) error {
	var status any
	if sd.Status == "" {
		status = nil
	} else {
		status = sd.Status
	}

	q := `
	INSERT INTO solicitacoes_diligencia (processo_aposentadoria_id, analista_id, status, enviada_em)
	VALUES ($1, $2, COALESCE($3, 'rascunho'::status_solicitacao_diligencia), $4)
	RETURNING id, status, criado_em, enviada_em`
	args := []any{sd.ProcessoAposentadoriaID, sd.AnalistaID, status, sd.EnviadaEm}

	err := s.db.QueryRow(ctx, q, args...).Scan(&sd.ID, &sd.Status, &sd.CriadoEm, &sd.EnviadaEm)
	if err != nil {
		return err
	}
	return nil
}

// UpdateSolicitacaoDiligencia atualiza os campos mutáveis (status, enviada_em)
// de uma solicitação de diligência.
func (s *Store) UpdateSolicitacaoDiligencia(ctx context.Context, sd *SolicitacaoDiligencia) error {
	q := `
	UPDATE solicitacoes_diligencia SET
		status = $2,
		enviada_em = $3
	WHERE id = $1`
	args := []any{sd.ID, sd.Status, sd.EnviadaEm}

	_, err := s.db.Exec(ctx, q, args...)
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
		id, processo_aposentadoria_id, analista_id, status, criado_em, enviada_em
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

// GetRascunhoDiligencia retorna o rascunho ativo de um analista para um processo
// de aposentadoria. Retorna [ErrNotFound] caso não exista.
func (s *Store) GetRascunhoDiligencia(ctx context.Context, paID, analistaID int64) (*SolicitacaoDiligencia, error) {
	q := `
	SELECT
		id, processo_aposentadoria_id, analista_id, status, criado_em, enviada_em
	FROM solicitacoes_diligencia
	WHERE processo_aposentadoria_id = $1
	  AND analista_id = $2
	  AND status = 'rascunho'`

	rows, err := s.db.Query(ctx, q, paID, analistaID)
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

// ListSolicitacoesDiligenciaParams define os filtros aceitos por
// [Store.ListSolicitacoesDiligenciaByProcesso].
type ListSolicitacoesDiligenciaParams struct {
	ProcessoAposentadoriaID int64
	Status                  StatusSolicitacaoDiligencia
}

// ListSolicitacoesDiligenciaByProcesso retorna as solicitações de diligência
// de um processo de aposentadoria, ordenadas da mais recente para a mais antiga.
// Se Status for vazio, retorna todas.
func (s *Store) ListSolicitacoesDiligenciaByProcesso(ctx context.Context, params ListSolicitacoesDiligenciaParams) ([]*SolicitacaoDiligencia, error) {
	q := `
	SELECT
		id, processo_aposentadoria_id, analista_id, status, criado_em, enviada_em
	FROM solicitacoes_diligencia
	WHERE processo_aposentadoria_id = $1
	  AND (status::text = $2 OR $2 = '')
	ORDER BY criado_em DESC, id DESC`

	rows, err := s.db.Query(ctx, q, params.ProcessoAposentadoriaID, string(params.Status))
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[SolicitacaoDiligencia])
}

// DeleteSolicitacaoDiligencia exclui uma solicitação de diligência. Os itens
// associados são removidos em cascata pela FK.
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

package diligencias

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/jackc/pgx/v5/pgconn"
)

// SolicitacaoDiligencia representa um lote de diligências solicitadas por um
// analista para um processo de aposentadoria, podendo estar em rascunho ou
// já enviada.
type SolicitacaoDiligencia struct {
	ID                      int64             `json:"id"`
	ProcessoAposentadoriaID int64             `json:"processo_aposentadoria_id"`
	AnalistaID              int64             `json:"analista_id"`
	Status                  string            `json:"status"`
	Itens                   []*ItemDiligencia `json:"itens"`
	CriadoEm                time.Time         `json:"criado_em"`
	EnviadaEm               *time.Time        `json:"enviada_em"`
}

// ItemDiligencia representa uma diligência individual dentro de uma solicitação.
type ItemDiligencia struct {
	ID            int64    `json:"id"`
	Tipo          string   `json:"tipo"`
	Subcategorias []string `json:"subcategorias"`
	Detalhe       string   `json:"detalhe"`
}

// NovoItem representa os dados de entrada para criação de um item de diligência.
type NovoItem struct {
	Tipo          string
	Subcategorias []string
	Detalhe       string
}

func mapItem(it *database.ItemDiligencia) *ItemDiligencia {
	return &ItemDiligencia{
		ID:            it.ID,
		Tipo:          it.Tipo,
		Subcategorias: it.Subcategorias,
		Detalhe:       it.Detalhe,
	}
}

func mapSolicitacao(sd *database.SolicitacaoDiligencia, itens []*database.ItemDiligencia) *SolicitacaoDiligencia {
	items := make([]*ItemDiligencia, len(itens))
	for i, it := range itens {
		items[i] = mapItem(it)
	}
	return &SolicitacaoDiligencia{
		ID:                      sd.ID,
		ProcessoAposentadoriaID: sd.ProcessoAposentadoriaID,
		AnalistaID:              sd.AnalistaID,
		Status:                  string(sd.Status),
		Itens:                   items,
		CriadoEm:                sd.CriadoEm,
		EnviadaEm:               database.Ptr(sd.EnviadaEm),
	}
}

// GetOrCreateRascunho retorna o rascunho ativo de diligência para o analista
// em um processo de aposentadoria. Cria um novo rascunho caso não exista.
// Retorna [ErrNotAssigned] se o processo não estiver atribuído ao analista e
// [ErrInvalidStatus] se o processo não estiver em análise.
func (s *Service) GetOrCreateRascunho(ctx context.Context, paID, analistaID int64) (*SolicitacaoDiligencia, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

	pa, err := store.GetProcessoAposentadoria(ctx, paID)
	if err != nil {
		return nil, err
	}
	if !pa.AnalistaID.Valid || pa.AnalistaID.V != analistaID {
		return nil, ErrNotAssigned
	}
	if pa.Status != database.StatusProcessoEmAnalise {
		return nil, ErrInvalidStatus
	}

	sd, err := store.GetRascunhoDiligencia(ctx, paID, analistaID)
	switch {
	case err == nil:
		// rascunho existente
	case errors.Is(err, database.ErrNotFound):
		sd = &database.SolicitacaoDiligencia{
			ProcessoAposentadoriaID: paID,
			AnalistaID:              analistaID,
		}
		if err := store.SaveSolicitacaoDiligencia(ctx, sd); err != nil {
			if isUniqueViolation(err) {
				sd, err = store.GetRascunhoDiligencia(ctx, paID, analistaID)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
	default:
		return nil, err
	}

	itens, err := store.ListItensDiligencia(ctx, sd.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return mapSolicitacao(sd, itens), nil
}

// GetRascunho retorna o rascunho ativo de um analista para um processo de
// aposentadoria, sem criar um novo caso não exista. Retorna
// [database.ErrNotFound] quando não há rascunho.
func (s *Service) GetRascunho(ctx context.Context, paID, analistaID int64) (*SolicitacaoDiligencia, error) {
	sd, err := s.store.GetRascunhoDiligencia(ctx, paID, analistaID)
	if err != nil {
		return nil, err
	}

	itens, err := s.store.ListItensDiligencia(ctx, sd.ID)
	if err != nil {
		return nil, err
	}

	return mapSolicitacao(sd, itens), nil
}

type SalvarRascunhoParams struct {
	SolicitacaoID int64
	AnalistaID    int64
	Itens         []NovoItem
}

// SalvarRascunho substitui o conjunto de itens de um rascunho pelos itens
// informados. Itens anteriores são descartados. Retorna [ErrNotAssigned] se a
// solicitação não pertencer ao analista informado, [ErrAlreadySent] se a
// solicitação já tiver sido enviada e [ErrInvalidStatus] se o processo não
// estiver em análise.
func (s *Service) SalvarRascunho(ctx context.Context, params SalvarRascunhoParams) (*SolicitacaoDiligencia, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

	sd, err := store.GetSolicitacaoDiligencia(ctx, params.SolicitacaoID)
	if err != nil {
		return nil, err
	}
	if sd.AnalistaID != params.AnalistaID {
		return nil, ErrNotAssigned
	}
	if sd.Status != database.StatusSolicitacaoRascunho {
		return nil, ErrAlreadySent
	}

	pa, err := store.GetProcessoAposentadoria(ctx, sd.ProcessoAposentadoriaID)
	if err != nil {
		return nil, err
	}
	if pa.Status != database.StatusProcessoEmAnalise {
		return nil, ErrInvalidStatus
	}
	if !pa.AnalistaID.Valid || pa.AnalistaID.V != params.AnalistaID {
		return nil, ErrNotAssigned
	}

	if err := store.DeleteItensDiligencia(ctx, sd.ID); err != nil {
		return nil, err
	}
	for _, it := range params.Itens {
		subs := it.Subcategorias
		if subs == nil {
			subs = []string{}
		}
		item := &database.ItemDiligencia{
			SolicitacaoDiligenciaID: sd.ID,
			Tipo:                    it.Tipo,
			Subcategorias:           subs,
			Detalhe:                 it.Detalhe,
		}
		if err := store.SaveItemDiligencia(ctx, item); err != nil {
			return nil, err
		}
	}

	itens, err := store.ListItensDiligencia(ctx, sd.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return mapSolicitacao(sd, itens), nil
}

// DescartarRascunho exclui um rascunho de diligência. Retorna [ErrNotAssigned]
// se a solicitação não pertencer ao analista e [ErrAlreadySent] se já tiver
// sido enviada.
func (s *Service) DescartarRascunho(ctx context.Context, solicitacaoID, analistaID int64) error {
	sd, err := s.store.GetSolicitacaoDiligencia(ctx, solicitacaoID)
	if err != nil {
		return err
	}
	if sd.AnalistaID != analistaID {
		return ErrNotAssigned
	}
	if sd.Status != database.StatusSolicitacaoRascunho {
		return ErrAlreadySent
	}

	return s.store.DeleteSolicitacaoDiligencia(ctx, solicitacaoID)
}

// EnviarDiligencia finaliza um rascunho, marcando-o como enviado, alterando o
// status do processo para EM_DILIGENCIA e desatribuindo o analista. Retorna
// [ErrNotAssigned] se a solicitação ou o processo não pertencer ao analista,
// [ErrAlreadySent] se a solicitação já tiver sido enviada, [ErrInvalidStatus]
// se o processo não estiver em análise e [ErrDraftEmpty] se o rascunho não
// possuir itens.
func (s *Service) EnviarDiligencia(ctx context.Context, solicitacaoID, analistaID int64) (*SolicitacaoDiligencia, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

	sd, err := store.GetSolicitacaoDiligencia(ctx, solicitacaoID)
	if err != nil {
		return nil, err
	}
	if sd.AnalistaID != analistaID {
		return nil, ErrNotAssigned
	}
	if sd.Status != database.StatusSolicitacaoRascunho {
		return nil, ErrAlreadySent
	}

	pa, err := store.GetProcessoAposentadoria(ctx, sd.ProcessoAposentadoriaID)
	if err != nil {
		return nil, err
	}
	if pa.Status != database.StatusProcessoEmAnalise {
		return nil, ErrInvalidStatus
	}
	if !pa.AnalistaID.Valid || pa.AnalistaID.V != analistaID {
		return nil, ErrNotAssigned
	}

	itens, err := store.ListItensDiligencia(ctx, sd.ID)
	if err != nil {
		return nil, err
	}
	if len(itens) == 0 {
		return nil, ErrDraftEmpty
	}

	now := time.Now().UTC()
	sd.Status = database.StatusSolicitacaoEnviada
	sd.EnviadaEm = sql.Null[time.Time]{V: now, Valid: true}
	if err := store.UpdateSolicitacaoDiligencia(ctx, sd); err != nil {
		return nil, err
	}

	statusAnterior := pa.Status
	pa.UltimoAnalistaID = pa.AnalistaID
	pa.AnalistaID = sql.Null[int64]{}
	pa.Status = database.StatusProcessoEmDiligencia
	if err := store.UpdateProcessoAposentadoria(ctx, pa); err != nil {
		return nil, err
	}

	if err := s.saveHistorico(ctx, store, saveHistoricoParams{
		ProcessoAposentadoriaID: pa.ID,
		StatusAnterior:          &statusAnterior,
		StatusNovo:              database.StatusProcessoEmDiligencia,
		UsuarioID:               &analistaID,
		Observacao:              "Diligência solicitada",
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return mapSolicitacao(sd, itens), nil
}

// GetSolicitacaoDiligencia retorna uma solicitação de diligência pelo ID,
// incluindo seus itens.
func (s *Service) GetSolicitacaoDiligencia(ctx context.Context, id int64) (*SolicitacaoDiligencia, error) {
	sd, err := s.store.GetSolicitacaoDiligencia(ctx, id)
	if err != nil {
		return nil, err
	}

	itens, err := s.store.ListItensDiligencia(ctx, sd.ID)
	if err != nil {
		return nil, err
	}

	return mapSolicitacao(sd, itens), nil
}

// ListSolicitacoesEnviadas retorna todas as solicitações de diligência enviadas
// para um processo de aposentadoria, ordenadas da mais recente para a mais
// antiga, com seus itens carregados.
func (s *Service) ListSolicitacoesEnviadas(ctx context.Context, paID int64) ([]*SolicitacaoDiligencia, error) {
	ss, err := s.store.ListSolicitacoesDiligenciaByProcesso(ctx, database.ListSolicitacoesDiligenciaParams{
		ProcessoAposentadoriaID: paID,
		Status:                  database.StatusSolicitacaoEnviada,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*SolicitacaoDiligencia, 0, len(ss))
	for _, sd := range ss {
		itens, err := s.store.ListItensDiligencia(ctx, sd.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, mapSolicitacao(sd, itens))
	}

	return result, nil
}

// isUniqueViolation retorna true se o erro for uma violação de constraint
// UNIQUE do Postgres (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

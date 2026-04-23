package fila

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/pagination"
	"github.com/google/uuid"
)

var (
	// ErrNotAssigned é o erro retornado quando o processo não está atribuído ao analista.
	ErrNotAssigned = errors.New("processo não está atribuído ao analista")
	// ErrInvalidStatus é o erro retornado quando o processo não está no status esperado.
	ErrInvalidStatus = errors.New("processo não está no status esperado para esta ação")
)

// Processo é um processo de aposentadoria processado pelo sistema.
type ProcessoAposentadoria struct {
	ID                       int64     `json:"id"`
	ProcessoID               uuid.UUID `json:"processo_id"`
	Numero                   string    `json:"numero"`
	DataRequerimento         time.Time `json:"data_requerimento"`
	CPFRequerente            string    `json:"cpf_requerente"`
	DataNascimentoRequerente time.Time `json:"data_nascimento_requerente"`
	Invalidez                bool      `json:"invalidez"`
	Judicial                 bool      `json:"judicial"`
	Prioridade               bool      `json:"prioridade"`
	Score                    int       `json:"score"`
	Status                   string    `json:"status"`
	Analista                 *string   `json:"analista"`
	AnalistaID               *int64    `json:"analista_id"`
	PossuiPreview            bool      `json:"possui_preview"`
	Alertas                  []string  `json:"alertas"`
	CriadoEm                 time.Time `json:"criado_em"`
	AtualizadoEm             time.Time `json:"atualizado_em"`
}

func mapProcesso(pa *database.ProcessoAposentadoria, p *database.Processo, analista *string) *ProcessoAposentadoria {
	return &ProcessoAposentadoria{
		ID:                       pa.ID,
		ProcessoID:               pa.ProcessoID,
		Numero:                   p.Numero,
		DataRequerimento:         pa.DataRequerimento,
		CPFRequerente:            pa.CPFRequerente,
		DataNascimentoRequerente: pa.DataNascimentoRequerente,
		Invalidez:                pa.Invalidez,
		Judicial:                 pa.Judicial,
		Prioridade:               pa.Prioridade,
		Score:                    pa.Score,
		Status:                   string(pa.Status),
		Analista:                 analista,
		AnalistaID:               database.Ptr(pa.AnalistaID),
		PossuiPreview:            p.PreviewHash.Valid,
		Alertas:                  pa.Alertas,
		CriadoEm:                 pa.CriadoEm,
		AtualizadoEm:             pa.AtualizadoEm,
	}
}

// GetProcessoAposentadoria retorna um processo de aposentadoria pelo ID.
func (s *Service) GetProcessoAposentadoria(ctx context.Context, id int64) (*ProcessoAposentadoria, error) {
	pa, err := s.store.GetProcessoAposentadoria(ctx, id)
	if err != nil {
		return nil, err
	}

	p, err := s.store.GetProcesso(ctx, pa.ProcessoID)
	if err != nil {
		return nil, err
	}

	var analista *string
	if pa.AnalistaID.Valid {
		nome, err := s.store.GetNomeAnalista(ctx, pa.AnalistaID.V)
		if err != nil {
			return nil, err
		}
		analista = &nome
	}

	return mapProcesso(pa, p, analista), nil
}

// GetProcessoAposentadoriaByNumero retorna um processo de aposentadoria pelo
// número do processo SEI.
func (s *Service) GetProcessoAposentadoriaByNumero(ctx context.Context, numero string) (*ProcessoAposentadoria, error) {
	p, err := s.store.GetProcessoByNumero(ctx, numero)
	if err != nil {
		return nil, err
	}

	pa, err := s.store.GetProcessoAposentadoriaByNumero(ctx, numero)
	if err != nil {
		return nil, err
	}

	var analista *string
	if pa.AnalistaID.Valid {
		nome, err := s.store.GetNomeAnalista(ctx, pa.AnalistaID.V)
		if err != nil {
			return nil, err
		}
		analista = &nome
	}

	return mapProcesso(pa, p, analista), nil
}

type ListProcessoAposentadoriaParams struct {
	Numero string
	Status string
	Page   int
	Limit  int
}

// ListProcesso retorna a lista paginada dos processos de aposentadoria com seus numeros.
func (s *Service) ListProcesso(ctx context.Context, params ListProcessoAposentadoriaParams) (*pagination.Result[*ProcessoAposentadoria], error) {
	offset := pagination.Offset(params.Page, params.Limit)

	paa, totalCount, err := s.store.ListProcessoAposentadoria(ctx, database.ListProcessoAposentadoriaParams{
		Numero: params.Numero,
		Status: params.Status,
		Limit:  params.Limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	if len(paa) == 0 {
		return pagination.NewResult[*ProcessoAposentadoria](nil, params.Page, 0, params.Limit), nil
	}

	processos := make([]*ProcessoAposentadoria, 0, len(paa))

	// Busca os números dos processos.
	for _, pa := range paa {
		p, err := s.store.GetProcesso(ctx, pa.ProcessoID)
		if err != nil {
			return nil, err
		}

		var analista *string
		if pa.AnalistaID.Valid {
			nome, err := s.store.GetNomeAnalista(ctx, pa.AnalistaID.V)
			if err != nil {
				return nil, err
			}
			analista = &nome
		}

		processos = append(processos, mapProcesso(pa, p, analista))
	}

	return pagination.NewResult(processos, params.Page, totalCount, params.Limit), nil
}

// ListHistoricoAnalista retorna os processos em que o analista foi o último responsável,
// nos status CONCLUIDO, LEITURA_INVALIDA ou EM_DILIGENCIA. Suporta busca por número e paginação.
func (s *Service) ListHistoricoAnalista(ctx context.Context, analistaID int64, numero string, page, limit int) (*pagination.Result[*ProcessoAposentadoria], error) {
	offset := pagination.Offset(page, limit)

	paa, totalCount, err := s.store.ListProcessoAposentadoria(ctx, database.ListProcessoAposentadoriaParams{
		Numero:           numero,
		UltimoAnalistaID: sql.Null[int64]{V: analistaID, Valid: true},
		StatusIn: []database.StatusProcesso{
			database.StatusProcessoConcluido,
			database.StatusProcessoLeituraInvalid,
			database.StatusProcessoEmDiligencia,
		},
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	if len(paa) == 0 {
		return pagination.NewResult[*ProcessoAposentadoria](nil, page, 0, limit), nil
	}

	processos := make([]*ProcessoAposentadoria, 0, len(paa))
	for _, pa := range paa {
		p, err := s.store.GetProcesso(ctx, pa.ProcessoID)
		if err != nil {
			return nil, err
		}

		var analista *string
		if pa.AnalistaID.Valid {
			nome, err := s.store.GetNomeAnalista(ctx, pa.AnalistaID.V)
			if err != nil {
				return nil, err
			}
			analista = &nome
		}

		processos = append(processos, mapProcesso(pa, p, analista))
	}

	return pagination.NewResult(processos, page, totalCount, limit), nil
}

type MarcarLeituraInvalidaParams struct {
	AnalistaID int64
	ProcessoID int64
	Motivo     string
}

// MarcarLeituraInvalida marca um processo de aposentadoria como leitura inválida,
// desatribuindo o analista. Retorna [ErrNotAssigned] caso o processo não esteja
// atribuído ao analista informado e [ErrInvalidStatus] caso o processo não esteja em análise.
func (s *Service) MarcarLeituraInvalida(ctx context.Context, params MarcarLeituraInvalidaParams) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

	pa, err := store.GetProcessoAposentadoria(ctx, params.ProcessoID)
	if err != nil {
		return err
	}

	if !pa.AnalistaID.Valid || pa.AnalistaID.V != params.AnalistaID {
		return ErrNotAssigned
	}

	if pa.Status != database.StatusProcessoEmAnalise {
		return ErrInvalidStatus
	}

	if err := s.saveHistorico(ctx, store, saveHistoricoParams{
		ProcessoAposentadoriaID: pa.ID,
		StatusAnterior:          &pa.Status,
		StatusNovo:              database.StatusProcessoLeituraInvalid,
		UsuarioID:               &params.AnalistaID,
		Observacao:              params.Motivo,
	}); err != nil {
		return err
	}

	pa.Status = database.StatusProcessoLeituraInvalid
	pa.UltimoAnalistaID = pa.AnalistaID
	pa.AnalistaID = sql.Null[int64]{}

	if err := store.UpdateProcessoAposentadoria(ctx, pa); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// RegistrarPublicacao marca um processo de aposentadoria como concluído,
// desatribuindo o analista. Retorna [ErrNotAssigned] caso o processo não esteja
// atribuído ao analista informado e [ErrInvalidStatus] caso o processo não esteja em análise.
func (s *Service) RegistrarPublicacao(ctx context.Context, paID, analistaID int64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

	pa, err := store.GetProcessoAposentadoria(ctx, paID)
	if err != nil {
		return err
	}

	if !pa.AnalistaID.Valid || pa.AnalistaID.V != analistaID {
		return ErrNotAssigned
	}

	if pa.Status != database.StatusProcessoEmAnalise {
		return ErrInvalidStatus
	}

	if err := s.saveHistorico(ctx, store, saveHistoricoParams{
		ProcessoAposentadoriaID: pa.ID,
		StatusAnterior:          &pa.Status,
		StatusNovo:              database.StatusProcessoConcluido,
		UsuarioID:               &analistaID,
	}); err != nil {
		return err
	}

	pa.Status = database.StatusProcessoConcluido
	pa.UltimoAnalistaID = pa.AnalistaID
	pa.AnalistaID = sql.Null[int64]{}

	if err := store.UpdateProcessoAposentadoria(ctx, pa); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetProcessoAtribuido retorna o processo de aposentadoria atribuído a um analista.
// Retorna [database.ErrNotFound] se o analista não tiver um processo EM_ANALISE.
func (s *Service) GetProcessoAtribuido(ctx context.Context, analistaID int64) (*ProcessoAposentadoria, error) {
	pa, err := s.store.GetProcessoAtribuido(ctx, analistaID)
	if err != nil {
		return nil, err
	}

	p, err := s.store.GetProcesso(ctx, pa.ProcessoID)
	if err != nil {
		return nil, err
	}

	var analista *string
	if pa.AnalistaID.Valid {
		nome, err := s.store.GetNomeAnalista(ctx, pa.AnalistaID.V)
		if err != nil {
			return nil, err
		}
		analista = &nome
	}

	return mapProcesso(pa, p, analista), nil
}

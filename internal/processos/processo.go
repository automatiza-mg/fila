package processos

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/pagination"
	"github.com/google/uuid"
)

var (
	ErrProcessoExists = errors.New("processo already exists")
)

type Processo struct {
	ID              uuid.UUID       `json:"id"`
	Numero          string          `json:"numero"`
	Status          string          `json:"status"`
	LinkAcesso      string          `json:"link_acesso"`
	SeiUnidadeID    string          `json:"sei_unidade_id"`
	SeiUnidadeSigla string          `json:"sei_unidade_sigla"`
	Aposentadoria   *bool           `json:"aposentadoria"`
	AnalisadoEm     *time.Time      `json:"analisado_em"`
	MetadadosIA     json.RawMessage `json:"metadados_ia"`
	CriadoEm        time.Time       `json:"criado_em"`
	AtualizadoEm    time.Time       `json:"atualizado_em"`
}

func mapProcesso(p *database.Processo) *Processo {
	return &Processo{
		ID:              p.ID,
		Numero:          p.Numero,
		Status:          p.StatusProcessamento,
		LinkAcesso:      p.LinkAcesso,
		SeiUnidadeID:    p.SeiUnidadeID,
		SeiUnidadeSigla: p.SeiUnidadeSigla,
		Aposentadoria:   database.Ptr(p.Aposentadoria),
		AnalisadoEm:     database.Ptr(p.AnalisadoEm),
		MetadadosIA:     p.MetadadosIA,
		CriadoEm:        p.CriadoEm,
		AtualizadoEm:    p.AtualizadoEm,
	}
}

// CreateProcesso cria um novo processo no banco de dados, colocando a análise
// na fila de processamento automaticamente.
func (s *Service) CreateProcesso(ctx context.Context, numeroProcesso string) (*Processo, error) {
	resp, err := s.sei.ConsultarProcedimento(ctx, numeroProcesso)
	if err != nil {
		return nil, err
	}

	unidade := resp.Parametros.AndamentoGeracao.Unidade
	linkAcesso := resp.Parametros.LinkAcesso

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

	p := &database.Processo{
		Numero:              numeroProcesso,
		StatusProcessamento: "PENDENTE",
		LinkAcesso:          linkAcesso,
		SeiUnidadeID:        unidade.IdUnidade,
		SeiUnidadeSigla:     unidade.Sigla,
	}

	err = store.SaveProcesso(ctx, p)
	if err != nil {
		if strings.Contains(err.Error(), "processos_numero_key") {
			return nil, ErrProcessoExists
		}
		return nil, err
	}

	if _, err := s.queue.EnqueueAnalyzeTx(ctx, tx, p.ID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return mapProcesso(p), nil
}

// GetProcessoByNumero retorna os dados de um processo pelo numero (protocolo).
func (s *Service) GetProcessoByNumero(ctx context.Context, numeroProcesso string) (*Processo, error) {
	p, err := s.store.GetProcessoByNumero(ctx, numeroProcesso)
	if err != nil {
		return nil, err
	}
	return mapProcesso(p), nil
}

// GetProcesso retorna os dados de um processo pelo ID.
func (s *Service) GetProcesso(ctx context.Context, processoID uuid.UUID) (*Processo, error) {
	p, err := s.store.GetProcesso(ctx, processoID)
	if err != nil {
		return nil, err
	}
	return mapProcesso(p), nil
}

type ListProcessosParams struct {
	Numero string
	Page   int
	Limit  int
}

// ListProcessos retorna a lista paginada dos processos analisados pela aplicação.
func (s *Service) ListProcessos(ctx context.Context, params ListProcessosParams) (*pagination.Result[*Processo], error) {
	offset := pagination.Offset(params.Page, params.Limit)

	pp, totalCount, err := s.store.ListProcessos(ctx, database.ListProcessosParams{
		Numero: params.Numero,
		Limit:  params.Limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	processos := make([]*Processo, len(pp))
	for i, p := range pp {
		processos[i] = mapProcesso(p)
	}

	return pagination.NewResult(processos, params.Page, totalCount, params.Limit), nil
}

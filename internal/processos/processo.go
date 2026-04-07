package processos

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/pagination"
	"github.com/automatiza-mg/fila/internal/tasks"
	"github.com/google/uuid"
	"github.com/riverqueue/river"
)

var (
	ErrProcessoExists     = errors.New("processo already exists")
	ErrPreviewUnavailable = errors.New("preview não disponível para este processo")
)

type Processo struct {
	ID              uuid.UUID       `json:"id"`
	Numero          string          `json:"numero"`
	Status          string          `json:"status"`
	Resumo          string          `json:"resumo"`
	LinkAcesso      string          `json:"link_acesso"`
	SeiUnidadeID    string          `json:"sei_unidade_id"`
	SeiUnidadeSigla string          `json:"sei_unidade_sigla"`
	Aposentadoria   *bool           `json:"aposentadoria"`
	PreviewHash     *string         `json:"preview_hash"`
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
		Resumo:          p.Resumo,
		LinkAcesso:      p.LinkAcesso,
		SeiUnidadeID:    p.SeiUnidadeID,
		SeiUnidadeSigla: p.SeiUnidadeSigla,
		Aposentadoria:   database.Ptr(p.Aposentadoria),
		PreviewHash:     database.Ptr(p.PreviewHash),
		AnalisadoEm:     database.Ptr(p.AnalisadoEm),
		MetadadosIA:     p.MetadadosIA,
		CriadoEm:        p.CriadoEm,
		AtualizadoEm:    p.AtualizadoEm,
	}
}

// CreateProcesso cria um novo processo no banco de dados, colocando a análise
// na fila de processamento automaticamente.
func (s *Service) CreateProcesso(ctx context.Context, num string) (*Processo, error) {
	resp, err := s.sei.ConsultarProcedimento(ctx, num)
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
		Numero:              num,
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

	_, err = s.queue.InsertManyTx(ctx, tx, []river.InsertManyParams{
		{Args: tasks.DownloadPreviewArgs{ProcessoID: p.ID}},
		{Args: tasks.DownloadProcessoArgs{ProcessoID: p.ID}},
	})

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

// Preview retorna um conteúdo do preview associado ao processo.
type Preview struct {
	Body        io.ReadCloser
	ContentType string
}

// GetPreview retorna o PDF de preview de um processo.
func (s *Service) GetPreview(ctx context.Context, processoID uuid.UUID) (*Preview, error) {
	p, err := s.store.GetProcesso(ctx, processoID)
	if err != nil {
		return nil, err
	}

	if !p.PreviewHash.Valid {
		return nil, ErrPreviewUnavailable
	}

	arq, err := s.store.GetArquivo(ctx, p.PreviewHash.V)
	if err != nil {
		return nil, err
	}

	body, err := s.storage.Get(ctx, arq.ChaveStorage)
	if err != nil {
		return nil, err
	}

	return &Preview{
		Body:        body,
		ContentType: arq.ContentType,
	}, nil
}

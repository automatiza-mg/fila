package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/automatiza-mg/fila/internal/aposentadoria"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/datalake"
	"github.com/automatiza-mg/fila/internal/llm"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

const (
	// UnidadeRecebimento é a sigla da unidade responsável por receber os
	// processos de aposentadoria.
	UnidadeRecebimento = "SEPLAG/DCCTA"
)

type AnalisarProcessoArgs struct {
	ProcessoID uuid.UUID `json:"processo_id"`
}

func (args AnalisarProcessoArgs) Kind() string {
	return "processo:analisar"
}

// DataRecebimentoFetcher busca a data de recebimento de um processo.
type DataRecebimentoFetcher interface {
	GetDataRecebimento(ctx context.Context, numero, unidade string) (time.Time, error)
}

// ServidorFetcher busca os dados de um servidor pelo CPF.
type ServidorFetcher interface {
	GetServidor(ctx context.Context, cpf string) (*datalake.Servidor, error)
}

type AnalisarProcessoWorker struct {
	pool            *pgxpool.Pool
	store           *database.Store
	llm             *llm.Client
	dataFetcher     DataRecebimentoFetcher
	servidorFetcher ServidorFetcher
	logger          *slog.Logger
	river.WorkerDefaults[AnalisarProcessoArgs]
}

func (w *AnalisarProcessoWorker) Work(ctx context.Context, job *river.Job[AnalisarProcessoArgs]) error {
	p, err := w.store.GetProcesso(ctx, job.Args.ProcessoID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return river.JobCancel(err)
		}
		return fmt.Errorf("failed to get processo: %w", err)
	}

	dd, err := w.store.ListDocumentos(ctx, p.ID)
	if err != nil {
		return fmt.Errorf("failed to list docs: %w", err)
	}

	hashes := make([]string, 0, len(dd))
	for _, d := range dd {
		hashes = append(hashes, d.ArquivoHash)
	}

	arquivoMap, err := w.store.GetArquivosMap(ctx, hashes)
	if err != nil {
		return fmt.Errorf("failed to load arquivos: %w", err)
	}

	docs, err := mapDocumentos(dd, arquivoMap)
	if err != nil {
		return err
	}

	analise, err := w.llm.AnalisarAposentadoria(ctx, docs)
	if err != nil {
		return fmt.Errorf("failed to run analyses: %w", err)
	}

	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start tx: %w", err)
	}
	defer tx.Rollback(ctx)

	store := w.store.WithTx(tx)

	p.SetAposentadoria(analise.Aposentadoria)
	p.SetAnalisadoEm()
	p.MetadadosIA, err = json.Marshal(analise)
	if err != nil {
		return fmt.Errorf("failed to marshal analise: %w", err)
	}

	// Atualiza e retorna.
	if !analise.Aposentadoria {
		return store.UpdateProcesso(ctx, p)
	}

	dataNascimento, err := time.Parse(time.DateOnly, analise.DataNascimento)
	if err != nil {
		return err
	}

	invalidez := analise.Invalidez

	// Enriquece os dados do processo com informações do servidor no datalake.
	servidor, err := w.servidorFetcher.GetServidor(ctx, analise.CPF)
	if err != nil {
		w.logger.Warn("falha ao buscar servidor no datalake, usando dados da IA",
			slog.String("cpf", analise.CPF),
			slog.String("erro", err.Error()),
		)
	} else {
		dataNascimento = servidor.DataNascimento
		invalidez = invalidez || servidor.PossuiDeficiencia
	}

	// Busca a informação complementar da data de recebimento do processo.
	dataRequerimento, err := w.dataFetcher.GetDataRecebimento(ctx, p.Numero, UnidadeRecebimento)
	if err != nil {
		dataRequerimento, err = time.Parse(time.DateOnly, analise.DataRequerimento)
		if err != nil {
			return err
		}
	}

	score := aposentadoria.CalculateScore(
		dataNascimento,
		invalidez,
		analise.Judicial,
		false,
	)

	pa := &database.ProcessoAposentadoria{
		ProcessoID:               p.ID,
		CPFRequerente:            analise.CPF,
		Invalidez:                invalidez,
		Judicial:                 analise.Judicial,
		DataNascimentoRequerente: dataNascimento,
		DataRequerimento:         dataRequerimento,
		Status:                   database.StatusProcessoAnalisePendente,
		Score:                    score,
	}
	err = store.SaveProcessoAposentadoria(ctx, pa)
	if err != nil {
		return err
	}

	hist := &database.HistoricoStatusProcesso{
		ProcessoAposentadoriaID: pa.ID,
		StatusNovo:              database.StatusProcessoAnalisePendente,
	}
	hist.SetObservacao("Processo criado após análise de IA")

	err = store.SaveHistoricoStatusProcesso(ctx, hist)
	if err != nil {
		return err
	}

	err = store.UpdateProcesso(ctx, p)
	if err != nil {
		return fmt.Errorf("failed to update processo: %w", err)
	}

	return tx.Commit(ctx)
}

func NewAnalisarProcessoWorker(pool *pgxpool.Pool, logger *slog.Logger, llm *llm.Client, dataFetcher DataRecebimentoFetcher, servidorFetcher ServidorFetcher) *AnalisarProcessoWorker {
	return &AnalisarProcessoWorker{
		pool:            pool,
		store:           database.New(pool),
		llm:             llm,
		dataFetcher:     dataFetcher,
		servidorFetcher: servidorFetcher,
		logger:          logger.With(slog.String("worker", "analisar_processo")),
	}
}

// Converte uma lista de documentos do banco de dados para o formato
// esperado pela IA.
func mapDocumentos(dd []*database.Documento, arquivoMap map[string]*database.Arquivo) ([]llm.Documento, error) {
	docs := make([]llm.Documento, 0, len(dd))

	for _, d := range dd {
		var seiDoc sei.RetornoConsultaDocumento
		err := json.Unmarshal(d.MetadadosAPI, &seiDoc)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal api data: %w", err)
		}

		assinaturas := make([]llm.Assinatura, 0, len(seiDoc.Assinaturas.Itens))
		for _, a := range seiDoc.Assinaturas.Itens {
			assinaturas = append(assinaturas, llm.Assinatura{
				Nome: a.Nome,
				CPF:  a.Sigla,
			})
		}

		var conteudo string
		if arq, ok := arquivoMap[d.ArquivoHash]; ok {
			conteudo = arq.Conteudo
		}

		docs = append(docs, llm.Documento{
			Tipo:        d.Tipo,
			Data:        seiDoc.Data,
			Conteudo:    conteudo,
			Assinaturas: assinaturas,
		})
	}

	return docs, nil
}

package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/automatiza-mg/fila/internal/aposentadoria"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/llm"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

type AnalisarProcessoArgs struct {
	ProcessoID uuid.UUID `json:"processo_id"`
}

func (args AnalisarProcessoArgs) Kind() string {
	return "processo:analisar"
}

type AnalisarProcessoWorker struct {
	pool  *pgxpool.Pool
	store *database.Store
	llm   *llm.Client
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

	docs := make([]llm.Documento, 0, len(dd))

	// Mapeia os dados do banco de dados para leitura da IA.
	for _, d := range dd {
		var seiDoc sei.RetornoConsultaDocumento
		err := json.Unmarshal(d.MetadadosAPI, &seiDoc)
		if err != nil {
			return fmt.Errorf("failed to unmarshal api data: %w", err)
		}

		assinaturas := make([]llm.Assinatura, 0, len(seiDoc.Assinaturas.Itens))
		for _, a := range seiDoc.Assinaturas.Itens {
			assinaturas = append(assinaturas, llm.Assinatura{
				Nome: a.Nome,
				CPF:  a.Sigla,
			})
		}

		docs = append(docs, llm.Documento{
			Tipo:        d.Tipo,
			Data:        seiDoc.Data,
			Conteudo:    d.OCR,
			Assinaturas: assinaturas,
		})
	}

	analise, err := w.llm.AnalisarAposentadoriaV2(ctx, docs)
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

	dataRequerimento, err := time.Parse(time.DateOnly, analise.DataRequerimento)
	if err != nil {
		return err
	}

	score := aposentadoria.CalculateScore(dataNascimento, analise.Invalidez)

	pa := &database.ProcessoAposentadoria{
		ProcessoID:               p.ID,
		CPFRequerente:            analise.CPF,
		Invalidez:                analise.Invalidez,
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

	err = store.UpdateProcesso(ctx, p)
	if err != nil {
		return fmt.Errorf("failed to update processo: %w", err)
	}

	return tx.Commit(ctx)
}

func NewAnalisarProcessoWorker(pool *pgxpool.Pool, llm *llm.Client) *AnalisarProcessoWorker {
	return &AnalisarProcessoWorker{
		pool:  pool,
		store: database.New(pool),
		llm:   llm,
	}
}

package fila

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	apos "github.com/automatiza-mg/fila/internal/aposentadoria"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/processos"
	"github.com/jackc/pgx/v5"
)

var _ processos.AnalyzeHook = (*Service)(nil)

// OnAnalyzeCompleteTx implementa [processos.AnalyzeHook].
// Executa a análise de IA sobre os documentos e atualiza o processo com o resultado.
func (s *Service) OnAnalyzeCompleteTx(ctx context.Context, tx pgx.Tx, proc *processos.Processo, dd []*processos.Documento) error {
	return s.analyzeAposentadoriaTx(ctx, tx, proc, dd)
}

func (s *Service) analyzeAposentadoriaTx(ctx context.Context, tx pgx.Tx, proc *processos.Processo, dd []*processos.Documento) error {
	store := s.store.WithTx(tx)

	p, err := store.GetProcesso(ctx, proc.ID)
	if err != nil {
		return err
	}

	_, err = store.GetProcessoAposentadoriaByNumero(ctx, proc.Numero)
	if err == nil {
		// TODO: Verificar se está EM_DILIGENCIA e tomar as devidas providências.
		return nil
	}
	if !errors.Is(err, database.ErrNotFound) {
		return err
	}

	res, err := s.analyzer.AnalisarAposentadoria(ctx, dd)
	if err != nil {
		return err
	}

	metadados, err := json.Marshal(res)
	if err != nil {
		return err
	}

	p.Aposentadoria = sql.Null[bool]{
		Valid: true,
	}
	p.AnalisadoEm = sql.Null[time.Time]{
		Valid: true,
		V:     time.Now(),
	}
	p.MetadadosIA = metadados

	if res.Aposentadoria {
		p.Aposentadoria.V = true

		dataNascimento, err := time.Parse(time.DateOnly, res.DataNascimento)
		if err != nil {
			return err
		}

		dataRequerimento, err := time.Parse(time.DateOnly, res.DataRequerimento)
		if err != nil {
			return err
		}

		score := apos.CalculateScore(dataNascimento, res.Invalidez)

		pa := &database.ProcessoAposentadoria{
			ProcessoID:               p.ID,
			CPFRequerente:            res.CPF,
			Invalidez:                res.Invalidez,
			Judicial:                 res.Judicial,
			DataNascimentoRequerente: dataNascimento,
			DataRequerimento:         dataRequerimento,
			Status:                   database.StatusProcessoAnalisePendente,
			Score:                    score,
		}
		err = store.SaveProcessoAposentadoria(ctx, pa)
		if err != nil {
			return err
		}

		err = store.SaveHistoricoStatusProcesso(ctx, &database.HistoricoStatusProcesso{
			ProcessoAposentadoriaID: pa.ID,
			StatusNovo:              database.StatusProcessoAnalisePendente,
			Observacao: sql.Null[string]{
				Valid: true,
				V:     "Processo criado automaticamente através de análise de IA",
			},
		})
		if err != nil {
			return err
		}
	}

	err = store.UpdateProcesso(ctx, p)
	if err != nil {
		return err
	}

	return nil
}

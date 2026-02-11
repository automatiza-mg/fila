package fila

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/automatiza-mg/fila/internal/aposentadoria"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/processos"
)

var _ processos.AnalyzeHook = (*Service)(nil)

// DocumentAnalyzer analisa documentos de um processo usando IA.
type DocumentAnalyzer interface {
	AnalisarAposentadoria(ctx context.Context, docs []*processos.Documento) (*aposentadoria.Analise, error)
}

// OnAnalyzeComplete implementa [processos.AnalyzeHook].
// Executa a análise de IA sobre os documentos e atualiza o processo com o resultado.
func (s *Service) OnAnalyzeComplete(ctx context.Context, proc *processos.Processo, dd []*processos.Documento) error {
	p, err := s.store.GetProcesso(ctx, proc.ID)
	if err != nil {
		return err
	}

	apos, err := s.analyzer.AnalisarAposentadoria(ctx, dd)
	if err != nil {
		return err
	}

	metadados, err := json.Marshal(apos)
	if err != nil {
		return err
	}

	p.MetadadosIA = metadados
	p.AnalisadoEm = sql.Null[time.Time]{
		V:     time.Now(),
		Valid: true,
	}
	p.Aposentadoria = sql.Null[bool]{
		Valid: true,
	}

	// Processo é de aposentadoria
	if apos.Aposentadoria {
		p.Aposentadoria.V = true

		dataNascimento, err := time.Parse(time.DateOnly, apos.DataNascimento)
		if err != nil {
			return err
		}

		dataRequerimento, err := time.Parse(time.DateOnly, apos.DataRequerimento)
		if err != nil {
			return err
		}

		err = s.store.SaveProcessoAposentadoria(ctx, &database.ProcessoAposentadoria{
			ProcessoID:               p.ID,
			CPFRequerente:            apos.CPF,
			Invalidez:                apos.Invalidez,
			Judicial:                 apos.Judicial,
			DataNascimentoRequerente: dataNascimento,
			DataRequerimento:         dataRequerimento,
			Status:                   database.StatusProcessoAnalisePendente,
		})
		if err != nil {
			return err
		}
	}

	p.StatusProcessamento = "SUCESSO"
	return s.store.UpdateProcesso(ctx, p)
}

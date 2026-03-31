package fila

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/logging"
)

// Atribui um processo de aposentadoria a um analista disponível.
func (s *Service) assignProcessoAposentadoria(ctx context.Context) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("falha ao iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

	analistaID, err := store.GetAnalistaDisponivel(ctx)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			logger := logging.FromContext(ctx)
			if logger != nil {
				logger.Debug("Nenhum analista disponível para atribuição")
			}
			return nil
		}
		return fmt.Errorf("erro ao obter analista disponível: %w", err)
	}

	processo, err := store.GetProcessoPrioriatario(ctx, analistaID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			logger := logging.FromContext(ctx)
			if logger != nil {
				logger.Debug("nenhum processo disponível para atribuição", slog.Int64("analista_id", analistaID))
			}
			return nil
		}
		return fmt.Errorf("erro ao obter processo prioritário: %w", err)
	}

	statusAnterior := processo.Status

	processo.AnalistaID = sql.Null[int64]{Valid: true, V: analistaID}
	processo.Status = database.StatusProcessoEmAnalise

	if err := store.UpdateProcessoAposentadoria(ctx, processo); err != nil {
		return fmt.Errorf("erro ao atualizar processo: %w", err)
	}

	if err := s.saveHistorico(ctx, store, saveHistoricoParams{
		ProcessoAposentadoriaID: processo.ID,
		StatusAnterior:          &statusAnterior,
		StatusNovo:              database.StatusProcessoEmAnalise,
		Observacao:              "Processo atribuído para análise",
	}); err != nil {
		return fmt.Errorf("erro ao salvar histórico: %w", err)
	}

	analista, err := store.GetAnalista(ctx, analistaID)
	if err != nil {
		return fmt.Errorf("erro ao obter dados do analista: %w", err)
	}

	analista.UltimaAtribuicaoEm = sql.Null[time.Time]{Valid: true, V: time.Now()}
	if err := store.UpdateAnalista(ctx, analista); err != nil {
		return fmt.Errorf("erro ao atualizar timestamp do analista: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("erro ao confirmar transação: %w", err)
	}

	return nil
}

// StartAssignmentWorker inicia uma goroutine que atribui processos a analistas periodicamente.
// A goroutine será cancelada quando o contexto for fechado.
func (s *Service) StartAssignmentWorker(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		logger := logging.FromContext(ctx)

		for {
			select {
			case <-ctx.Done():
				logger.Debug("Encerrando worker de atribuição de processos")
				return
			case <-ticker.C:
				if err := s.assignProcessoAposentadoria(ctx); err != nil {
					logger.Error("Erro ao atribuir processo",
						slog.String("error", err.Error()),
					)
				}
			}
		}
	}()
}

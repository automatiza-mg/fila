package fila

import (
	"context"
	"database/sql"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
)

type HistoricoStatusProcesso struct {
	StatusAnterior *string   `json:"status_anterior"`
	StatusNovo     string    `json:"status_novo"`
	UsuarioID      *int64    `json:"usuario_id"`
	Usuario        *string   `json:"usuario"`
	Observacao     *string   `json:"observacao"`
	AlteradoEm     time.Time `json:"alterado_em"`
}

func mapHistoricoStatusProcesso(h *database.HistoricoStatusProcesso, usuario *string) *HistoricoStatusProcesso {
	statusAnterior := (*string)(nil)
	if h.StatusAnterior.Valid {
		s := string(h.StatusAnterior.V)
		statusAnterior = &s
	}

	return &HistoricoStatusProcesso{
		StatusAnterior: statusAnterior,
		StatusNovo:     string(h.StatusNovo),
		UsuarioID:      database.Ptr(h.UsuarioID),
		Usuario:        usuario,
		Observacao:     database.Ptr(h.Observacao),
		AlteradoEm:     h.AlteradoEm,
	}
}

type saveHistoricoParams struct {
	ProcessoAposentadoriaID int64
	StatusAnterior          *database.StatusProcesso
	StatusNovo              database.StatusProcesso
	UsuarioID               *int64
	Observacao              string
}

// saveHistorico registra uma entrada no histórico de status de um processo de aposentadoria.
func (s *Service) saveHistorico(ctx context.Context, store *database.Store, params saveHistoricoParams) error {
	var statusAnterior sql.Null[database.StatusProcesso]
	if params.StatusAnterior != nil {
		statusAnterior = sql.Null[database.StatusProcesso]{V: *params.StatusAnterior, Valid: true}
	}

	return store.SaveHistoricoStatusProcesso(ctx, &database.HistoricoStatusProcesso{
		ProcessoAposentadoriaID: params.ProcessoAposentadoriaID,
		StatusAnterior:          statusAnterior,
		StatusNovo:              params.StatusNovo,
		UsuarioID:               database.Null(params.UsuarioID),
		Observacao:              sql.Null[string]{V: params.Observacao, Valid: params.Observacao != ""},
	})
}

// ListHistorico retorna o histórico completo de mudanças de status de um processo de aposentadoria.
func (s *Service) ListHistorico(ctx context.Context, paID int64) ([]*HistoricoStatusProcesso, error) {
	hh, err := s.store.ListHistoricoStatusProcesso(ctx, paID)
	if err != nil {
		return nil, err
	}

	historico := make([]*HistoricoStatusProcesso, len(hh))
	for i, h := range hh {
		var usuario *string
		if h.UsuarioID.Valid {
			nome, err := s.store.GetNomeAnalista(ctx, h.UsuarioID.V)
			if err == nil {
				usuario = &nome
			}
		}
		historico[i] = mapHistoricoStatusProcesso(h, usuario)
	}

	return historico, nil
}

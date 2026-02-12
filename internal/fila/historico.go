package fila

import (
	"context"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
)

// HistoricoStatusProcesso representa uma mudança de status em um processo de aposentadoria.
type HistoricoStatusProcesso struct {
	ID                      int64     `json:"id"`
	ProcessoAposentadoriaID int64     `json:"processo_aposentadoria_id"`
	StatusAnterior          *string   `json:"status_anterior"`
	StatusNovo              string    `json:"status_novo"`
	UsuarioID               *int64    `json:"usuario_id"`
	Observacao              *string   `json:"observacao"`
	AlteradoEm              time.Time `json:"alterado_em"`
}

func mapHistoricoStatusProcesso(h *database.HistoricoStatusProcesso) *HistoricoStatusProcesso {
	statusAnterior := (*string)(nil)
	if h.StatusAnterior.Valid {
		s := string(h.StatusAnterior.V)
		statusAnterior = &s
	}

	return &HistoricoStatusProcesso{
		ID:                      h.ID,
		ProcessoAposentadoriaID: h.ProcessoAposentadoriaID,
		StatusAnterior:          statusAnterior,
		StatusNovo:              string(h.StatusNovo),
		UsuarioID:               database.Ptr(h.UsuarioID),
		Observacao:              database.Ptr(h.Observacao),
		AlteradoEm:              h.AlteradoEm,
	}
}

// ListHistorico retorna o histórico completo de mudanças de status de um processo de aposentadoria.
func (s *Service) ListHistorico(ctx context.Context, processoAposentadoriaID int64) ([]*HistoricoStatusProcesso, error) {
	hh, err := s.store.ListHistoricoStatusProcesso(ctx, processoAposentadoriaID)
	if err != nil {
		return nil, err
	}

	historico := make([]*HistoricoStatusProcesso, len(hh))
	for i, h := range hh {
		historico[i] = mapHistoricoStatusProcesso(h)
	}

	return historico, nil
}

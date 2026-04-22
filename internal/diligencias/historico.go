package diligencias

import (
	"context"
	"database/sql"

	"github.com/automatiza-mg/fila/internal/database"
)

type saveHistoricoParams struct {
	ProcessoAposentadoriaID int64
	StatusAnterior          *database.StatusProcesso
	StatusNovo              database.StatusProcesso
	UsuarioID               *int64
	Observacao              string
}

// saveHistorico registra uma entrada no histórico de status de um processo
// de aposentadoria.
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

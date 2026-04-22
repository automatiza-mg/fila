// Package diligencias fornece operações de negócio para gerenciamento de
// solicitações de diligência em processos de aposentadoria.
package diligencias

import (
	"errors"
	"log/slog"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotAssigned é retornado quando o processo não está atribuído ao analista.
var ErrNotAssigned = errors.New("process is not assigned to the analyst")

// ErrInvalidStatus é retornado quando o processo não está no status esperado
// para a ação solicitada.
var ErrInvalidStatus = errors.New("process is not in the expected status for this action")

// ErrAlreadySent é retornado quando há uma tentativa de modificar uma
// solicitação de diligência que já foi enviada.
var ErrAlreadySent = errors.New("diligência has already been sent and cannot be modified")

// ErrDraftEmpty é retornado quando uma tentativa de envio é feita em um
// rascunho sem itens.
var ErrDraftEmpty = errors.New("rascunho has no items to send")

// Service gerencia solicitações de diligência em processos de aposentadoria.
type Service struct {
	pool   *pgxpool.Pool
	store  *database.Store
	logger *slog.Logger
}

// New cria uma nova instância de [Service].
func New(pool *pgxpool.Pool, logger *slog.Logger) *Service {
	return &Service{
		pool:   pool,
		store:  database.New(pool),
		logger: logger.With(slog.String("service", "diligencias")),
	}
}

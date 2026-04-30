package fila

import (
	"context"
	"database/sql"
	"errors"

	"github.com/automatiza-mg/fila/internal/auth"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

var _ auth.UsuarioHook = (*Service)(nil)

// TaskInserter define a interface para inserção de tarefas na fila.
type TaskInserter interface {
	InsertTx(ctx context.Context, tx pgx.Tx, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error)
}

type SEIService interface {
	EnviarProcesso(ctx context.Context, protocolo string, unidadeOrigem string, unidadesDestino []string) (*sei.EnviarProcessoResponse, error)
}

// Service gerencia a fila de processos de aposentadoria.
type Service struct {
	pool  *pgxpool.Pool
	store *database.Store
	queue TaskInserter
	sei   SEIService
}

// New cria uma nova instância de [Service].
func New(pool *pgxpool.Pool, queue TaskInserter, sei SEIService) *Service {
	return &Service{
		pool:  pool,
		store: database.New(pool),
		queue: queue,
		sei:   sei,
	}
}

// Label retorna o nome para [auth.UsuarioHook].
func (s *Service) Label() string {
	return "fila"
}

// GetActions implementa a interface para adicionar
// ações pendentes em usuários relacionados ao cadastro de dados de analista.
func (s *Service) GetActions(ctx context.Context, u *auth.Usuario) ([]auth.PendingAction, error) {
	if !u.IsAnalista() {
		return nil, nil
	}

	var actions []auth.PendingAction

	_, err := s.store.GetAnalista(ctx, u.ID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			actions = append(actions, auth.PendingAction{
				Slug:  "dados-analista",
				Title: "Registrar dados de analista",
			})
		} else {
			return nil, err
		}
	}

	return actions, nil
}

// TODO: Implementar as ações necessárias.
//
// Cleanup executa as ações de limpeza da fila quando um usuário é removido
// da aplicação ou tem seu papel alterado.
func (s *Service) Cleanup(ctx context.Context, tx pgx.Tx, trigger auth.CleanupTrigger, usuario *auth.Usuario) error {
	if !usuario.IsAnalista() {
		return nil
	}

	store := s.store.WithTx(tx)

	analista, err := store.GetAnalista(ctx, usuario.ID)
	if errors.Is(err, database.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	pa, err := store.GetProcessoAtribuido(ctx, analista.UsuarioID)
	if errors.Is(err, database.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	if err := s.saveHistorico(ctx, store, saveHistoricoParams{
		ProcessoAposentadoriaID: pa.ID,
		StatusAnterior:          &pa.Status,
		StatusNovo:              database.StatusProcessoAnalisePendente,
		Observacao:              "Processo desatribuído em razão de alteração do usuário",
	}); err != nil {
		return err
	}

	pa.Status = database.StatusProcessoAnalisePendente
	pa.AnalistaID = sql.Null[int64]{}
	pa.UltimoAnalistaID = sql.Null[int64]{}

	return store.UpdateProcessoAposentadoria(ctx, pa)
}

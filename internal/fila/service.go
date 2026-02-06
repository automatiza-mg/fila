package fila

import (
	"context"
	"errors"

	"github.com/automatiza-mg/fila/internal/auth"
	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ auth.LifecycleProvider = (*Service)(nil)

type SeiService interface {
	ListarUnidades(ctx context.Context) (*sei.ListarUnidadesResponse, error)
}

type IdentityService interface {
	GetUsuario(ctx context.Context, id int64) (*auth.Usuario, error)
}

type Service struct {
	pool  *pgxpool.Pool
	store *database.Store
	auth  IdentityService
	sei   SeiService
	cache cache.Cache
}

func New(pool *pgxpool.Pool, auth IdentityService, sei SeiService, cache cache.Cache) *Service {
	return &Service{
		pool:  pool,
		store: database.New(pool),
		auth:  auth,
		sei:   sei,
		cache: cache,
	}
}

// Label retorna o nome para [auth.LifecycleProvider].
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
func (s *Service) Cleanup(ctx context.Context, tx pgx.Tx, usuario *auth.Usuario) error {
	if !usuario.IsAnalista() {
		return nil
	}

	_, err := s.store.GetAnalista(ctx, usuario.ID)
	if errors.Is(err, database.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	return nil
}

package fila

import (
	"context"
	"database/sql"
	"errors"

	"github.com/automatiza-mg/fila/internal/auth"
	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/datalake"
	"github.com/automatiza-mg/fila/internal/llm"
	"github.com/automatiza-mg/fila/internal/processos"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ auth.UsuarioHook = (*Service)(nil)

type SeiClient interface {
	ListarUnidades(ctx context.Context) (*sei.ListarUnidadesResponse, error)
}

type ServidorProvider interface {
	GetServidorByCPF(ctx context.Context, cpf string) (*datalake.Servidor, error)
}

type AposentadoriaAnalyzer interface {
	AnalisarAposentadoria(ctx context.Context, docs []*processos.Documento) (*llm.AnaliseAposentadoria, error)
}

type Service struct {
	pool       *pgxpool.Pool
	store      *database.Store
	sei        SeiClient
	cache      cache.Cache
	analyzer   AposentadoriaAnalyzer
	servidores ServidorProvider
}

func New(pool *pgxpool.Pool, sei SeiClient, cache cache.Cache, analyzer AposentadoriaAnalyzer, servidores ServidorProvider) *Service {
	return &Service{
		pool:       pool,
		store:      database.New(pool),
		sei:        sei,
		cache:      cache,
		analyzer:   analyzer,
		servidores: servidores,
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

	pa.AnalistaID = sql.Null[int64]{}
	pa.UltimoAnalistaID = sql.Null[int64]{}

	store.SaveHistoricoStatusProcesso(ctx, &database.HistoricoStatusProcesso{
		ProcessoAposentadoriaID: pa.ID,
		StatusAnterior: sql.Null[database.StatusProcesso]{
			V:     pa.Status,
			Valid: true,
		},
		StatusNovo: database.StatusProcessoAnalisePendente,
		Observacao: sql.Null[string]{
			V: "Processo desatribuído em razão de alteração do usuário",
		},
	})

	return nil
}

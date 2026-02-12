package fila

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"testing"

	"github.com/automatiza-mg/fila/internal/auth"
	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/postgres"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/jackc/pgx/v5"
)

var (
	ti *postgres.TestInstance

	_ SeiService  = (*seiService)(nil)
	_ AuthService = (*authService)(nil)
)

func ptr[T any](v T) *T {
	return &v
}

type seiService struct{}

func (s *seiService) ListarUnidades(ctx context.Context) (*sei.ListarUnidadesResponse, error) {
	unidades := make([]sei.Unidade, 20)
	for i := range unidades {
		unidades[i] = sei.Unidade{
			IdUnidade: strconv.Itoa(i + 1),
			Sigla:     fmt.Sprintf("SEPLAG/AP%02d", i+1),
		}
	}

	unidades = append(unidades, sei.Unidade{
		IdUnidade: "TESTE",
		Sigla:     "TESTE",
	})

	return &sei.ListarUnidadesResponse{
		Parametros: sei.Parametros[sei.Unidade]{
			Items: unidades,
		},
	}, nil
}

type authService struct {
	usuarios map[int64]*auth.Usuario
}

func (a *authService) GetUsuario(ctx context.Context, id int64) (*auth.Usuario, error) {
	u, ok := a.usuarios[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return u, nil
}

func newTestService(tb testing.TB) *Service {
	tb.Helper()

	pool := ti.NewDatabase(tb)

	s := &seiService{}
	a := &authService{
		usuarios: make(map[int64]*auth.Usuario),
	}
	fila := New(pool, a, s, cache.NewMemoryCache(), nil)

	u := &database.Usuario{
		Nome:            "Fulano da Silva",
		CPF:             "123.456.789-09",
		Email:           "fulano@email.com",
		EmailVerificado: true,
		Papel: sql.Null[string]{
			V:     auth.PapelAnalista,
			Valid: true,
		},
	}
	err := fila.store.SaveUsuario(tb.Context(), u)
	if err != nil {
		tb.Fatal(err)
	}

	a.usuarios[u.ID] = auth.MapUsuario(u)

	return fila
}

func TestMain(m *testing.M) {
	ti = postgres.MustTestInstance()
	defer ti.Close()

	m.Run()
}

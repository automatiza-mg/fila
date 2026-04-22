package diligencias

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ti *postgres.TestInstance

func TestMain(m *testing.M) {
	ti = postgres.MustTestInstance()
	defer ti.Close()
	m.Run()
}

type testEnv struct {
	pool     *pgxpool.Pool
	store    *database.Store
	service  *Service
	pa       *database.ProcessoAposentadoria
	analista *database.Analista
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	pool := ti.NewDatabase(t)
	store := database.New(pool)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := New(pool, logger)

	pa, analista := seedProcessoEmAnalise(t, store)

	return &testEnv{
		pool:     pool,
		store:    store,
		service:  svc,
		pa:       pa,
		analista: analista,
	}
}

func seedProcessoEmAnalise(t *testing.T, store *database.Store) (*database.ProcessoAposentadoria, *database.Analista) {
	t.Helper()

	usuario := &database.Usuario{
		CPF:   rand.Text(),
		Email: rand.Text(),
		Papel: sql.Null[string]{V: "ANALISTA", Valid: true},
	}
	if err := store.SaveUsuario(t.Context(), usuario); err != nil {
		t.Fatal(err)
	}

	analista := &database.Analista{
		UsuarioID:       usuario.ID,
		Orgao:           "SEPLAG",
		SEIUnidadeID:    rand.Text(),
		SEIUnidadeSigla: "SEPLAG/AP00",
	}
	if err := store.SaveAnalista(t.Context(), analista); err != nil {
		t.Fatal(err)
	}

	p := &database.Processo{Numero: rand.Text()}
	if err := store.SaveProcesso(t.Context(), p); err != nil {
		t.Fatal(err)
	}

	pa := &database.ProcessoAposentadoria{
		ProcessoID: p.ID,
		Status:     database.StatusProcessoEmAnalise,
		AnalistaID: sql.Null[int64]{V: analista.UsuarioID, Valid: true},
	}
	if err := store.SaveProcessoAposentadoria(t.Context(), pa); err != nil {
		t.Fatal(err)
	}

	return pa, analista
}

func TestGetOrCreateRascunho_CreatesWhenMissing(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)

	sd, err := env.service.GetOrCreateRascunho(t.Context(), env.pa.ID, env.analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}
	if sd.ID == 0 {
		t.Fatal("expected ID to be set")
	}
	if sd.Status != string(database.StatusSolicitacaoRascunho) {
		t.Fatalf("expected status rascunho, got %q", sd.Status)
	}
	if len(sd.Itens) != 0 {
		t.Fatalf("expected no itens, got %d", len(sd.Itens))
	}
	if sd.EnviadaEm != nil {
		t.Fatal("expected EnviadaEm nil on new rascunho")
	}
}

func TestGetOrCreateRascunho_ReturnsExisting(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)

	first, err := env.service.GetOrCreateRascunho(t.Context(), env.pa.ID, env.analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}

	second, err := env.service.GetOrCreateRascunho(t.Context(), env.pa.ID, env.analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != first.ID {
		t.Fatalf("expected same rascunho id=%d, got %d", first.ID, second.ID)
	}
}

func TestGetOrCreateRascunho_NotAssigned(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)

	_, outro := seedProcessoEmAnalise(t, env.store)

	_, err := env.service.GetOrCreateRascunho(t.Context(), env.pa.ID, outro.UsuarioID)
	if !errors.Is(err, ErrNotAssigned) {
		t.Fatalf("want ErrNotAssigned, got %v", err)
	}
}

func TestGetOrCreateRascunho_InvalidStatus(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)

	env.pa.Status = database.StatusProcessoEmDiligencia
	if err := env.store.UpdateProcessoAposentadoria(t.Context(), env.pa); err != nil {
		t.Fatal(err)
	}

	_, err := env.service.GetOrCreateRascunho(t.Context(), env.pa.ID, env.analista.UsuarioID)
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("want ErrInvalidStatus, got %v", err)
	}
}

func TestGetRascunho_NotFound(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)

	_, err := env.service.GetRascunho(t.Context(), env.pa.ID, env.analista.UsuarioID)
	if !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestSalvarRascunho_ReplacesItems(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)

	sd, err := env.service.GetOrCreateRascunho(t.Context(), env.pa.ID, env.analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = env.service.SalvarRascunho(t.Context(), SalvarRascunhoParams{
		SolicitacaoID: sd.ID,
		AnalistaID:    env.analista.UsuarioID,
		Itens: []NovoItem{
			{Tipo: "Documentos Obrigatórios Ausentes", Subcategorias: []string{"FIPA - Dados Cadastrais"}},
			{Tipo: "Divergências de Informações entre Processo e SISAP", Detalhe: "X"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := env.service.SalvarRascunho(t.Context(), SalvarRascunhoParams{
		SolicitacaoID: sd.ID,
		AnalistaID:    env.analista.UsuarioID,
		Itens: []NovoItem{
			{Tipo: "Alteração de Dados Após o Envio", Detalhe: "Novo"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Itens) != 1 {
		t.Fatalf("expected 1 item after replace, got %d", len(got.Itens))
	}
	if got.Itens[0].Tipo != "Alteração de Dados Após o Envio" {
		t.Fatalf("unexpected tipo: %q", got.Itens[0].Tipo)
	}
}

func TestSalvarRascunho_EmptyAllowed(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)

	sd, err := env.service.GetOrCreateRascunho(t.Context(), env.pa.ID, env.analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}

	got, err := env.service.SalvarRascunho(t.Context(), SalvarRascunhoParams{
		SolicitacaoID: sd.ID,
		AnalistaID:    env.analista.UsuarioID,
		Itens:         nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Itens) != 0 {
		t.Fatalf("expected 0 itens, got %d", len(got.Itens))
	}
}

func TestSalvarRascunho_ForbidsSentSolicitacao(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)

	sd, err := env.service.GetOrCreateRascunho(t.Context(), env.pa.ID, env.analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = env.service.SalvarRascunho(t.Context(), SalvarRascunhoParams{
		SolicitacaoID: sd.ID,
		AnalistaID:    env.analista.UsuarioID,
		Itens:         []NovoItem{{Tipo: "X", Detalhe: "y"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := env.service.EnviarDiligencia(t.Context(), sd.ID, env.analista.UsuarioID); err != nil {
		t.Fatal(err)
	}

	_, err = env.service.SalvarRascunho(t.Context(), SalvarRascunhoParams{
		SolicitacaoID: sd.ID,
		AnalistaID:    env.analista.UsuarioID,
		Itens:         []NovoItem{{Tipo: "Z"}},
	})
	if !errors.Is(err, ErrAlreadySent) {
		t.Fatalf("want ErrAlreadySent, got %v", err)
	}
}

func TestSalvarRascunho_NotAssigned(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)

	sd, err := env.service.GetOrCreateRascunho(t.Context(), env.pa.ID, env.analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}

	_, outro := seedProcessoEmAnalise(t, env.store)

	_, err = env.service.SalvarRascunho(t.Context(), SalvarRascunhoParams{
		SolicitacaoID: sd.ID,
		AnalistaID:    outro.UsuarioID,
		Itens:         []NovoItem{{Tipo: "X"}},
	})
	if !errors.Is(err, ErrNotAssigned) {
		t.Fatalf("want ErrNotAssigned, got %v", err)
	}
}

func TestDescartarRascunho_Success(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)

	sd, err := env.service.GetOrCreateRascunho(t.Context(), env.pa.ID, env.analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}

	if err := env.service.DescartarRascunho(t.Context(), sd.ID, env.analista.UsuarioID); err != nil {
		t.Fatal(err)
	}

	if _, err := env.service.GetRascunho(t.Context(), env.pa.ID, env.analista.UsuarioID); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("want ErrNotFound after discard, got %v", err)
	}

	if _, err := env.service.GetOrCreateRascunho(t.Context(), env.pa.ID, env.analista.UsuarioID); err != nil {
		t.Fatalf("should allow new rascunho after discard, got %v", err)
	}
}

func TestDescartarRascunho_ForbidsSent(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)

	sd, err := env.service.GetOrCreateRascunho(t.Context(), env.pa.ID, env.analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = env.service.SalvarRascunho(t.Context(), SalvarRascunhoParams{
		SolicitacaoID: sd.ID,
		AnalistaID:    env.analista.UsuarioID,
		Itens:         []NovoItem{{Tipo: "X"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := env.service.EnviarDiligencia(t.Context(), sd.ID, env.analista.UsuarioID); err != nil {
		t.Fatal(err)
	}

	if err := env.service.DescartarRascunho(t.Context(), sd.ID, env.analista.UsuarioID); !errors.Is(err, ErrAlreadySent) {
		t.Fatalf("want ErrAlreadySent, got %v", err)
	}

	if _, err := env.store.GetSolicitacaoDiligencia(t.Context(), sd.ID); err != nil {
		t.Fatalf("sent solicitacao should remain, got %v", err)
	}
}

func TestEnviarDiligencia_Success(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)

	sd, err := env.service.GetOrCreateRascunho(t.Context(), env.pa.ID, env.analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = env.service.SalvarRascunho(t.Context(), SalvarRascunhoParams{
		SolicitacaoID: sd.ID,
		AnalistaID:    env.analista.UsuarioID,
		Itens: []NovoItem{
			{Tipo: "Documentos Obrigatórios Ausentes", Subcategorias: []string{"FIPA - Dados Cadastrais"}},
			{Tipo: "Alteração de Dados Após o Envio", Detalhe: "Algum detalhe"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	before := time.Now().UTC().Add(-time.Second)
	sent, err := env.service.EnviarDiligencia(t.Context(), sd.ID, env.analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}
	if sent.Status != string(database.StatusSolicitacaoEnviada) {
		t.Fatalf("expected status enviada, got %q", sent.Status)
	}
	if sent.EnviadaEm == nil || sent.EnviadaEm.Before(before) {
		t.Fatalf("EnviadaEm not set correctly: %v", sent.EnviadaEm)
	}
	if len(sent.Itens) != 2 {
		t.Fatalf("expected 2 itens, got %d", len(sent.Itens))
	}

	pa, err := env.store.GetProcessoAposentadoria(t.Context(), env.pa.ID)
	if err != nil {
		t.Fatal(err)
	}
	if pa.Status != database.StatusProcessoEmDiligencia {
		t.Fatalf("expected pa status EM_DILIGENCIA, got %q", pa.Status)
	}
	if pa.AnalistaID.Valid {
		t.Fatal("expected analista_id to be null")
	}
	if !pa.UltimoAnalistaID.Valid || pa.UltimoAnalistaID.V != env.analista.UsuarioID {
		t.Fatalf("expected ultimo_analista_id=%d, got %+v", env.analista.UsuarioID, pa.UltimoAnalistaID)
	}

	hh, err := env.store.ListHistoricoStatusProcesso(t.Context(), pa.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(hh) != 1 {
		t.Fatalf("expected 1 historico entry, got %d", len(hh))
	}
	h := hh[0]
	if h.StatusNovo != database.StatusProcessoEmDiligencia {
		t.Fatalf("expected StatusNovo=EM_DILIGENCIA, got %q", h.StatusNovo)
	}
	if !h.StatusAnterior.Valid || h.StatusAnterior.V != database.StatusProcessoEmAnalise {
		t.Fatalf("expected StatusAnterior=EM_ANALISE, got %+v", h.StatusAnterior)
	}
	if !h.UsuarioID.Valid || h.UsuarioID.V != env.analista.UsuarioID {
		t.Fatalf("expected UsuarioID=%d, got %+v", env.analista.UsuarioID, h.UsuarioID)
	}
	if !h.Observacao.Valid || h.Observacao.V != "Diligência solicitada" {
		t.Fatalf("unexpected observacao: %+v", h.Observacao)
	}
}

func TestEnviarDiligencia_DraftEmpty(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)

	sd, err := env.service.GetOrCreateRascunho(t.Context(), env.pa.ID, env.analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = env.service.EnviarDiligencia(t.Context(), sd.ID, env.analista.UsuarioID)
	if !errors.Is(err, ErrDraftEmpty) {
		t.Fatalf("want ErrDraftEmpty, got %v", err)
	}

	pa, err := env.store.GetProcessoAposentadoria(t.Context(), env.pa.ID)
	if err != nil {
		t.Fatal(err)
	}
	if pa.Status != database.StatusProcessoEmAnalise {
		t.Fatalf("expected pa still EM_ANALISE, got %q", pa.Status)
	}
}

func TestEnviarDiligencia_NotAssigned(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)

	sd, err := env.service.GetOrCreateRascunho(t.Context(), env.pa.ID, env.analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = env.service.SalvarRascunho(t.Context(), SalvarRascunhoParams{
		SolicitacaoID: sd.ID,
		AnalistaID:    env.analista.UsuarioID,
		Itens:         []NovoItem{{Tipo: "X"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, outro := seedProcessoEmAnalise(t, env.store)

	_, err = env.service.EnviarDiligencia(t.Context(), sd.ID, outro.UsuarioID)
	if !errors.Is(err, ErrNotAssigned) {
		t.Fatalf("want ErrNotAssigned, got %v", err)
	}
}

func TestListSolicitacoesEnviadas(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)

	// Rascunho not sent — should be excluded.
	_, err := env.service.GetOrCreateRascunho(t.Context(), env.pa.ID, env.analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}

	result, err := env.service.ListSolicitacoesEnviadas(t.Context(), env.pa.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 enviadas before any send, got %d", len(result))
	}

	// Send a solicitation.
	rascunho, err := env.service.GetRascunho(t.Context(), env.pa.ID, env.analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = env.service.SalvarRascunho(t.Context(), SalvarRascunhoParams{
		SolicitacaoID: rascunho.ID,
		AnalistaID:    env.analista.UsuarioID,
		Itens:         []NovoItem{{Tipo: "Algo", Detalhe: "x"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	sent, err := env.service.EnviarDiligencia(t.Context(), rascunho.ID, env.analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}

	result, err = env.service.ListSolicitacoesEnviadas(t.Context(), env.pa.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 enviada, got %d", len(result))
	}
	if result[0].ID != sent.ID {
		t.Fatalf("expected id=%d, got %d", sent.ID, result[0].ID)
	}
	if len(result[0].Itens) != 1 {
		t.Fatalf("expected 1 item in result, got %d", len(result[0].Itens))
	}
}

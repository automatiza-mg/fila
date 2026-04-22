package database

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func seedProcessoAposentadoria(t *testing.T, store *Store, numero string) *ProcessoAposentadoria {
	t.Helper()

	p := &Processo{Numero: numero}
	if err := store.SaveProcesso(t.Context(), p); err != nil {
		t.Fatal(err)
	}

	pa := &ProcessoAposentadoria{
		ProcessoID: p.ID,
		Status:     StatusProcessoAnalisePendente,
	}
	if err := store.SaveProcessoAposentadoria(t.Context(), pa); err != nil {
		t.Fatal(err)
	}
	return pa
}

func TestDiligenciaLifecycle(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	pa := seedProcessoAposentadoria(t, store, "38201-000010/2024-10")
	_, analista := seedAnalista(t, store)

	sd := &SolicitacaoDiligencia{
		ProcessoAposentadoriaID: pa.ID,
		AnalistaID:              analista.UsuarioID,
	}
	if err := store.SaveSolicitacaoDiligencia(t.Context(), sd); err != nil {
		t.Fatal(err)
	}
	if sd.ID == 0 {
		t.Fatal("expected ID to be populated")
	}
	if sd.CriadoEm.IsZero() {
		t.Fatal("expected CriadoEm to be populated")
	}
	if sd.Status != StatusSolicitacaoRascunho {
		t.Fatalf("expected default status rascunho, got %q", sd.Status)
	}
	if sd.EnviadaEm.Valid {
		t.Fatal("expected EnviadaEm to be null on creation")
	}

	item1 := &ItemDiligencia{
		SolicitacaoDiligenciaID: sd.ID,
		Tipo:                    "Documentos Obrigatórios Ausentes",
		Subcategorias:           []string{"FIPA - Dados Cadastrais", "FIPA - Tempo Averbado"},
		Detalhe:                 "",
	}
	if err := store.SaveItemDiligencia(t.Context(), item1); err != nil {
		t.Fatal(err)
	}

	item2 := &ItemDiligencia{
		SolicitacaoDiligenciaID: sd.ID,
		Tipo:                    "Divergências de Informações entre Processo e SISAP",
		Subcategorias:           []string{},
		Detalhe:                 "Data de nascimento divergente entre os sistemas.",
	}
	if err := store.SaveItemDiligencia(t.Context(), item2); err != nil {
		t.Fatal(err)
	}

	read, err := store.GetSolicitacaoDiligencia(t.Context(), sd.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(sd, read); diff != "" {
		t.Fatalf("solicitacao mismatch:\n%s", diff)
	}

	ss, err := store.ListSolicitacoesDiligenciaByProcesso(t.Context(), ListSolicitacoesDiligenciaParams{
		ProcessoAposentadoriaID: pa.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(ss) != 1 {
		t.Fatalf("expected 1 solicitacao, got %d", len(ss))
	}
	if diff := cmp.Diff(sd, ss[0]); diff != "" {
		t.Fatalf("list mismatch:\n%s", diff)
	}

	items, err := store.ListItensDiligencia(t.Context(), sd.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 itens, got %d", len(items))
	}
	if diff := cmp.Diff(item1, items[0]); diff != "" {
		t.Fatalf("item1 mismatch:\n%s", diff)
	}
	if diff := cmp.Diff(item2, items[1]); diff != "" {
		t.Fatalf("item2 mismatch:\n%s", diff)
	}

	if err := store.DeleteItensDiligencia(t.Context(), sd.ID); err != nil {
		t.Fatal(err)
	}
	items, err = store.ListItensDiligencia(t.Context(), sd.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 itens after delete, got %d", len(items))
	}

	item3 := &ItemDiligencia{
		SolicitacaoDiligenciaID: sd.ID,
		Tipo:                    "Alteração de Dados Após o Envio",
		Subcategorias:           []string{},
		Detalhe:                 "Nova observação",
	}
	if err := store.SaveItemDiligencia(t.Context(), item3); err != nil {
		t.Fatal(err)
	}
	items, err = store.ListItensDiligencia(t.Context(), sd.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item after reinsert, got %d", len(items))
	}

	if err := store.DeleteSolicitacaoDiligencia(t.Context(), sd.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetSolicitacaoDiligencia(t.Context(), sd.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound after delete, got %v", err)
	}
	items, err = store.ListItensDiligencia(t.Context(), sd.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 itens after cascade delete, got %d", len(items))
	}
}

func TestSolicitacaoDiligencia_NotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	_, err := store.GetSolicitacaoDiligencia(t.Context(), 999999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}

	_, err = store.GetRascunhoDiligencia(t.Context(), 999999, 999999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound for rascunho, got %v", err)
	}
}

func TestGetRascunhoDiligencia(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	pa := seedProcessoAposentadoria(t, store, "38201-000020/2024-10")
	_, analista := seedAnalista(t, store)

	sd := &SolicitacaoDiligencia{
		ProcessoAposentadoriaID: pa.ID,
		AnalistaID:              analista.UsuarioID,
	}
	if err := store.SaveSolicitacaoDiligencia(t.Context(), sd); err != nil {
		t.Fatal(err)
	}

	got, err := store.GetRascunhoDiligencia(t.Context(), pa.ID, analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(sd, got); diff != "" {
		t.Fatalf("rascunho mismatch:\n%s", diff)
	}

	sd.Status = StatusSolicitacaoEnviada
	sd.EnviadaEm = sql.Null[time.Time]{V: time.Now().UTC(), Valid: true}
	if err := store.UpdateSolicitacaoDiligencia(t.Context(), sd); err != nil {
		t.Fatal(err)
	}

	_, err = store.GetRascunhoDiligencia(t.Context(), pa.ID, analista.UsuarioID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound once enviada, got %v", err)
	}
}

func TestRascunhoUniqueness(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	pa := seedProcessoAposentadoria(t, store, "38201-000030/2024-10")
	_, analista := seedAnalista(t, store)

	first := &SolicitacaoDiligencia{
		ProcessoAposentadoriaID: pa.ID,
		AnalistaID:              analista.UsuarioID,
	}
	if err := store.SaveSolicitacaoDiligencia(t.Context(), first); err != nil {
		t.Fatal(err)
	}

	dup := &SolicitacaoDiligencia{
		ProcessoAposentadoriaID: pa.ID,
		AnalistaID:              analista.UsuarioID,
	}
	err := store.SaveSolicitacaoDiligencia(t.Context(), dup)
	if err == nil {
		t.Fatal("expected unique violation for second rascunho, got nil")
	}

	first.Status = StatusSolicitacaoEnviada
	first.EnviadaEm = sql.Null[time.Time]{V: time.Now().UTC(), Valid: true}
	if err := store.UpdateSolicitacaoDiligencia(t.Context(), first); err != nil {
		t.Fatal(err)
	}

	newRascunho := &SolicitacaoDiligencia{
		ProcessoAposentadoriaID: pa.ID,
		AnalistaID:              analista.UsuarioID,
	}
	if err := store.SaveSolicitacaoDiligencia(t.Context(), newRascunho); err != nil {
		t.Fatalf("should allow new rascunho after prior one was enviada: %v", err)
	}
}

func TestDescartaRascunhoLiberaUnicidade(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	pa := seedProcessoAposentadoria(t, store, "38201-000040/2024-10")
	_, analista := seedAnalista(t, store)

	first := &SolicitacaoDiligencia{
		ProcessoAposentadoriaID: pa.ID,
		AnalistaID:              analista.UsuarioID,
	}
	if err := store.SaveSolicitacaoDiligencia(t.Context(), first); err != nil {
		t.Fatal(err)
	}

	if err := store.DeleteSolicitacaoDiligencia(t.Context(), first.ID); err != nil {
		t.Fatal(err)
	}

	novo := &SolicitacaoDiligencia{
		ProcessoAposentadoriaID: pa.ID,
		AnalistaID:              analista.UsuarioID,
	}
	if err := store.SaveSolicitacaoDiligencia(t.Context(), novo); err != nil {
		t.Fatalf("should allow new rascunho after discard: %v", err)
	}
}

func TestListSolicitacoesDiligencia_StatusFilter(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	pa := seedProcessoAposentadoria(t, store, "38201-000050/2024-10")
	_, analistaA := seedAnalista(t, store)
	_, analistaB := seedAnalista(t, store)

	enviada := &SolicitacaoDiligencia{
		ProcessoAposentadoriaID: pa.ID,
		AnalistaID:              analistaA.UsuarioID,
	}
	if err := store.SaveSolicitacaoDiligencia(t.Context(), enviada); err != nil {
		t.Fatal(err)
	}
	enviada.Status = StatusSolicitacaoEnviada
	enviada.EnviadaEm = sql.Null[time.Time]{V: time.Now().UTC(), Valid: true}
	if err := store.UpdateSolicitacaoDiligencia(t.Context(), enviada); err != nil {
		t.Fatal(err)
	}

	rascunho := &SolicitacaoDiligencia{
		ProcessoAposentadoriaID: pa.ID,
		AnalistaID:              analistaB.UsuarioID,
	}
	if err := store.SaveSolicitacaoDiligencia(t.Context(), rascunho); err != nil {
		t.Fatal(err)
	}

	all, err := store.ListSolicitacoesDiligenciaByProcesso(t.Context(), ListSolicitacoesDiligenciaParams{
		ProcessoAposentadoriaID: pa.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 solicitacoes without filter, got %d", len(all))
	}

	onlyEnviadas, err := store.ListSolicitacoesDiligenciaByProcesso(t.Context(), ListSolicitacoesDiligenciaParams{
		ProcessoAposentadoriaID: pa.ID,
		Status:                  StatusSolicitacaoEnviada,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(onlyEnviadas) != 1 {
		t.Fatalf("expected 1 enviada, got %d", len(onlyEnviadas))
	}
	if onlyEnviadas[0].ID != enviada.ID {
		t.Fatalf("expected enviada id %d, got %d", enviada.ID, onlyEnviadas[0].ID)
	}

	onlyRascunhos, err := store.ListSolicitacoesDiligenciaByProcesso(t.Context(), ListSolicitacoesDiligenciaParams{
		ProcessoAposentadoriaID: pa.ID,
		Status:                  StatusSolicitacaoRascunho,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(onlyRascunhos) != 1 {
		t.Fatalf("expected 1 rascunho, got %d", len(onlyRascunhos))
	}
	if onlyRascunhos[0].ID != rascunho.ID {
		t.Fatalf("expected rascunho id %d, got %d", rascunho.ID, onlyRascunhos[0].ID)
	}
}

package database

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDiligenciaLifecycle(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	p := &Processo{Numero: "38201-000010/2024-10"}
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

	ss, err := store.ListSolicitacoesDiligenciaByProcesso(t.Context(), pa.ID)
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
}

package database

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestDocumentoLifecycle(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	p := &Processo{
		Numero: "123123",
	}
	err := store.SaveProcesso(t.Context(), p)
	if err != nil {
		t.Fatal(err)
	}

	d := &Documento{
		Numero:       "123123",
		ProcessoID:   p.ID,
		MetadadosAPI: []byte("{}"),
	}
	err = store.SaveDocumento(t.Context(), d)
	if err != nil {
		t.Fatal(err)
	}

	d2, err := store.GetDocumento(t.Context(), d.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(d, d2); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}

	d2, err = store.GetDocumentoByNumero(t.Context(), d.Numero)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(d, d2); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}
}

func TestDocumento_NotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	_, err := store.GetDocumento(t.Context(), 999999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got: %v", err)
	}

	_, err = store.GetDocumentoByNumero(t.Context(), "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got: %v", err)
	}
}

func TestDocumento_ListDocumentos(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	p := &Processo{
		Numero: "list-doc-processo",
	}
	err := store.SaveProcesso(t.Context(), p)
	if err != nil {
		t.Fatal(err)
	}

	docs := make([]*Documento, 3)
	for i := range docs {
		docs[i] = &Documento{
			Numero:       fmt.Sprintf("list-doc-%d", i),
			ProcessoID:   p.ID,
			MetadadosAPI: []byte("{}"),
		}
		err := store.SaveDocumento(t.Context(), docs[i])
		if err != nil {
			t.Fatal(err)
		}
	}

	dd, err := store.ListDocumentos(t.Context(), p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(dd) != 3 {
		t.Fatalf("expected len=3, got len=%d", len(dd))
	}

	// No documentos for a random processo
	dd, err = store.ListDocumentos(t.Context(), uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if len(dd) != 0 {
		t.Fatalf("expected len=0, got len=%d", len(dd))
	}
}

func TestDocumento_GetDocumentosMap(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	p1 := &Processo{Numero: "map-doc-p1"}
	p2 := &Processo{Numero: "map-doc-p2"}
	err := store.SaveProcesso(t.Context(), p1)
	if err != nil {
		t.Fatal(err)
	}
	err = store.SaveProcesso(t.Context(), p2)
	if err != nil {
		t.Fatal(err)
	}

	// 2 documentos under p1
	for i := range 2 {
		d := &Documento{
			Numero:       fmt.Sprintf("map-doc-p1-%d", i),
			ProcessoID:   p1.ID,
			MetadadosAPI: []byte("{}"),
		}
		err := store.SaveDocumento(t.Context(), d)
		if err != nil {
			t.Fatal(err)
		}
	}

	// 1 documento under p2
	d := &Documento{
		Numero:       "map-doc-p2-0",
		ProcessoID:   p2.ID,
		MetadadosAPI: []byte("{}"),
	}
	err = store.SaveDocumento(t.Context(), d)
	if err != nil {
		t.Fatal(err)
	}

	docMap, err := store.GetDocumentosMap(t.Context(), []uuid.UUID{p1.ID, p2.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(docMap) != 2 {
		t.Fatalf("expected 2 keys in map, got %d", len(docMap))
	}
	if len(docMap[p1.ID]) != 2 {
		t.Fatalf("expected 2 docs for p1, got %d", len(docMap[p1.ID]))
	}
	if len(docMap[p2.ID]) != 1 {
		t.Fatalf("expected 1 doc for p2, got %d", len(docMap[p2.ID]))
	}

	// Empty input returns empty map
	docMap, err = store.GetDocumentosMap(t.Context(), []uuid.UUID{})
	if err != nil {
		t.Fatal(err)
	}
	if len(docMap) != 0 {
		t.Fatalf("expected empty map, got %d keys", len(docMap))
	}

	// Non-existent processo returns empty map
	docMap, err = store.GetDocumentosMap(t.Context(), []uuid.UUID{uuid.New()})
	if err != nil {
		t.Fatal(err)
	}
	if len(docMap) != 0 {
		t.Fatalf("expected empty map, got %d keys", len(docMap))
	}
}

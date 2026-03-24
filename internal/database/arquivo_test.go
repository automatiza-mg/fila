package database

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestArquivoLifecycle(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	a := &Arquivo{
		Hash:         "abc123hash",
		ChaveStorage: "processos/abc123hash.pdf",
		OCR:          "conteúdo do documento",
		ContentType:  "application/pdf",
	}
	err := store.SaveArquivo(t.Context(), a)
	if err != nil {
		t.Fatal(err)
	}
	if a.CriadoEm.IsZero() {
		t.Fatal("expected criado_em to be set")
	}

	a2, err := store.GetArquivo(t.Context(), a.Hash)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(a, a2); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}

	err = store.DeleteArquivo(t.Context(), a.Hash)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.GetArquivo(t.Context(), a.Hash)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestArquivo_SaveConflict(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	a := &Arquivo{
		Hash:         "conflict-hash",
		ChaveStorage: "processos/conflict-hash.pdf",
		OCR:          "texto original",
		ContentType:  "application/pdf",
	}
	err := store.SaveArquivo(t.Context(), a)
	if err != nil {
		t.Fatal(err)
	}

	// Segundo insert com mesmo hash deve ser ignorado sem erro.
	a2 := &Arquivo{
		Hash:         "conflict-hash",
		ChaveStorage: "processos/outro.pdf",
		OCR:          "texto diferente",
		ContentType:  "image/png",
	}
	err = store.SaveArquivo(t.Context(), a2)
	if err != nil {
		t.Fatal(err)
	}
	if !a2.CriadoEm.IsZero() {
		t.Fatal("expected criado_em to remain zero on conflict")
	}

	// Dados originais devem permanecer inalterados.
	got, err := store.GetArquivo(t.Context(), "conflict-hash")
	if err != nil {
		t.Fatal(err)
	}
	if got.ChaveStorage != "processos/conflict-hash.pdf" {
		t.Fatalf("expected original chave_storage, got %q", got.ChaveStorage)
	}
	if got.OCR != "texto original" {
		t.Fatalf("expected original ocr, got %q", got.OCR)
	}
}

func TestArquivo_GetArquivosMap(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	a1 := &Arquivo{Hash: "map-hash-1", ChaveStorage: "arquivos/map-hash-1", OCR: "texto 1", ContentType: "application/pdf"}
	a2 := &Arquivo{Hash: "map-hash-2", ChaveStorage: "arquivos/map-hash-2", OCR: "texto 2", ContentType: "image/png"}
	err := store.SaveArquivo(t.Context(), a1)
	if err != nil {
		t.Fatal(err)
	}
	err = store.SaveArquivo(t.Context(), a2)
	if err != nil {
		t.Fatal(err)
	}

	arquivoMap, err := store.GetArquivosMap(t.Context(), []string{a1.Hash, a2.Hash})
	if err != nil {
		t.Fatal(err)
	}
	if len(arquivoMap) != 2 {
		t.Fatalf("expected 2 keys in map, got %d", len(arquivoMap))
	}
	if diff := cmp.Diff(a1, arquivoMap[a1.Hash]); diff != "" {
		t.Fatalf("a1 mismatch:\n%s", diff)
	}
	if diff := cmp.Diff(a2, arquivoMap[a2.Hash]); diff != "" {
		t.Fatalf("a2 mismatch:\n%s", diff)
	}

	// Empty input returns empty map
	arquivoMap, err = store.GetArquivosMap(t.Context(), []string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(arquivoMap) != 0 {
		t.Fatalf("expected empty map, got %d keys", len(arquivoMap))
	}

	// Non-existent hashes return empty map
	arquivoMap, err = store.GetArquivosMap(t.Context(), []string{"nonexistent"})
	if err != nil {
		t.Fatal(err)
	}
	if len(arquivoMap) != 0 {
		t.Fatalf("expected empty map, got %d keys", len(arquivoMap))
	}
}

func TestArquivo_NotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	_, err := store.GetArquivo(t.Context(), "nonexistent-hash")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got: %v", err)
	}
}

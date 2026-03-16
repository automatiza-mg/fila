package pipeline

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestPipeline_RunsStepsInOrder(t *testing.T) {
	t.Parallel()

	var order []string
	step1 := NewStep("step1", func(_ context.Context, _ *State) error {
		order = append(order, "step1")
		return nil
	})
	step2 := NewStep("step2", func(_ context.Context, _ *State) error {
		order = append(order, "step2")
		return nil
	})
	step3 := NewStep("step3", func(_ context.Context, _ *State) error {
		order = append(order, "step3")
		return nil
	})

	p := New(step1, step2, step3)
	err := p.Run(t.Context(), &State{})
	if err != nil {
		t.Fatal(err)
	}

	if len(order) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(order))
	}
	for i, want := range []string{"step1", "step2", "step3"} {
		if order[i] != want {
			t.Fatalf("step %d: expected %q, got %q", i, want, order[i])
		}
	}
}

func TestPipeline_StopsOnError(t *testing.T) {
	t.Parallel()

	step1 := NewStep("step1", func(_ context.Context, _ *State) error {
		return nil
	})
	step2 := NewStep("step2", func(_ context.Context, _ *State) error {
		return fmt.Errorf("boom")
	})
	step3 := NewStep("step3", func(_ context.Context, _ *State) error {
		t.Fatal("step3 should not run")
		return nil
	})

	p := New(step1, step2, step3)
	err := p.Run(t.Context(), &State{})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "step2: boom" {
		t.Fatalf("expected 'step2: boom', got %q", err.Error())
	}
}

func TestPipeline_EmptyPipeline(t *testing.T) {
	t.Parallel()

	p := New()
	err := p.Run(t.Context(), &State{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestListDocumentos(t *testing.T) {
	t.Parallel()

	lister := &fakeLister{
		docs: []string{"DOC-001", "DOC-002"},
	}

	state := &State{
		LinkAcesso: "https://sei.example.com/processo/123",
	}

	step := ListDocumentos(lister)
	if step.Name() != "listar_documentos" {
		t.Fatalf("expected name 'listar_documentos', got %q", step.Name())
	}

	err := step.Run(t.Context(), state)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.DocumentosListados) != 2 {
		t.Fatalf("expected 2 docs, got %d", len(state.DocumentosListados))
	}
	if state.DocumentosListados[0] != "DOC-001" {
		t.Fatalf("expected DOC-001, got %s", state.DocumentosListados[0])
	}
}

func TestListDocumentos_Error(t *testing.T) {
	t.Parallel()

	lister := &fakeLister{err: fmt.Errorf("sei unavailable")}

	state := &State{
		LinkAcesso: "https://sei.example.com",
	}

	step := ListDocumentos(lister)
	err := step.Run(t.Context(), state)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFiltrarNovos(t *testing.T) {
	t.Parallel()

	checker := &fakeChecker{
		existing: map[string]bool{"DOC-001": true},
	}

	state := &State{
		DocumentosListados: []string{"DOC-001", "DOC-002", "DOC-003"},
	}

	step := FiltrarNovos(checker)
	if step.Name() != "filtrar_novos" {
		t.Fatalf("expected name 'filtrar_novos', got %q", step.Name())
	}

	err := step.Run(t.Context(), state)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.DocumentosListados) != 2 {
		t.Fatalf("expected 2 new docs, got %d", len(state.DocumentosListados))
	}
	if state.DocumentosListados[0] != "DOC-002" {
		t.Fatalf("expected DOC-002, got %s", state.DocumentosListados[0])
	}
	if state.DocumentosListados[1] != "DOC-003" {
		t.Fatalf("expected DOC-003, got %s", state.DocumentosListados[1])
	}
}

func TestFiltrarNovos_AllExisting(t *testing.T) {
	t.Parallel()

	checker := &fakeChecker{
		existing: map[string]bool{"DOC-001": true, "DOC-002": true},
	}

	state := &State{
		DocumentosListados: []string{"DOC-001", "DOC-002"},
	}

	step := FiltrarNovos(checker)
	err := step.Run(t.Context(), state)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.DocumentosListados) != 0 {
		t.Fatalf("expected 0 new docs, got %d", len(state.DocumentosListados))
	}
}

func TestFiltrarNovos_Error(t *testing.T) {
	t.Parallel()

	checker := &fakeChecker{err: fmt.Errorf("db error")}

	state := &State{
		DocumentosListados: []string{"DOC-001"},
	}

	step := FiltrarNovos(checker)
	err := step.Run(t.Context(), state)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuscarDocumentos(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{
		docs: []DocBuscado{
			{Numero: "DOC-001", Conteudo: "texto extraido"},
		},
	}

	state := &State{
		DocumentosListados: []string{"DOC-001"},
	}

	step := BuscarDocumentos(fetcher)
	if step.Name() != "buscar_documentos" {
		t.Fatalf("expected name 'buscar_documentos', got %q", step.Name())
	}

	err := step.Run(t.Context(), state)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.DocumentosBuscados) != 1 {
		t.Fatalf("expected 1 doc, got %d", len(state.DocumentosBuscados))
	}
	if state.DocumentosBuscados[0].Conteudo != "texto extraido" {
		t.Fatalf("expected 'texto extraido', got %q", state.DocumentosBuscados[0].Conteudo)
	}
}

func TestBuscarDocumentos_Empty(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{}
	state := &State{
		DocumentosListados: nil,
	}

	step := BuscarDocumentos(fetcher)
	err := step.Run(t.Context(), state)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.DocumentosBuscados) != 0 {
		t.Fatalf("expected 0 docs, got %d", len(state.DocumentosBuscados))
	}
}

func TestBuscarDocumentos_Error(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{err: fmt.Errorf("fetch error")}
	state := &State{
		DocumentosListados: []string{"DOC-001"},
	}

	step := BuscarDocumentos(fetcher)
	err := step.Run(t.Context(), state)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAtualizarStatus(t *testing.T) {
	t.Parallel()

	updater := &fakeStatusUpdater{}

	state := &State{
		ProcessoID: uuid.New(),
		Status:     "PENDENTE",
	}

	step := AtualizarStatus("PROCESSANDO", updater)
	err := step.Run(t.Context(), state)
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != "PROCESSANDO" {
		t.Fatalf("expected PROCESSANDO, got %s", state.Status)
	}
	if !updater.called {
		t.Fatal("expected updater to be called")
	}
	if updater.status != "PROCESSANDO" {
		t.Fatalf("expected updater status PROCESSANDO, got %s", updater.status)
	}
}

func TestAtualizarStatus_Error(t *testing.T) {
	t.Parallel()

	updater := &fakeStatusUpdater{err: fmt.Errorf("db error")}

	state := &State{
		ProcessoID: uuid.New(),
		Status:     "PENDENTE",
	}

	step := AtualizarStatus("PROCESSANDO", updater)
	err := step.Run(t.Context(), state)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPersistirDocumentos(t *testing.T) {
	t.Parallel()

	persister := &fakePersister{}

	state := &State{
		ProcessoID: uuid.New(),
		DocumentosBuscados: []DocBuscado{
			{Numero: "DOC-001", Tipo: "Oficio", Conteudo: "text1"},
			{Numero: "DOC-002", Tipo: "Despacho", Conteudo: "text2"},
		},
	}

	step := PersistirDocumentos(persister)
	err := step.Run(t.Context(), state)
	if err != nil {
		t.Fatal(err)
	}
	if !persister.called {
		t.Fatal("expected persister to be called")
	}
	if len(persister.docs) != 2 {
		t.Fatalf("expected 2 docs, got %d", len(persister.docs))
	}
}

func TestPersistirDocumentos_Error(t *testing.T) {
	t.Parallel()

	persister := &fakePersister{err: fmt.Errorf("db error")}

	state := &State{
		ProcessoID: uuid.New(),
		DocumentosBuscados: []DocBuscado{
			{Numero: "DOC-001"},
		},
	}

	step := PersistirDocumentos(persister)
	err := step.Run(t.Context(), state)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFullPipeline(t *testing.T) {
	t.Parallel()

	lister := &fakeLister{
		docs: []string{"DOC-001", "DOC-002", "DOC-003"},
	}

	checker := &fakeChecker{
		existing: map[string]bool{"DOC-001": true},
	}

	fetcher := &fakeFetcher{
		docs: []DocBuscado{
			{Numero: "DOC-002", Conteudo: "text2", Tipo: "Oficio"},
			{Numero: "DOC-003", Conteudo: "text3", Tipo: "Despacho"},
		},
	}

	updater := &fakeStatusUpdater{}
	persister := &fakePersister{}

	state := &State{
		ProcessoID: uuid.New(),
		LinkAcesso: "https://sei.example.com",
		Status:     "PENDENTE",
	}

	p := New(
		ListDocumentos(lister),
		FiltrarNovos(checker),
		BuscarDocumentos(fetcher),
		AtualizarStatus("PROCESSANDO", updater),
		PersistirDocumentos(persister),
		AtualizarStatus("SUCESSO", updater),
	)

	err := p.Run(t.Context(), state)
	if err != nil {
		t.Fatal(err)
	}

	if len(state.DocumentosBuscados) != 2 {
		t.Fatalf("expected 2 fetched docs, got %d", len(state.DocumentosBuscados))
	}
	if state.Status != "SUCESSO" {
		t.Fatalf("expected SUCESSO, got %s", state.Status)
	}
	if !persister.called {
		t.Fatal("expected persister to be called")
	}
}

func TestStepFunc(t *testing.T) {
	t.Parallel()

	called := false
	step := NewStep("test", func(_ context.Context, _ *State) error {
		called = true
		return nil
	})

	if step.Name() != "test" {
		t.Fatalf("expected name 'test', got %q", step.Name())
	}

	err := step.Run(t.Context(), &State{})
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected function to be called")
	}
}

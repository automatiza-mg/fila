package processos

import (
	"context"
	"fmt"
	"testing"

	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var ignoreDocumentoFields = cmpopts.IgnoreFields(Documento{}, "ID")

func TestRegisterHook(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)

	if len(ts.svc.hooks) != 0 {
		t.Fatalf("expected 0 hooks, got %d", len(ts.svc.hooks))
	}

	ts.svc.RegisterHook(&testHook{})

	if len(ts.svc.hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(ts.svc.hooks))
	}
}

func TestAnalyze_NotifiesHooks(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)
	proc := seedProcesso(t, ts.svc, "hook-notify")

	apiData := sei.RetornoConsultaDocumento{
		Data: "05/02/2026",
		Serie: sei.Serie{
			IdSerie: "1",
			Nome:    "Oficio",
		},
		UnidadeElaboradora: sei.UnidadeElaboradora{
			IdUnidade: "100",
			Sigla:     "SEPLAG/AP01",
		},
		LinkAcesso: "https://sei.example.com/doc/hook-001",
		Assinaturas: sei.Assinaturas{
			Itens: []sei.Assinatura{
				{Nome: "Joao Silva", Sigla: "123.456.789-00"},
			},
		},
	}

	// Stub SEI to return one document in the listing.
	ts.sei.listarDocumentosFn = func(_ context.Context, _ string) ([]sei.LinhaDocumento, error) {
		return []sei.LinhaDocumento{
			{Numero: "HOOK-DOC-001"},
		}, nil
	}

	// Stub fetcher to return the document data.
	ts.fetcher.docs = []DocumentoSei{
		{
			Numero:      "HOOK-DOC-001",
			Conteudo:    "conteudo hook doc",
			ContentType: "application/pdf",
			APIData:     apiData,
		},
	}
	hook := &testHook{}
	ts.svc.RegisterHook(hook)

	err := ts.svc.Analyze(t.Context(), proc.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !hook.called {
		t.Fatal("expected hook to be called")
	}

	wantDocs := []*Documento{
		{
			Numero:          "HOOK-DOC-001",
			Tipo:            "Oficio",
			Conteudo:        "conteudo hook doc",
			LinkAcesso:      "https://sei.example.com/doc/hook-001",
			Data:            "05/02/2026",
			UnidadeGeradora: "SEPLAG/AP01",
			Assinaturas: []Assinatura{
				{Nome: "Joao Silva", CPF: "123.456.789-00"},
			},
		},
	}
	if diff := cmp.Diff(wantDocs, hook.documentos, ignoreDocumentoFields); diff != "" {
		t.Fatalf("hook documentos mismatch (-want +got):\n%s", diff)
	}
}

func TestAnalyze_MultipleHooks(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)
	proc := seedProcesso(t, ts.svc, "multi-hook")

	// Stub SEI to return no documents (simplest path through processDocs).
	ts.sei.listarDocumentosFn = func(_ context.Context, _ string) ([]sei.LinhaDocumento, error) {
		return nil, nil
	}

	hook1 := &testHook{}
	hook2 := &testHook{}
	ts.svc.RegisterHook(hook1)
	ts.svc.RegisterHook(hook2)

	err := ts.svc.Analyze(t.Context(), proc.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !hook1.called {
		t.Fatal("expected hook1 to be called")
	}
	if !hook2.called {
		t.Fatal("expected hook2 to be called")
	}

	wantProcesso := &Processo{
		ID:              proc.ID,
		Numero:          "multi-hook",
		Status:          "SUCESSO",
		LinkAcesso:      "https://sei.example.com/processo/multi-hook",
		SeiUnidadeID:    "100",
		SeiUnidadeSigla: "SEPLAG/AP01",
	}
	if diff := cmp.Diff(wantProcesso, hook1.processo, ignoreProcessoFields); diff != "" {
		t.Fatalf("hook1 processo mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(wantProcesso, hook2.processo, ignoreProcessoFields); diff != "" {
		t.Fatalf("hook2 processo mismatch (-want +got):\n%s", diff)
	}
}

func TestAnalyze_HookError(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)
	proc := seedProcesso(t, ts.svc, "hook-error")

	ts.sei.listarDocumentosFn = func(_ context.Context, _ string) ([]sei.LinhaDocumento, error) {
		return nil, nil
	}

	errHook := &testErrorHook{err: fmt.Errorf("hook failed")}
	hook2 := &testHook{}
	ts.svc.RegisterHook(errHook)
	ts.svc.RegisterHook(hook2)

	err := ts.svc.Analyze(t.Context(), proc.ID)
	if err == nil {
		t.Fatal("expected error from hook")
	}
	if err.Error() != "hook failed" {
		t.Fatalf("expected 'hook failed', got: %v", err)
	}

	// Second hook should NOT be called because first one errored.
	if hook2.called {
		t.Fatal("expected hook2 NOT to be called after hook1 error")
	}
}

// testErrorHook is a hook that always returns an error.
type testErrorHook struct {
	err error
}

func (h *testErrorHook) OnAnalyzeComplete(_ context.Context, _ *Processo, _ []*Documento) error {
	return h.err
}

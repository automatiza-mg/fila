package processos

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
)

var (
	ignoreProcessoFields   = cmpopts.IgnoreFields(Processo{}, "Aposentadoria", "AnalisadoEm", "CriadoEm", "AtualizadoEm", "MetadadosIA")
	ignoreProcessoIDFields = cmpopts.IgnoreFields(Processo{}, "ID", "Aposentadoria", "AnalisadoEm", "CriadoEm", "AtualizadoEm", "MetadadosIA")
)

func consultarProcedimentoOK(_ context.Context, protocolo string) (*sei.ConsultarProcedimentoResponse, error) {
	return &sei.ConsultarProcedimentoResponse{
		Parametros: sei.RetornoConsultaProcedimento{
			LinkAcesso: "https://sei.example.com/processo/" + protocolo,
			AndamentoGeracao: sei.Andamento{
				Unidade: sei.Unidade{
					IdUnidade: "100",
					Sigla:     "SEPLAG/AP01",
				},
			},
		},
	}, nil
}

func TestCreateProcesso(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)
	ts.sei.consultarProcedimentoFn = consultarProcedimentoOK

	p, err := ts.svc.CreateProcesso(t.Context(), "123456")
	if err != nil {
		t.Fatal(err)
	}

	want := &Processo{
		Numero:          "123456",
		Status:          "PENDENTE",
		LinkAcesso:      "https://sei.example.com/processo/123456",
		SeiUnidadeID:    "100",
		SeiUnidadeSigla: "SEPLAG/AP01",
	}
	if diff := cmp.Diff(want, p, ignoreProcessoIDFields); diff != "" {
		t.Fatalf("CreateProcesso mismatch (-want +got):\n%s", diff)
	}

	// Verify persisted
	p2, err := ts.svc.GetProcessoByNumero(t.Context(), "123456")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(p, p2); diff != "" {
		t.Fatalf("persisted processo mismatch (-want +got):\n%s", diff)
	}
}

func TestCreateProcesso_Duplicate(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)
	ts.sei.consultarProcedimentoFn = consultarProcedimentoOK

	_, err := ts.svc.CreateProcesso(t.Context(), "dup-001")
	if err != nil {
		t.Fatal(err)
	}

	_, err = ts.svc.CreateProcesso(t.Context(), "dup-001")
	if !errors.Is(err, ErrProcessoExists) {
		t.Fatalf("expected ErrProcessoExists, got: %v", err)
	}
}

func TestGetProcesso(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)
	seeded := seedProcesso(t, ts.svc, "get-by-id")

	p, err := ts.svc.GetProcesso(t.Context(), seeded.ID)
	if err != nil {
		t.Fatal(err)
	}

	want := &Processo{
		ID:              seeded.ID,
		Numero:          "get-by-id",
		Status:          "PENDENTE",
		LinkAcesso:      "https://sei.example.com/processo/get-by-id",
		SeiUnidadeID:    "100",
		SeiUnidadeSigla: "SEPLAG/AP01",
	}
	if diff := cmp.Diff(want, p, ignoreProcessoFields); diff != "" {
		t.Fatalf("GetProcesso mismatch (-want +got):\n%s", diff)
	}
}

func TestGetProcesso_NotFound(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)

	_, err := ts.svc.GetProcesso(t.Context(), uuid.New())
	if !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestGetProcessoByNumero(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)
	seeded := seedProcesso(t, ts.svc, "get-by-numero")

	p, err := ts.svc.GetProcessoByNumero(t.Context(), "get-by-numero")
	if err != nil {
		t.Fatal(err)
	}

	want := &Processo{
		ID:              seeded.ID,
		Numero:          "get-by-numero",
		Status:          "PENDENTE",
		LinkAcesso:      "https://sei.example.com/processo/get-by-numero",
		SeiUnidadeID:    "100",
		SeiUnidadeSigla: "SEPLAG/AP01",
	}
	if diff := cmp.Diff(want, p, ignoreProcessoFields); diff != "" {
		t.Fatalf("GetProcessoByNumero mismatch (-want +got):\n%s", diff)
	}
}

func TestGetProcessoByNumero_NotFound(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)

	_, err := ts.svc.GetProcessoByNumero(t.Context(), "nonexistent")
	if !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestListProcessos(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)

	for i := range 3 {
		seedProcesso(t, ts.svc, fmt.Sprintf("list-proc-%03d", i+1))
	}

	result, err := ts.svc.ListProcessos(t.Context(), ListProcessosParams{Page: 1, Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Data) != 3 {
		t.Fatalf("expected len=3, got len=%d", len(result.Data))
	}
	if result.TotalCount != 3 {
		t.Fatalf("expected totalCount=3, got %d", result.TotalCount)
	}
	if result.HasNext {
		t.Fatal("expected hasNext=false for first page")
	}

	result, err = ts.svc.ListProcessos(t.Context(), ListProcessosParams{Numero: "002", Page: 1, Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected len=1, got len=%d", len(result.Data))
	}
	want := &Processo{
		Numero:          "list-proc-002",
		Status:          "PENDENTE",
		LinkAcesso:      "https://sei.example.com/processo/list-proc-002",
		SeiUnidadeID:    "100",
		SeiUnidadeSigla: "SEPLAG/AP01",
	}
	if diff := cmp.Diff(want, result.Data[0], ignoreProcessoIDFields); diff != "" {
		t.Fatalf("filtered processo mismatch (-want +got):\n%s", diff)
	}

	result, err = ts.svc.ListProcessos(t.Context(), ListProcessosParams{Numero: "nonexistent", Page: 1, Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Data) != 0 {
		t.Fatalf("expected len=0, got len=%d", len(result.Data))
	}
}

func TestListProcessos_Pagination(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)

	for i := range 5 {
		seedProcesso(t, ts.svc, fmt.Sprintf("paginated-proc-%03d", i+1))
	}

	result, err := ts.svc.ListProcessos(t.Context(), ListProcessosParams{Page: 1, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Data) != 2 {
		t.Fatalf("page 1: expected 2 items, got %d", len(result.Data))
	}
	if result.TotalCount != 5 {
		t.Fatalf("expected totalCount=5, got %d", result.TotalCount)
	}
	if result.TotalPages != 3 {
		t.Fatalf("expected totalPages=3, got %d", result.TotalPages)
	}
	if !result.HasNext {
		t.Fatal("page 1: expected hasNext=true")
	}
	if result.HasPrev {
		t.Fatal("page 1: expected hasPrev=false")
	}

	result, err = ts.svc.ListProcessos(t.Context(), ListProcessosParams{Page: 2, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Data) != 2 {
		t.Fatalf("page 2: expected 2 items, got %d", len(result.Data))
	}
	if !result.HasNext {
		t.Fatal("page 2: expected hasNext=true")
	}
	if !result.HasPrev {
		t.Fatal("page 2: expected hasPrev=true")
	}

	result, err = ts.svc.ListProcessos(t.Context(), ListProcessosParams{Page: 3, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("page 3: expected 1 item, got %d", len(result.Data))
	}
	if result.HasNext {
		t.Fatal("page 3: expected hasNext=false")
	}
	if !result.HasPrev {
		t.Fatal("page 3: expected hasPrev=true")
	}
}

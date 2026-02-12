package fila

import (
	"fmt"
	"testing"

	"github.com/automatiza-mg/fila/internal/database"
)

func seedProcessoAposentadoria(t *testing.T, store *database.Store, numero string, status database.StatusProcesso) *database.ProcessoAposentadoria {
	t.Helper()

	p, err := store.GetProcessoByNumero(t.Context(), numero)
	if err != nil {
		t.Fatal(err)
	}

	pa := &database.ProcessoAposentadoria{
		ProcessoID:               p.ID,
		DataRequerimento:         p.CriadoEm,
		CPFRequerente:            "123.456.789-00",
		DataNascimentoRequerente: p.CriadoEm,
		Invalidez:                false,
		Judicial:                 false,
		Prioridade:               false,
		Score:                    75,
		Status:                   status,
	}
	err = store.SaveProcessoAposentadoria(t.Context(), pa)
	if err != nil {
		t.Fatal(err)
	}
	return pa
}

func TestListProcesso(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)

	for i := range 3 {
		numero := fmt.Sprintf("list-ap-%03d", i+1)
		p := &database.Processo{
			Numero:              numero,
			StatusProcessamento: "PENDENTE",
			LinkAcesso:          "https://sei.example.com/processo/" + numero,
			SeiUnidadeID:        "100",
			SeiUnidadeSigla:     "SEPLAG/AP01",
		}
		err := ts.store.SaveProcesso(t.Context(), p)
		if err != nil {
			t.Fatal(err)
		}
		seedProcessoAposentadoria(t, ts.store, numero, database.StatusProcessoAnalisePendente)
	}

	result, err := ts.ListProcesso(t.Context(), ListProcessoAposentadoriaParams{Page: 1, Limit: 20})
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

	// Results are ordered by created_at DESC, so we expect them in reverse order
	for i, p := range result.Data {
		expectedNumero := fmt.Sprintf("list-ap-%03d", 3-i)
		if p.Numero != expectedNumero {
			t.Fatalf("expected numero=%s, got %s", expectedNumero, p.Numero)
		}
	}
}

func TestListProcesso_FilterByStatus(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)

	for i, status := range []database.StatusProcesso{
		database.StatusProcessoAnalisePendente,
		database.StatusProcessoEmAnalise,
		database.StatusProcessoAnalisePendente,
	} {
		numero := fmt.Sprintf("filter-ap-%03d", i+1)
		p := &database.Processo{
			Numero:              numero,
			StatusProcessamento: "PENDENTE",
			LinkAcesso:          "https://sei.example.com/processo/" + numero,
			SeiUnidadeID:        "100",
			SeiUnidadeSigla:     "SEPLAG/AP01",
		}
		err := ts.store.SaveProcesso(t.Context(), p)
		if err != nil {
			t.Fatal(err)
		}
		seedProcessoAposentadoria(t, ts.store, numero, status)
	}

	result, err := ts.ListProcesso(t.Context(), ListProcessoAposentadoriaParams{
		Status: string(database.StatusProcessoAnalisePendente),
		Page:   1,
		Limit:  20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Data) != 2 {
		t.Fatalf("expected len=2, got len=%d", len(result.Data))
	}
	if result.TotalCount != 2 {
		t.Fatalf("expected totalCount=2, got %d", result.TotalCount)
	}
}

func TestListProcesso_CaseInsensitiveStatus(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)

	numero := "case-ap-001"
	p := &database.Processo{
		Numero:              numero,
		StatusProcessamento: "PENDENTE",
		LinkAcesso:          "https://sei.example.com/processo/" + numero,
		SeiUnidadeID:        "100",
		SeiUnidadeSigla:     "SEPLAG/AP01",
	}
	err := ts.store.SaveProcesso(t.Context(), p)
	if err != nil {
		t.Fatal(err)
	}
	seedProcessoAposentadoria(t, ts.store, numero, database.StatusProcessoAnalisePendente)

	result, err := ts.ListProcesso(t.Context(), ListProcessoAposentadoriaParams{
		Status: "analise_pendente",
		Page:   1,
		Limit:  20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected len=1, got len=%d", len(result.Data))
	}
}

func TestListProcesso_Pagination(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)

	for i := range 5 {
		numero := fmt.Sprintf("paginated-ap-%03d", i+1)
		p := &database.Processo{
			Numero:              numero,
			StatusProcessamento: "PENDENTE",
			LinkAcesso:          "https://sei.example.com/processo/" + numero,
			SeiUnidadeID:        "100",
			SeiUnidadeSigla:     "SEPLAG/AP01",
		}
		err := ts.store.SaveProcesso(t.Context(), p)
		if err != nil {
			t.Fatal(err)
		}
		seedProcessoAposentadoria(t, ts.store, numero, database.StatusProcessoAnalisePendente)
	}

	result, err := ts.ListProcesso(t.Context(), ListProcessoAposentadoriaParams{Page: 1, Limit: 2})
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

	result, err = ts.ListProcesso(t.Context(), ListProcessoAposentadoriaParams{Page: 2, Limit: 2})
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

	result, err = ts.ListProcesso(t.Context(), ListProcessoAposentadoriaParams{Page: 3, Limit: 2})
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

func TestListProcesso_EmptyResult(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)

	result, err := ts.ListProcesso(t.Context(), ListProcessoAposentadoriaParams{Page: 1, Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Data) != 0 {
		t.Fatalf("expected len=0, got len=%d", len(result.Data))
	}
	if result.TotalCount != 0 {
		t.Fatalf("expected totalCount=0, got %d", result.TotalCount)
	}
	if result.TotalPages != 0 {
		t.Fatalf("expected totalPages=0, got %d", result.TotalPages)
	}
}

func TestListProcesso_IncludesNumero(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)

	numero := "numero-ap-001"
	p := &database.Processo{
		Numero:              numero,
		StatusProcessamento: "PENDENTE",
		LinkAcesso:          "https://sei.example.com/processo/" + numero,
		SeiUnidadeID:        "100",
		SeiUnidadeSigla:     "SEPLAG/AP01",
	}
	err := ts.store.SaveProcesso(t.Context(), p)
	if err != nil {
		t.Fatal(err)
	}
	seedProcessoAposentadoria(t, ts.store, numero, database.StatusProcessoAnalisePendente)

	result, err := ts.ListProcesso(t.Context(), ListProcessoAposentadoriaParams{Page: 1, Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected len=1, got len=%d", len(result.Data))
	}

	pAp := result.Data[0]
	if pAp.Numero != numero {
		t.Fatalf("expected numero=%s, got %s", numero, pAp.Numero)
	}
	if pAp.Status != string(database.StatusProcessoAnalisePendente) {
		t.Fatalf("expected status=%s, got %s", database.StatusProcessoAnalisePendente, pAp.Status)
	}
}

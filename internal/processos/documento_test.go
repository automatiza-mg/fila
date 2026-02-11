package processos

import (
	"encoding/json"
	"testing"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func seedDocumento(t *testing.T, svc *Service, proc *database.Processo, numero string, apiData sei.RetornoConsultaDocumento) *database.Documento {
	t.Helper()

	metadados, err := json.Marshal(apiData)
	if err != nil {
		t.Fatal(err)
	}

	d := &database.Documento{
		Numero:       numero,
		ProcessoID:   proc.ID,
		Tipo:         apiData.Serie.Nome,
		Unidade:      apiData.UnidadeElaboradora.Sigla,
		ContentType:  "application/pdf",
		OCR:          "conteudo do documento " + numero,
		LinkAcesso:   apiData.LinkAcesso,
		MetadadosAPI: metadados,
	}
	err = svc.store.SaveDocumento(t.Context(), d)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestListDocumentos(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)
	proc := seedProcesso(t, ts.svc, "doc-list-001")

	api1 := sei.RetornoConsultaDocumento{
		Data: "10/01/2026",
		Serie: sei.Serie{
			IdSerie: "1",
			Nome:    "Oficio",
		},
		UnidadeElaboradora: sei.UnidadeElaboradora{
			IdUnidade: "100",
			Sigla:     "SEPLAG/AP01",
		},
		LinkAcesso: "https://sei.example.com/doc/001",
		Assinaturas: sei.Assinaturas{
			Itens: []sei.Assinatura{
				{Nome: "Joao Silva", Sigla: "123.456.789-00"},
			},
		},
	}

	api2 := sei.RetornoConsultaDocumento{
		Data: "11/01/2026",
		Serie: sei.Serie{
			IdSerie: "2",
			Nome:    "Despacho",
		},
		UnidadeElaboradora: sei.UnidadeElaboradora{
			IdUnidade: "200",
			Sigla:     "SEPLAG/AP02",
		},
		LinkAcesso: "https://sei.example.com/doc/002",
		Assinaturas: sei.Assinaturas{
			Itens: []sei.Assinatura{
				{Nome: "Maria Souza", Sigla: "987.654.321-00"},
				{Nome: "Pedro Costa", Sigla: "111.222.333-44"},
			},
		},
	}

	seedDocumento(t, ts.svc, proc, "DOC-001", api1)
	seedDocumento(t, ts.svc, proc, "DOC-002", api2)

	docs, err := ts.svc.ListDocumentos(t.Context(), proc.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(docs) != 2 {
		t.Fatalf("expected 2 documentos, got %d", len(docs))
	}

	// Find doc by numero for deterministic assertions.
	byNumero := map[string]*Documento{}
	for _, d := range docs {
		byNumero[d.Numero] = d
	}

	ignore := cmpopts.IgnoreFields(Documento{}, "ID")

	wantDoc1 := &Documento{
		Numero:          "DOC-001",
		Tipo:            "Oficio",
		Conteudo:        "conteudo do documento DOC-001",
		LinkAcesso:      "https://sei.example.com/doc/001",
		Data:            "10/01/2026",
		UnidadeGeradora: "SEPLAG/AP01",
		Assinaturas: []Assinatura{
			{Nome: "Joao Silva", CPF: "123.456.789-00"},
		},
	}
	if diff := cmp.Diff(wantDoc1, byNumero["DOC-001"], ignore); diff != "" {
		t.Fatalf("DOC-001 mismatch (-want +got):\n%s", diff)
	}

	wantDoc2 := &Documento{
		Numero:          "DOC-002",
		Tipo:            "Despacho",
		Conteudo:        "conteudo do documento DOC-002",
		LinkAcesso:      "https://sei.example.com/doc/002",
		Data:            "11/01/2026",
		UnidadeGeradora: "SEPLAG/AP02",
		Assinaturas: []Assinatura{
			{Nome: "Maria Souza", CPF: "987.654.321-00"},
			{Nome: "Pedro Costa", CPF: "111.222.333-44"},
		},
	}
	if diff := cmp.Diff(wantDoc2, byNumero["DOC-002"], ignore); diff != "" {
		t.Fatalf("DOC-002 mismatch (-want +got):\n%s", diff)
	}
}

func TestListDocumentos_Empty(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)
	proc := seedProcesso(t, ts.svc, "doc-list-empty")

	docs, err := ts.svc.ListDocumentos(t.Context(), proc.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 0 {
		t.Fatalf("expected 0 documentos, got %d", len(docs))
	}
}

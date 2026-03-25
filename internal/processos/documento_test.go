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

	arq := &database.Arquivo{
		Hash:            "hash-" + numero,
		ChaveStorage:    "processos/hash-" + numero + ".pdf",
		ContentType:     "application/pdf",
		Conteudo:        "conteudo do documento " + numero,
		FormatoConteudo: "plain",
	}
	err := svc.store.SaveArquivo(t.Context(), arq)
	if err != nil {
		t.Fatal(err)
	}

	metadados, err := json.Marshal(apiData)
	if err != nil {
		t.Fatal(err)
	}

	d := &database.Documento{
		Numero:       numero,
		ProcessoID:   proc.ID,
		Tipo:         apiData.Serie.Nome,
		Unidade:      apiData.UnidadeElaboradora.Sigla,
		ArquivoHash:  arq.Hash,
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
		ContentType:     "application/pdf",
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
		ContentType:     "application/pdf",
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

func seedDocumentoComArquivo(t *testing.T, svc *Service, proc *database.Processo, numero string, apiData sei.RetornoConsultaDocumento, arq *database.Arquivo) *database.Documento {
	t.Helper()

	err := svc.store.SaveArquivo(t.Context(), arq)
	if err != nil {
		t.Fatal(err)
	}

	metadados, err := json.Marshal(apiData)
	if err != nil {
		t.Fatal(err)
	}

	d := &database.Documento{
		Numero:       numero,
		ProcessoID:   proc.ID,
		Tipo:         apiData.Serie.Nome,
		Unidade:      apiData.UnidadeElaboradora.Sigla,
		LinkAcesso:   apiData.LinkAcesso,
		ArquivoHash:  arq.Hash,
		MetadadosAPI: metadados,
	}
	err = svc.store.SaveDocumento(t.Context(), d)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestListDocumentos_ComArquivo(t *testing.T) {
	t.Parallel()

	ts := newTestService(t)
	proc := seedProcesso(t, ts.svc, "doc-list-arq-001")

	api1 := sei.RetornoConsultaDocumento{
		Data: "15/03/2026",
		Serie: sei.Serie{
			IdSerie: "3",
			Nome:    "Certidao",
		},
		UnidadeElaboradora: sei.UnidadeElaboradora{
			IdUnidade: "300",
			Sigla:     "SEPLAG/AP03",
		},
		LinkAcesso: "https://sei.example.com/doc/003",
		Assinaturas: sei.Assinaturas{
			Itens: []sei.Assinatura{
				{Nome: "Ana Lima", Sigla: "555.666.777-88"},
			},
		},
	}

	arq := &database.Arquivo{
		Hash:            "hash-arquivo-001",
		ChaveStorage:    "arquivos/hash-arquivo-001",
		ContentType:     "application/pdf",
		Conteudo:        "conteudo extraido do arquivo",
		FormatoConteudo: "plain",
	}

	seedDocumentoComArquivo(t, ts.svc, proc, "DOC-ARQ-001", api1, arq)

	// Cria um segundo documento no mesmo processo via seedDocumento.
	api2 := sei.RetornoConsultaDocumento{
		Data: "16/03/2026",
		Serie: sei.Serie{
			IdSerie: "4",
			Nome:    "Despacho",
		},
		UnidadeElaboradora: sei.UnidadeElaboradora{
			IdUnidade: "400",
			Sigla:     "SEPLAG/AP04",
		},
		LinkAcesso: "https://sei.example.com/doc/004",
		Assinaturas: sei.Assinaturas{
			Itens: []sei.Assinatura{
				{Nome: "Carlos Dias", Sigla: "999.888.777-66"},
			},
		},
	}
	seedDocumento(t, ts.svc, proc, "DOC-LEGACY-001", api2)

	docs, err := ts.svc.ListDocumentos(t.Context(), proc.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 2 {
		t.Fatalf("expected 2 documentos, got %d", len(docs))
	}

	byNumero := map[string]*Documento{}
	for _, d := range docs {
		byNumero[d.Numero] = d
	}

	ignore := cmpopts.IgnoreFields(Documento{}, "ID")

	// Documento com arquivo explícito: conteudo vem do Arquivo.Conteudo.
	wantArq := &Documento{
		Numero:          "DOC-ARQ-001",
		Tipo:            "Certidao",
		Conteudo:        "conteudo extraido do arquivo",
		LinkAcesso:      "https://sei.example.com/doc/003",
		ContentType:     "application/pdf",
		Data:            "15/03/2026",
		UnidadeGeradora: "SEPLAG/AP03",
		Assinaturas: []Assinatura{
			{Nome: "Ana Lima", CPF: "555.666.777-88"},
		},
	}
	if diff := cmp.Diff(wantArq, byNumero["DOC-ARQ-001"], ignore); diff != "" {
		t.Fatalf("DOC-ARQ-001 mismatch (-want +got):\n%s", diff)
	}

	// Segundo documento: conteudo vem do Arquivo.Conteudo criado pelo seedDocumento.
	wantDoc2 := &Documento{
		Numero:          "DOC-LEGACY-001",
		Tipo:            "Despacho",
		Conteudo:        "conteudo do documento DOC-LEGACY-001",
		LinkAcesso:      "https://sei.example.com/doc/004",
		ContentType:     "application/pdf",
		Data:            "16/03/2026",
		UnidadeGeradora: "SEPLAG/AP04",
		Assinaturas: []Assinatura{
			{Nome: "Carlos Dias", CPF: "999.888.777-66"},
		},
	}
	if diff := cmp.Diff(wantDoc2, byNumero["DOC-LEGACY-001"], ignore); diff != "" {
		t.Fatalf("DOC-LEGACY-001 mismatch (-want +got):\n%s", diff)
	}
}

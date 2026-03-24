package processos

import (
	"context"
	"fmt"
	"testing"

	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/postgres"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

var (
	ti *postgres.TestInstance

	_ SeiClient = (*fakeSeiClient)(nil)
)

type fakeSeiClient struct {
	consultarProcedimentoFn func(ctx context.Context, protocolo string) (*sei.ConsultarProcedimentoResponse, error)
	listarDocumentosFn      func(ctx context.Context, linkAcesso string) ([]sei.LinhaDocumento, error)
}

func (m *fakeSeiClient) ConsultarProcedimento(ctx context.Context, protocolo string) (*sei.ConsultarProcedimentoResponse, error) {
	if m.consultarProcedimentoFn != nil {
		return m.consultarProcedimentoFn(ctx, protocolo)
	}
	return nil, fmt.Errorf("ConsultarProcedimento not implemented")
}

func (m *fakeSeiClient) ListarDocumentos(ctx context.Context, linkAcesso string) ([]sei.LinhaDocumento, error) {
	if m.listarDocumentosFn != nil {
		return m.listarDocumentosFn(ctx, linkAcesso)
	}
	return nil, fmt.Errorf("ListarDocumentos not implemented")
}

type fakeHook struct {
	called     bool
	processo   *Processo
	documentos []*Documento
}

func (m *fakeHook) OnAnalyzeCompleteTx(_ context.Context, _ pgx.Tx, p *Processo, dd []*Documento) error {
	m.called = true
	m.processo = p
	m.documentos = dd
	return nil
}

type newTestServiceResult struct {
	svc *Service
	sei *fakeSeiClient
}

type fakeTaskInserter struct {
	//
}

func (ti *fakeTaskInserter) InsertTx(ctx context.Context, tx pgx.Tx, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	return nil, nil
}

func newTestService(t *testing.T) *newTestServiceResult {
	t.Helper()

	pool := ti.NewDatabase(t)
	seiTest := &fakeSeiClient{}

	svc := New(pool, seiTest, cache.NewMemoryCache(), &fakeTaskInserter{})

	return &newTestServiceResult{
		svc: svc,
		sei: seiTest,
	}
}

func seedProcesso(t *testing.T, svc *Service, numero string) *database.Processo {
	t.Helper()

	p := &database.Processo{
		Numero:              numero,
		StatusProcessamento: "PENDENTE",
		LinkAcesso:          "https://sei.example.com/processo/" + numero,
		SeiUnidadeID:        "100",
		SeiUnidadeSigla:     "SEPLAG/AP01",
	}
	err := svc.store.SaveProcesso(t.Context(), p)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestMain(m *testing.M) {
	ti = postgres.MustTestInstance()
	defer ti.Close()

	m.Run()
}

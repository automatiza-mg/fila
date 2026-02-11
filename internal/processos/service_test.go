package processos

import (
	"context"
	"fmt"
	"testing"

	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/postgres"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ti *postgres.TestInstance

	_ SeiClient        = (*testSeiClient)(nil)
	_ AnalyzeEnqueuer  = (*testEnqueuer)(nil)
	_ DocumentoFetcher = (*testFetcher)(nil)
	_ AnalyzeHook      = (*testHook)(nil)
)

type testSeiClient struct {
	consultarProcedimentoFn func(ctx context.Context, protocolo string) (*sei.ConsultarProcedimentoResponse, error)
	listarDocumentosFn      func(ctx context.Context, linkAcesso string) ([]sei.LinhaDocumento, error)
}

func (m *testSeiClient) ConsultarProcedimento(ctx context.Context, protocolo string) (*sei.ConsultarProcedimentoResponse, error) {
	if m.consultarProcedimentoFn != nil {
		return m.consultarProcedimentoFn(ctx, protocolo)
	}
	return nil, fmt.Errorf("ConsultarProcedimento not implemented")
}

func (m *testSeiClient) ListarDocumentos(ctx context.Context, linkAcesso string) ([]sei.LinhaDocumento, error) {
	if m.listarDocumentosFn != nil {
		return m.listarDocumentosFn(ctx, linkAcesso)
	}
	return nil, fmt.Errorf("ListarDocumentos not implemented")
}

type testEnqueuer struct {
	inserted bool
	err      error
}

func (m *testEnqueuer) EnqueueAnalyzeTx(_ context.Context, _ pgx.Tx, _ uuid.UUID) (bool, error) {
	return m.inserted, m.err
}

type testFetcher struct {
	docs []DocumentoSei
	err  error
}

func (m *testFetcher) FetchDocumentos(_ context.Context, _ []string) ([]DocumentoSei, error) {
	return m.docs, m.err
}

type testHook struct {
	called     bool
	processo   *Processo
	documentos []*Documento
}

func (m *testHook) OnAnalyzeComplete(_ context.Context, p *Processo, dd []*Documento) error {
	m.called = true
	m.processo = p
	m.documentos = dd
	return nil
}

type newTestServiceResult struct {
	svc     *Service
	sei     *testSeiClient
	queue   *testEnqueuer
	fetcher *testFetcher
}

func newTestService(t *testing.T) *newTestServiceResult {
	t.Helper()

	pool := ti.NewDatabase(t)
	seiTest := &testSeiClient{}
	queue := &testEnqueuer{inserted: true}
	fetcher := &testFetcher{}

	svc := New(&ServiceOpts{
		Pool:    pool,
		Sei:     seiTest,
		Cache:   cache.NewMemoryCache(),
		Queue:   queue,
		Fetcher: fetcher,
	})

	return &newTestServiceResult{
		svc:     svc,
		sei:     seiTest,
		queue:   queue,
		fetcher: fetcher,
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

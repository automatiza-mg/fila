package fila

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/automatiza-mg/fila/internal/aposentadoria"
	"github.com/automatiza-mg/fila/internal/datalake"
	"github.com/automatiza-mg/fila/internal/postgres"
	"github.com/automatiza-mg/fila/internal/processos"
	"github.com/automatiza-mg/fila/internal/sei"
)

var (
	ti *postgres.TestInstance

	_ SeiClient             = (*fakeSei)(nil)
	_ AposentadoriaAnalyzer = (*fakeAnalyzer)(nil)
)

type fakeSei struct{}

func (s *fakeSei) ListarUnidades(ctx context.Context) (*sei.ListarUnidadesResponse, error) {
	unidades := make([]sei.Unidade, 20)
	for i := range unidades {
		unidades[i] = sei.Unidade{
			IdUnidade: strconv.Itoa(i + 1),
			Sigla:     fmt.Sprintf("SEPLAG/AP%02d", i+1),
		}
	}

	unidades = append(unidades, sei.Unidade{
		IdUnidade: "TESTE",
		Sigla:     "TESTE",
	})

	return &sei.ListarUnidadesResponse{
		Parametros: sei.Parametros[sei.Unidade]{
			Items: unidades,
		},
	}, nil
}

type fakeServidores struct {
	//
}

func (f *fakeServidores) GetServidorByCPF(ctx context.Context, cpf string) (*datalake.Servidor, error) {
	return &datalake.Servidor{
		CPF:            cpf,
		DataNascimento: time.Now().AddDate(-50, 0, 0),
	}, nil
}

type fakeAnalyzer struct {
	analise aposentadoria.Analise
}

func (f *fakeAnalyzer) AnalisarAposentadoria(ctx context.Context, docs []*processos.Documento) (*aposentadoria.Analise, error) {
	return &f.analise, nil
}

func TestMain(m *testing.M) {
	ti = postgres.MustTestInstance()
	defer ti.Close()

	m.Run()
}

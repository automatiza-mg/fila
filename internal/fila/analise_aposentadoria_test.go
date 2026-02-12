package fila

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/automatiza-mg/fila/internal/aposentadoria"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/processos"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

type testAnalyzer struct {
	analise *aposentadoria.Analise
	err     error
}

func (ta *testAnalyzer) AnalisarAposentadoria(ctx context.Context, docs []*processos.Documento) (*aposentadoria.Analise, error) {
	return ta.analise, ta.err
}

func newTestAnalyzer(analise *aposentadoria.Analise, err error) DocumentAnalyzer {
	return &testAnalyzer{analise: analise, err: err}
}

func seedProcesso(t *testing.T, svc *Service) *processos.Processo {
	t.Helper()

	p := &database.Processo{
		Numero:              "0000001-12.2024.1.00.0000",
		StatusProcessamento: "NOVO",
		LinkAcesso:          "https://example.com/processo",
		SeiUnidadeID:        "1",
		SeiUnidadeSigla:     "SEPLAG",
	}

	err := svc.store.SaveProcesso(t.Context(), p)
	if err != nil {
		t.Fatal(err)
	}

	return &processos.Processo{
		ID:           p.ID,
		Numero:       p.Numero,
		Status:       p.StatusProcessamento,
		LinkAcesso:   p.LinkAcesso,
		SeiUnidadeID: p.SeiUnidadeID,
	}
}

func TestAnalyzeAposentadoria(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupAnalyzer func() DocumentAnalyzer
		setupProcesso func(*testing.T, *Service) (*processos.Processo, *database.Processo)
		expectError   bool
		checkFunc     func(*testing.T, *Service, *database.Processo, *aposentadoria.Analise)
	}{
		{
			name: "success with retirement case",
			setupAnalyzer: func() DocumentAnalyzer {
				return newTestAnalyzer(&aposentadoria.Analise{
					Aposentadoria:    true,
					CPF:              "12345678900",
					DataNascimento:   "1970-01-01",
					DataRequerimento: "2024-01-01",
					Judicial:         false,
					Invalidez:        false,
				}, nil)
			},
			setupProcesso: func(t *testing.T, svc *Service) (*processos.Processo, *database.Processo) {
				proc := seedProcesso(t, svc)
				dbProc, err := svc.store.GetProcesso(t.Context(), proc.ID)
				if err != nil {
					t.Fatal(err)
				}
				return proc, dbProc
			},
			expectError: false,
			checkFunc: func(t *testing.T, svc *Service, dbProc *database.Processo, analise *aposentadoria.Analise) {
				// Verify Processo updated with metadata
				updated, err := svc.store.GetProcesso(t.Context(), dbProc.ID)
				if err != nil {
					t.Fatal(err)
				}
				if !updated.Aposentadoria.Valid || !updated.Aposentadoria.V {
					t.Error("expected Aposentadoria to be true")
				}
				if !updated.AnalisadoEm.Valid {
					t.Error("expected AnalisadoEm to be set")
				}
				if updated.MetadadosIA == nil {
					t.Error("expected MetadadosIA to be set")
				}

				// Verify ProcessoAposentadoria created
				pa, err := svc.store.GetProcessoAposentadoriaByNumero(t.Context(), dbProc.Numero)
				if err != nil {
					t.Fatal(err)
				}
				if pa.CPFRequerente != analise.CPF {
					t.Errorf("expected CPF %s, got %s", analise.CPF, pa.CPFRequerente)
				}
				if pa.Judicial != analise.Judicial {
					t.Errorf("expected Judicial %v, got %v", analise.Judicial, pa.Judicial)
				}
				if pa.Invalidez != analise.Invalidez {
					t.Errorf("expected Invalidez %v, got %v", analise.Invalidez, pa.Invalidez)
				}
				if pa.Status != database.StatusProcessoAnalisePendente {
					t.Errorf("expected status %s, got %s", database.StatusProcessoAnalisePendente, pa.Status)
				}
			},
		},
		{
			name: "success without retirement case",
			setupAnalyzer: func() DocumentAnalyzer {
				return newTestAnalyzer(&aposentadoria.Analise{
					Aposentadoria:    false,
					CPF:              "12345678900",
					DataNascimento:   "1970-01-01",
					DataRequerimento: "2024-01-01",
					Judicial:         false,
					Invalidez:        false,
				}, nil)
			},
			setupProcesso: func(t *testing.T, svc *Service) (*processos.Processo, *database.Processo) {
				proc := seedProcesso(t, svc)
				dbProc, err := svc.store.GetProcesso(t.Context(), proc.ID)
				if err != nil {
					t.Fatal(err)
				}
				return proc, dbProc
			},
			expectError: false,
			checkFunc: func(t *testing.T, svc *Service, dbProc *database.Processo, analise *aposentadoria.Analise) {
				// Verify Processo updated
				updated, err := svc.store.GetProcesso(t.Context(), dbProc.ID)
				if err != nil {
					t.Fatal(err)
				}
				if updated.MetadadosIA == nil {
					t.Error("expected MetadadosIA to be set")
				}

				// Verify NO ProcessoAposentadoria created
				_, err = svc.store.GetProcessoAposentadoriaByNumero(t.Context(), dbProc.Numero)
				if !errors.Is(err, database.ErrNotFound) {
					t.Error("expected ProcessoAposentadoria to not exist")
				}
			},
		},
		{
			name: "processo not found",
			setupAnalyzer: func() DocumentAnalyzer {
				return newTestAnalyzer(&aposentadoria.Analise{}, nil)
			},
			setupProcesso: func(t *testing.T, svc *Service) (*processos.Processo, *database.Processo) {
				return &processos.Processo{ID: uuid.New()}, nil
			},
			expectError: true,
		},
		{
			name: "analyzer error",
			setupAnalyzer: func() DocumentAnalyzer {
				return newTestAnalyzer(nil, errors.New("analyzer failed"))
			},
			setupProcesso: func(t *testing.T, svc *Service) (*processos.Processo, *database.Processo) {
				proc := seedProcesso(t, svc)
				dbProc, err := svc.store.GetProcesso(t.Context(), proc.ID)
				if err != nil {
					t.Fatal(err)
				}
				return proc, dbProc
			},
			expectError: true,
		},
		{
			name: "invalid birth date format",
			setupAnalyzer: func() DocumentAnalyzer {
				return newTestAnalyzer(&aposentadoria.Analise{
					Aposentadoria:    true,
					CPF:              "12345678900",
					DataNascimento:   "invalid-date",
					DataRequerimento: "2024-01-01",
					Judicial:         false,
					Invalidez:        false,
				}, nil)
			},
			setupProcesso: func(t *testing.T, svc *Service) (*processos.Processo, *database.Processo) {
				proc := seedProcesso(t, svc)
				dbProc, err := svc.store.GetProcesso(t.Context(), proc.ID)
				if err != nil {
					t.Fatal(err)
				}
				return proc, dbProc
			},
			expectError: true,
		},
		{
			name: "invalid requerimento date format",
			setupAnalyzer: func() DocumentAnalyzer {
				return newTestAnalyzer(&aposentadoria.Analise{
					Aposentadoria:    true,
					CPF:              "12345678900",
					DataNascimento:   "1970-01-01",
					DataRequerimento: "invalid-date",
					Judicial:         false,
					Invalidez:        false,
				}, nil)
			},
			setupProcesso: func(t *testing.T, svc *Service) (*processos.Processo, *database.Processo) {
				proc := seedProcesso(t, svc)
				dbProc, err := svc.store.GetProcesso(t.Context(), proc.ID)
				if err != nil {
					t.Fatal(err)
				}
				return proc, dbProc
			},
			expectError: true,
		},
		{
			name: "processo aposentadoria already exists",
			setupAnalyzer: func() DocumentAnalyzer {
				return newTestAnalyzer(&aposentadoria.Analise{
					Aposentadoria:    true,
					CPF:              "12345678900",
					DataNascimento:   "1970-01-01",
					DataRequerimento: "2024-01-01",
					Judicial:         false,
					Invalidez:        false,
				}, nil)
			},
			setupProcesso: func(t *testing.T, svc *Service) (*processos.Processo, *database.Processo) {
				proc := seedProcesso(t, svc)
				dbProc, err := svc.store.GetProcesso(t.Context(), proc.ID)
				if err != nil {
					t.Fatal(err)
				}
				// Pre-create ProcessoAposentadoria
				pa := &database.ProcessoAposentadoria{
					ProcessoID:               dbProc.ID,
					CPFRequerente:            "99999999999",
					DataNascimentoRequerente: time.Now(),
					DataRequerimento:         time.Now(),
					Status:                   database.StatusProcessoAnalisePendente,
				}
				err = svc.store.SaveProcessoAposentadoria(t.Context(), pa)
				if err != nil {
					t.Fatal(err)
				}
				return proc, dbProc
			},
			expectError: false,
			checkFunc: func(t *testing.T, svc *Service, dbProc *database.Processo, analise *aposentadoria.Analise) {
				// Verify existing ProcessoAposentadoria not updated
				pa, err := svc.store.GetProcessoAposentadoriaByNumero(t.Context(), dbProc.Numero)
				if err != nil {
					t.Fatal(err)
				}
				if pa.CPFRequerente != "99999999999" {
					t.Error("expected existing ProcessoAposentadoria to remain unchanged")
				}
			},
		},
		{
			name: "with judicial and invalidez flags",
			setupAnalyzer: func() DocumentAnalyzer {
				return newTestAnalyzer(&aposentadoria.Analise{
					Aposentadoria:    true,
					CPF:              "98765432100",
					DataNascimento:   "1965-05-15",
					DataRequerimento: "2024-06-20",
					Judicial:         true,
					Invalidez:        true,
				}, nil)
			},
			setupProcesso: func(t *testing.T, svc *Service) (*processos.Processo, *database.Processo) {
				proc := seedProcesso(t, svc)
				dbProc, err := svc.store.GetProcesso(t.Context(), proc.ID)
				if err != nil {
					t.Fatal(err)
				}
				return proc, dbProc
			},
			expectError: false,
			checkFunc: func(t *testing.T, svc *Service, dbProc *database.Processo, analise *aposentadoria.Analise) {
				pa, err := svc.store.GetProcessoAposentadoriaByNumero(t.Context(), dbProc.Numero)
				if err != nil {
					t.Fatal(err)
				}
				if !pa.Judicial {
					t.Error("expected Judicial to be true")
				}
				if !pa.Invalidez {
					t.Error("expected Invalidez to be true")
				}
			},
		},
		{
			name: "metadata json marshaled correctly",
			setupAnalyzer: func() DocumentAnalyzer {
				return newTestAnalyzer(&aposentadoria.Analise{
					Aposentadoria:    true,
					CPF:              "12345678900",
					DataNascimento:   "1970-01-01",
					DataRequerimento: "2024-01-01",
					Judicial:         false,
					Invalidez:        false,
					CPFDiligencia:    "98765432100",
				}, nil)
			},
			setupProcesso: func(t *testing.T, svc *Service) (*processos.Processo, *database.Processo) {
				proc := seedProcesso(t, svc)
				dbProc, err := svc.store.GetProcesso(t.Context(), proc.ID)
				if err != nil {
					t.Fatal(err)
				}
				return proc, dbProc
			},
			expectError: false,
			checkFunc: func(t *testing.T, svc *Service, dbProc *database.Processo, analise *aposentadoria.Analise) {
				updated, err := svc.store.GetProcesso(t.Context(), dbProc.ID)
				if err != nil {
					t.Fatal(err)
				}

				var unmarshaled aposentadoria.Analise
				err = json.Unmarshal(updated.MetadadosIA, &unmarshaled)
				if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(analise, &unmarshaled); diff != "" {
					t.Fatalf("metadata mismatch:\n%s", diff)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t)
			svc.analyzer = tt.setupAnalyzer()

			proc, dbProc := tt.setupProcesso(t, svc)
			var analise *aposentadoria.Analise
			if tt.setupAnalyzer() != nil {
				if ta, ok := tt.setupAnalyzer().(*testAnalyzer); ok {
					analise = ta.analise
				}
			}

			err := svc.AnalyzeAposentadoria(t.Context(), proc, nil)
			if (err != nil) != tt.expectError {
				t.Fatalf("expectError %v, got error: %v", tt.expectError, err)
			}

			if !tt.expectError && tt.checkFunc != nil {
				tt.checkFunc(t, svc, dbProc, analise)
			}
		})
	}
}

func TestOnAnalyzeComplete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFunc   func(*testing.T, *Service) (*processos.Processo, []*processos.Documento)
		expectError bool
	}{
		{
			name: "success delegates to AnalyzeAposentadoria",
			setupFunc: func(t *testing.T, svc *Service) (*processos.Processo, []*processos.Documento) {
				svc.analyzer = newTestAnalyzer(&aposentadoria.Analise{
					Aposentadoria:    true,
					CPF:              "12345678900",
					DataNascimento:   "1970-01-01",
					DataRequerimento: "2024-01-01",
					Judicial:         false,
					Invalidez:        false,
				}, nil)

				proc := seedProcesso(t, svc)
				return proc, []*processos.Documento{}
			},
			expectError: false,
		},
		{
			name: "error propagated from AnalyzeAposentadoria",
			setupFunc: func(t *testing.T, svc *Service) (*processos.Processo, []*processos.Documento) {
				svc.analyzer = newTestAnalyzer(nil, errors.New("analyzer error"))

				proc := seedProcesso(t, svc)
				return proc, []*processos.Documento{}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t)
			proc, docs := tt.setupFunc(t, svc)

			err := svc.OnAnalyzeComplete(t.Context(), proc, docs)
			if (err != nil) != tt.expectError {
				t.Fatalf("expectError %v, got error: %v", tt.expectError, err)
			}
		})
	}
}

func TestAnalyzeAposentadoria_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupAnalyzer func() DocumentAnalyzer
		setupProcesso func(*testing.T, *Service) (*processos.Processo, *database.Processo)
		expectError   bool
		checkFunc     func(*testing.T, *Service, *database.Processo)
	}{
		{
			name: "empty documents list",
			setupAnalyzer: func() DocumentAnalyzer {
				return newTestAnalyzer(&aposentadoria.Analise{
					Aposentadoria:    true,
					CPF:              "12345678900",
					DataNascimento:   "1970-01-01",
					DataRequerimento: "2024-01-01",
					Judicial:         false,
					Invalidez:        false,
				}, nil)
			},
			setupProcesso: func(t *testing.T, svc *Service) (*processos.Processo, *database.Processo) {
				proc := seedProcesso(t, svc)
				dbProc, err := svc.store.GetProcesso(t.Context(), proc.ID)
				if err != nil {
					t.Fatal(err)
				}
				return proc, dbProc
			},
			expectError: false,
			checkFunc: func(t *testing.T, svc *Service, dbProc *database.Processo) {
				pa, err := svc.store.GetProcessoAposentadoriaByNumero(t.Context(), dbProc.Numero)
				if err != nil {
					t.Fatal(err)
				}
				if pa == nil {
					t.Error("expected ProcessoAposentadoria to be created")
				}
			},
		},
		{
			name: "cpf without formatting",
			setupAnalyzer: func() DocumentAnalyzer {
				return newTestAnalyzer(&aposentadoria.Analise{
					Aposentadoria:    true,
					CPF:              "12345678900",
					DataNascimento:   "1970-01-01",
					DataRequerimento: "2024-01-01",
					Judicial:         false,
					Invalidez:        false,
				}, nil)
			},
			setupProcesso: func(t *testing.T, svc *Service) (*processos.Processo, *database.Processo) {
				proc := seedProcesso(t, svc)
				dbProc, err := svc.store.GetProcesso(t.Context(), proc.ID)
				if err != nil {
					t.Fatal(err)
				}
				return proc, dbProc
			},
			expectError: false,
			checkFunc: func(t *testing.T, svc *Service, dbProc *database.Processo) {
				pa, err := svc.store.GetProcessoAposentadoriaByNumero(t.Context(), dbProc.Numero)
				if err != nil {
					t.Fatal(err)
				}
				if pa.CPFRequerente != "12345678900" {
					t.Errorf("expected CPF unchanged, got %s", pa.CPFRequerente)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t)
			svc.analyzer = tt.setupAnalyzer()

			proc, dbProc := tt.setupProcesso(t, svc)

			err := svc.AnalyzeAposentadoria(t.Context(), proc, nil)
			if (err != nil) != tt.expectError {
				t.Fatalf("expectError %v, got error: %v", tt.expectError, err)
			}

			if !tt.expectError && tt.checkFunc != nil {
				tt.checkFunc(t, svc, dbProc)
			}
		})
	}
}

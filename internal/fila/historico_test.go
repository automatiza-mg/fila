package fila

import (
	"testing"

	"github.com/automatiza-mg/fila/internal/database"
)

func seedHistorico(t *testing.T, store *database.Store, paID int64, statusAnterior *database.StatusProcesso, statusNovo database.StatusProcesso, observacao *string) *database.HistoricoStatusProcesso {
	t.Helper()

	h := &database.HistoricoStatusProcesso{
		ProcessoAposentadoriaID: paID,
		StatusAnterior:          database.Null(statusAnterior),
		StatusNovo:              statusNovo,
		Observacao:              database.Null(observacao),
	}
	err := store.SaveHistoricoStatusProcesso(t.Context(), h)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestListHistorico(t *testing.T) {
	tests := []struct {
		name        string
		numero      string
		setup       func(*testing.T, *Service) int64
		expectedLen int
		assertFn    func(*testing.T, []*HistoricoStatusProcesso)
	}{
		{
			name:   "single entry",
			numero: "historico-single-001",
			setup: func(t *testing.T, ts *Service) int64 {
				p := &database.Processo{
					Numero:              "historico-single-001",
					StatusProcessamento: "PENDENTE",
					LinkAcesso:          "https://sei.example.com/processo/historico-single-001",
					SeiUnidadeID:        "100",
					SeiUnidadeSigla:     "SEPLAG/AP01",
				}
				err := ts.store.SaveProcesso(t.Context(), p)
				if err != nil {
					t.Fatal(err)
				}

				pa := seedProcessoAposentadoria(t, ts.store, "historico-single-001", database.StatusProcessoAnalisePendente)
				seedHistorico(t, ts.store, pa.ID, nil, database.StatusProcessoAnalisePendente, ptr("Processo criado automaticamente através de análise de IA"))

				return pa.ID
			},
			expectedLen: 1,
			assertFn: func(t *testing.T, historico []*HistoricoStatusProcesso) {
				h := historico[0]
				if h.StatusNovo != string(database.StatusProcessoAnalisePendente) {
					t.Fatalf("expected statusNovo=%s, got %s", database.StatusProcessoAnalisePendente, h.StatusNovo)
				}
				if h.StatusAnterior != nil {
					t.Fatalf("expected statusAnterior=nil, got %v", h.StatusAnterior)
				}
				if h.UsuarioID != nil {
					t.Fatalf("expected usuarioID=nil, got %v", h.UsuarioID)
				}
				if h.Observacao == nil || *h.Observacao != "Processo criado automaticamente através de análise de IA" {
					t.Fatalf("expected observacao='Processo criado automaticamente...', got %v", h.Observacao)
				}
			},
		},
		{
			name:   "multiple entries",
			numero: "historico-multiple-001",
			setup: func(t *testing.T, ts *Service) int64 {
				p := &database.Processo{
					Numero:              "historico-multiple-001",
					StatusProcessamento: "PENDENTE",
					LinkAcesso:          "https://sei.example.com/processo/historico-multiple-001",
					SeiUnidadeID:        "100",
					SeiUnidadeSigla:     "SEPLAG/AP01",
				}
				err := ts.store.SaveProcesso(t.Context(), p)
				if err != nil {
					t.Fatal(err)
				}

				pa := seedProcessoAposentadoria(t, ts.store, "historico-multiple-001", database.StatusProcessoAnalisePendente)

				seedHistorico(t, ts.store, pa.ID, nil, database.StatusProcessoAnalisePendente, ptr("Criado"))
				seedHistorico(t, ts.store, pa.ID, ptr(database.StatusProcessoAnalisePendente), database.StatusProcessoEmAnalise, ptr("Em análise"))
				seedHistorico(t, ts.store, pa.ID, ptr(database.StatusProcessoEmAnalise), database.StatusProcessoConcluido, ptr("Concluído"))

				return pa.ID
			},
			expectedLen: 3,
			assertFn: func(t *testing.T, historico []*HistoricoStatusProcesso) {
				expectedStatuses := []string{
					string(database.StatusProcessoAnalisePendente),
					string(database.StatusProcessoEmAnalise),
					string(database.StatusProcessoConcluido),
				}

				for i, h := range historico {
					if h.StatusNovo != expectedStatuses[i] {
						t.Fatalf("entry %d: expected statusNovo=%s, got %s", i, expectedStatuses[i], h.StatusNovo)
					}
				}
			},
		},
		{
			name:   "empty",
			numero: "",
			setup: func(t *testing.T, ts *Service) int64 {
				return 9999
			},
			expectedLen: 0,
			assertFn: func(t *testing.T, historico []*HistoricoStatusProcesso) {
				// No additional assertions needed
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := newTestService(t)
			paID := tt.setup(t, ts)

			historico, err := ts.ListHistorico(t.Context(), paID)
			if err != nil {
				t.Fatal(err)
			}

			if len(historico) != tt.expectedLen {
				t.Fatalf("expected len=%d, got len=%d", tt.expectedLen, len(historico))
			}

			tt.assertFn(t, historico)
		})
	}
}

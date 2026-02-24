package fila

import (
	"database/sql"
	"testing"
	"time"

	"github.com/automatiza-mg/fila/internal/auth"
	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/google/uuid"
)

func TestGetProcessoAtribuido(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(*testing.T, *database.Store, *Service) int64
		expectError bool
		expectFound bool
	}{
		{
			name: "returns assigned process",
			setup: func(t *testing.T, store *database.Store, svc *Service) int64 {
				// Create analista user
				usuarioID := seedUsuario(t, store, auth.PapelAnalista)

				// Verify user
				usuario, err := store.GetUsuario(t.Context(), usuarioID)
				if err != nil {
					t.Fatal(err)
				}
				usuario.EmailVerificado = true
				if err := store.UpdateUsuario(t.Context(), usuario); err != nil {
					t.Fatal(err)
				}

				// Create analista
				analista := &database.Analista{
					UsuarioID:       usuarioID,
					Orgao:           "SEPLAG",
					SEIUnidadeID:    "001",
					SEIUnidadeSigla: "SEPLAG/AP00",
					Afastado:        false,
				}
				if err := store.SaveAnalista(t.Context(), analista); err != nil {
					t.Fatal(err)
				}

				// Create processo
				processo := &database.Processo{
					ID:     uuid.New(),
					Numero: "0000001-00.0000.0.00000.0000000-00",
				}
				if err := store.SaveProcesso(t.Context(), processo); err != nil {
					t.Fatal(err)
				}

				// Create processo aposentadoria with EM_ANALISE status
				pa := &database.ProcessoAposentadoria{
					ProcessoID:               processo.ID,
					DataRequerimento:         time.Now(),
					CPFRequerente:            "12345678900",
					DataNascimentoRequerente: time.Now().AddDate(-50, 0, 0),
					Invalidez:                false,
					Judicial:                 false,
					Prioridade:               false,
					Score:                    100,
					Status:                   database.StatusProcessoEmAnalise,
					AnalistaID:               sql.Null[int64]{Valid: true, V: usuarioID},
				}
				if err := store.SaveProcessoAposentadoria(t.Context(), pa); err != nil {
					t.Fatal(err)
				}

				return usuarioID
			},
			expectError: false,
			expectFound: true,
		},
		{
			name: "returns ErrNotFound when no process assigned",
			setup: func(t *testing.T, store *database.Store, svc *Service) int64 {
				// Create analista user without assigned process
				usuarioID := seedUsuario(t, store, auth.PapelAnalista)

				// Verify user
				usuario, err := store.GetUsuario(t.Context(), usuarioID)
				if err != nil {
					t.Fatal(err)
				}
				usuario.EmailVerificado = true
				if err := store.UpdateUsuario(t.Context(), usuario); err != nil {
					t.Fatal(err)
				}

				// Create analista
				analista := &database.Analista{
					UsuarioID:       usuarioID,
					Orgao:           "SEPLAG",
					SEIUnidadeID:    "001",
					SEIUnidadeSigla: "SEPLAG/AP00",
					Afastado:        false,
				}
				if err := store.SaveAnalista(t.Context(), analista); err != nil {
					t.Fatal(err)
				}

				return usuarioID
			},
			expectError: true,
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pool := ti.NewDatabase(t)
			store := database.New(pool)
			svc := &Service{
				pool:       pool,
				store:      store,
				sei:        &fakeSei{},
				cache:      cache.NewMemoryCache(),
				analyzer:   &fakeAnalyzer{},
				servidores: &fakeServidores{},
			}

			analistaID := tt.setup(t, store, svc)

			pa, err := svc.GetProcessoAtribuido(t.Context(), analistaID)

			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectFound && pa == nil {
				t.Errorf("expected processo, got nil")
			}
			if !tt.expectFound && pa != nil {
				t.Errorf("expected nil, got processo")
			}

			if tt.expectFound && pa != nil {
				if pa.Status != "EM_ANALISE" {
					t.Errorf("expected status EM_ANALISE, got %s", pa.Status)
				}
				if pa.AnalistaID == nil || *pa.AnalistaID != analistaID {
					t.Errorf("expected analista_id=%d, got %v", analistaID, pa.AnalistaID)
				}
				if pa.Numero != "0000001-00.0000.0.00000.0000000-00" {
					t.Errorf("expected numero, got %s", pa.Numero)
				}
			}
		})
	}
}

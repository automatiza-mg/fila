package fila

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/automatiza-mg/fila/internal/auth"
	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/google/uuid"
)

func TestAssignProcessoAposentadoria(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(*testing.T, *database.Store) (int64, int64)
		expectError bool
		expectNone  bool
	}{
		{
			name: "atribui processo com sucesso",
			setup: func(t *testing.T, store *database.Store) (int64, int64) {
				// Cria usuário e analista disponível
				usuarioID := seedUsuario(t, store, auth.PapelAnalista)

				// Marca o usuário como verificado
				usuario, err := store.GetUsuario(t.Context(), usuarioID)
				if err != nil {
					t.Fatal(err)
				}
				usuario.EmailVerificado = true
				if err := store.UpdateUsuario(t.Context(), usuario); err != nil {
					t.Fatal(err)
				}

				// Cria analista
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

				// Cria processo pendente
				processo := &database.Processo{
					ID:     uuid.New(),
					Numero: "0000001-00.0000.0.00000.0000000-00",
				}
				if err := store.SaveProcesso(t.Context(), processo); err != nil {
					t.Fatal(err)
				}

				pa := &database.ProcessoAposentadoria{
					ProcessoID:               processo.ID,
					DataRequerimento:         time.Now(),
					CPFRequerente:            "12345678900",
					DataNascimentoRequerente: time.Now().AddDate(-50, 0, 0),
					Invalidez:                false,
					Judicial:                 false,
					Prioridade:               false,
					Score:                    100,
					Status:                   database.StatusProcessoAnalisePendente,
				}
				if err := store.SaveProcessoAposentadoria(t.Context(), pa); err != nil {
					t.Fatal(err)
				}

				return pa.ID, usuarioID
			},
			expectError: false,
			expectNone:  false,
		},
		{
			name: "retorna nil quando nenhum analista disponível",
			setup: func(t *testing.T, store *database.Store) (int64, int64) {
				// Sem criar analista, apenas cria um processo
				processo := &database.Processo{
					ID:     uuid.New(),
					Numero: "0000002-00.0000.0.00000.0000000-00",
				}
				if err := store.SaveProcesso(t.Context(), processo); err != nil {
					t.Fatal(err)
				}

				pa := &database.ProcessoAposentadoria{
					ProcessoID:               processo.ID,
					DataRequerimento:         time.Now(),
					CPFRequerente:            "12345678901",
					DataNascimentoRequerente: time.Now().AddDate(-50, 0, 0),
					Invalidez:                false,
					Judicial:                 false,
					Prioridade:               false,
					Score:                    100,
					Status:                   database.StatusProcessoAnalisePendente,
				}
				if err := store.SaveProcessoAposentadoria(t.Context(), pa); err != nil {
					t.Fatal(err)
				}

				return pa.ID, 0
			},
			expectError: false,
			expectNone:  true,
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

			processoID, analistaID := tt.setup(t, store)

			err := svc.assignProcessoAposentadoria(t.Context())

			if (err != nil) != tt.expectError {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectNone {
				return
			}

			// Verifica se o processo foi atualizado
			processo, err := store.GetProcessoAposentadoria(t.Context(), processoID)
			if err != nil {
				t.Fatal(err)
			}

			if !processo.AnalistaID.Valid {
				t.Errorf("esperava processo.AnalistaID.Valid = true, got false")
			}

			if processo.Status != database.StatusProcessoEmAnalise {
				t.Errorf("esperava status = %q, got %q", database.StatusProcessoEmAnalise, processo.Status)
			}

			// Verifica se o timestamp do analista foi atualizado
			analista, err := store.GetAnalista(t.Context(), analistaID)
			if err != nil {
				t.Fatal(err)
			}

			if !analista.UltimaAtribuicaoEm.Valid {
				t.Errorf("esperava analista.UltimaAtribuicaoEm.Valid = true, got false")
			}

			// Timestamp deve estar próximo a agora (dentro de 1 segundo)
			if analista.UltimaAtribuicaoEm.Valid {
				diff := time.Since(analista.UltimaAtribuicaoEm.V)
				if diff > 1*time.Second {
					t.Errorf("timestamp muito antigo: %v", diff)
				}
			}
		})
	}
}

func TestStartAssignmentWorker(t *testing.T) {
	t.Parallel()

	pool := ti.NewDatabase(t)
	svc := &Service{
		pool:       pool,
		store:      database.New(pool),
		sei:        &fakeSei{},
		cache:      cache.NewMemoryCache(),
		analyzer:   &fakeAnalyzer{},
		servidores: &fakeServidores{},
	}

	// Cria contexto que será cancelado
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Inicia o worker com intervalo curto
	svc.StartAssignmentWorker(ctx, 10*time.Millisecond)

	// Aguarda o contexto ser cancelado (timeout)
	<-ctx.Done()

	// Se chegou aqui sem panic, o worker foi iniciado corretamente
}

func TestAssignProcessoAposentadoriaPriority(t *testing.T) {
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

	// Cria usuário analista
	usuarioID := seedUsuario(t, store, auth.PapelAnalista)

	// Marca o usuário como verificado
	usuario, err := store.GetUsuario(t.Context(), usuarioID)
	if err != nil {
		t.Fatal(err)
	}
	usuario.EmailVerificado = true
	if err := store.UpdateUsuario(t.Context(), usuario); err != nil {
		t.Fatal(err)
	}

	// Cria analista
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

	// Cria dois processos: um em ANALISE_PENDENTE e outro em RETORNO_DILIGENCIA
	createProcess := func(status database.StatusProcesso) int64 {
		processo := &database.Processo{
			ID:     uuid.New(),
			Numero: uuid.New().String(),
		}
		if err := store.SaveProcesso(t.Context(), processo); err != nil {
			t.Fatal(err)
		}

		pa := &database.ProcessoAposentadoria{
			ProcessoID:               processo.ID,
			DataRequerimento:         time.Now(),
			CPFRequerente:            uuid.New().String()[:11],
			DataNascimentoRequerente: time.Now().AddDate(-50, 0, 0),
			Invalidez:                false,
			Judicial:                 false,
			Prioridade:               false,
			Score:                    100,
			Status:                   status,
		}
		if err := store.SaveProcessoAposentadoria(t.Context(), pa); err != nil {
			t.Fatal(err)
		}
		return pa.ID
	}

	createProcess(database.StatusProcessoAnalisePendente)
	returnID := createProcess(database.StatusProcessoRetornoDiligencia)

	// Define o último analista do processo de retorno
	returndProc, _ := store.GetProcessoAposentadoria(t.Context(), returnID)
	returndProc.UltimoAnalistaID = sql.Null[int64]{Valid: true, V: usuarioID}
	store.UpdateProcessoAposentadoria(t.Context(), returndProc)

	// Executa atribuição
	err = svc.assignProcessoAposentadoria(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	// Verifica qual processo foi atribuído (deve ser o de RETORNO_DILIGENCIA com mesmo analista anterior)
	returnedProc, _ := store.GetProcessoAposentadoria(t.Context(), returnID)
	if returnedProc.AnalistaID.Valid && returnedProc.AnalistaID.V == usuarioID {
		// Sucesso: o processo de retorno com mesmo analista foi priorizado
		return
	}

	t.Errorf("esperava que o processo de RETORNO_DILIGENCIA fosse atribuído ao seu analista anterior")
}

package auth

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

const (
	CleanupTriggerDelete CleanupTrigger = iota
	CleanupTriggerPapelUpdate
)

// CleanupTrigger é a ação que causou o método Cleanup de um UsuarioHook
// ser chamado.
type CleanupTrigger int

func (t CleanupTrigger) String() string {
	switch t {
	case CleanupTriggerDelete:
		return "delete"
	case CleanupTriggerPapelUpdate:
		return "papel:update"
	default:
		return "unknown"
	}
}

// PendingAction representam ações pendentes de um usuários, como
// conclusão de cadastro, etc.
type PendingAction struct {
	Slug  string `json:"slug"`
	Title string `json:"titulo"`
}

// TODO: Alterar API de GetActions para enviar uma lista de usuários para carregar dados de forma
// mais eficiente.
type UsuarioHook interface {
	// Label retorna um ID do provider registrado.
	Label() string
	// GetActions retorna pendenciais específicas de um usuário.
	GetActions(ctx context.Context, u *Usuario) ([]PendingAction, error)
	// Cleanup executa ações de limpeza durante a exclusão ou alteração de papel
	// de um usuário. O campo Pendencias pode ou não estar carregado.
	Cleanup(ctx context.Context, tx pgx.Tx, trigger CleanupTrigger, usuario *Usuario) error
}

// Retorna as pendências de cadastro de um usuário.
func checkCoreActions(u *Usuario) []PendingAction {
	actions := make([]PendingAction, 0)

	if !u.EmailVerificado {
		actions = append(actions, PendingAction{
			Slug:  "concluir-cadastro",
			Title: "Concluir o cadastro da conta (verificar email e definir senha)",
		})
	}

	return actions
}

// Coleta as pendências de todos os providers registrados.
func (s *Service) getPendingActions(ctx context.Context, u *Usuario) []PendingAction {
	seen := make(map[string]struct{})

	actions := checkCoreActions(u)
	for _, a := range actions {
		seen[a.Slug] = struct{}{}
	}

	for _, h := range s.hooks {
		extraActions, err := h.GetActions(ctx, u)
		if err != nil {
			s.logger.Error(
				"Falha ao coletar pendências",
				slog.String("provider", h.Label()),
				slog.Any("err", err),
			)
			continue
		}

		for _, a := range extraActions {
			if _, ok := seen[a.Slug]; ok {
				continue
			}
			seen[a.Slug] = struct{}{}
			actions = append(actions, a)
		}
	}

	return actions
}

// Executa a limpeza de todos os providers registrados.
func (s *Service) cleanupAll(ctx context.Context, tx pgx.Tx, trigger CleanupTrigger, u *Usuario) error {
	s.logger.Debug(
		"Executando limpeza",
		slog.Int64("usuario_id", u.ID),
		slog.String("trigger", trigger.String()),
	)

	for _, h := range s.hooks {
		if err := h.Cleanup(ctx, tx, trigger, u); err != nil {
			return err
		}
	}

	return nil
}

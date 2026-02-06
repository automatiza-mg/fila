package auth

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

// PendingAction representam ações pendentes de um usuários, como
// conclusão de cadastro, etc.
type PendingAction struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
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

	for _, p := range s.providers {
		extraActions, err := p.GetActions(ctx, u)
		if err != nil {
			s.logger.Error(
				"Falha ao coletar pendências",
				slog.String("provider", p.Label()),
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
func (s *Service) cleanupAll(ctx context.Context, tx pgx.Tx, u *Usuario) error {
	for _, p := range s.providers {
		if err := p.Cleanup(ctx, tx, u); err != nil {
			return err
		}
	}
	return nil
}

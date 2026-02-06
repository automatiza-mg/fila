package auth

import (
	"context"
)

type PendingAction struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

type ActionProvider interface {
	GetActions(ctx context.Context, u *Usuario) ([]PendingAction, error)
}

// Retorna as pendências de cadastro de um usuário.
func (s *Service) checkCoreActions(u *Usuario) []PendingAction {
	actions := make([]PendingAction, 0)

	if !u.EmailVerificado {
		actions = append(actions, PendingAction{
			Slug:  "concluir-cadastro",
			Title: "Concluir o cadastro da conta (verificar email e definir senha)",
		})
	}

	return actions
}

// GetActions coleta as pendências de todos os [ActionProvider]'s registrados
// no serviço.
func (s *Service) GetActions(ctx context.Context, u *Usuario) ([]PendingAction, error) {
	actions := s.checkCoreActions(u)

	for _, p := range s.providers {
		extraActions, err := p.GetActions(ctx, u)
		if err != nil {
			return nil, err
		}
		if len(extraActions) > 0 {
			actions = append(actions, extraActions...)
		}
	}

	return actions, nil
}

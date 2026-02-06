package fila

import (
	"context"
	"errors"

	"github.com/automatiza-mg/fila/internal/auth"
	"github.com/automatiza-mg/fila/internal/database"
)

var _ auth.ActionProvider = (*Service)(nil)

// GetActions implementa a interface [auth.ActionProvider] para adicionar
// ações pendentes em usuários relacionados ao cadastro de dados de analista.
func (s *Service) GetActions(ctx context.Context, u *auth.Usuario) ([]auth.PendingAction, error) {
	if !u.IsAnalista() {
		return nil, nil
	}

	var actions []auth.PendingAction

	_, err := s.store.GetAnalista(ctx, u.ID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			actions = append(actions, auth.PendingAction{
				Slug:  "dados-analista",
				Title: "Registrar dados de analista",
			})
		} else {
			return nil, err
		}
	}

	return actions, nil
}

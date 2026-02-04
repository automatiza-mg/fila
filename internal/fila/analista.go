package fila

import (
	"context"
	"fmt"

	"github.com/automatiza-mg/fila/internal/database"
)

func (s *Service) DesativarAnalista(ctx context.Context, usuarioID int64) error {
	usuario, err := s.store.GetUsuario(ctx, usuarioID)
	if err != nil {
		return err
	}
	if !usuario.HasPapel(database.PapelAnalista) {
		return fmt.Errorf("invalid papel")
	}

	err = s.store.DeleteAnalista(ctx, usuarioID)
	if err != nil {
		return err
	}

	return nil
}

package main

import (
	"context"

	"github.com/automatiza-mg/fila/internal/database"
)

type contextKey int

const (
	usuarioContextKey contextKey = iota
)

func (app *application) getUsuario(ctx context.Context) *database.Usuario {
	usuario, ok := ctx.Value(usuarioContextKey).(*database.Usuario)
	if !ok {
		panic("usuario not present in context")
	}
	return usuario
}

func (app *application) setUsuario(ctx context.Context, usuario *database.Usuario) context.Context {
	return context.WithValue(ctx, usuarioContextKey, usuario)
}

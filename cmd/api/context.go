package main

import (
	"context"

	"github.com/automatiza-mg/fila/internal/auth"
)

type contextKey int

const (
	authContextKey contextKey = iota
	usuarioContextKey
)

// Retorna o usuário autenticado. Não confundir com o método getUsuario.
func (app *application) getAuth(ctx context.Context) *auth.Usuario {
	usuario, ok := ctx.Value(authContextKey).(*auth.Usuario)
	if !ok {
		panic("usuario not present in context")
	}
	return usuario
}

func (app *application) setAuth(ctx context.Context, usuario *auth.Usuario) context.Context {
	return context.WithValue(ctx, authContextKey, usuario)
}

// Retorna um usuário {usuarioID}. Não confundir com o método getAuth.
func (app *application) getUsuario(ctx context.Context) *auth.Usuario {
	usuario, ok := ctx.Value(usuarioContextKey).(*auth.Usuario)
	if !ok {
		panic("usuario not present in context")
	}
	return usuario
}

func (app *application) setUsuario(ctx context.Context, usuario *auth.Usuario) context.Context {
	return context.WithValue(ctx, usuarioContextKey, usuario)
}

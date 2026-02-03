package main

import (
	"errors"
	"net/http"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/validator"
	"github.com/justinas/nosurf"
)

type cadastrarPage struct {
	basePage
	CSRFToken string
	Token     string
	Usuario   *database.Usuario
	Form      cadastrarForm
}

func (app *application) handleCadastrar(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		app.serveErrorPage(w, r, http.StatusBadRequest, "O token de cadastro não foi informado")
		return
	}

	usuario, err := app.store.GetUsuarioForToken(r.Context(), token, database.EscopoSetup)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrInvalidToken):
			app.serveErrorPage(w, r, http.StatusBadRequest, "O token informado é inválido ou expirou")
		default:
			app.serverError(w, r, err)
		}
		return
	}

	app.servePage(w, r, http.StatusOK, "cadastrar.tmpl", cadastrarPage{
		basePage:  app.newBasePage(r, "Cadastrar"),
		CSRFToken: nosurf.Token(r),
		Token:     token,
		Usuario:   usuario,
	})
}

type cadastrarForm struct {
	Token         string `form:"token"`
	Senha         string `form:"senha"`
	ConfirmaSenha string `form:"confirmar_senha"`

	validator.Validator `form:"-"`
}

type cadastrarComponent struct {
	Form      cadastrarForm
	CSRFToken string
	Usuario   *database.Usuario
	Token     string
}

func (app *application) handleCadastrarPost(w http.ResponseWriter, r *http.Request) {
	var form cadastrarForm
	err := app.decodeForm(r, &form)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	ctx := r.Context()

	usuario, err := app.store.GetUsuarioForToken(ctx, form.Token, database.EscopoSetup)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrInvalidToken):
			app.serveErrorPage(w, r, http.StatusBadRequest, "O token informado é inválido ou expirou.")
		default:
			app.serverError(w, r, err)
		}
		return
	}

	form.Check(validator.StrongPassword(form.Senha), "senha", "Deve possuir pelo menos 8 caracteres, um dígito e um caractere especial")
	form.Check(form.Senha == form.ConfirmaSenha, "confirmar_senha", "Senhas devem ser idênticas")
	if !form.Valid() {
		app.serveComponent(w, r, http.StatusUnprocessableEntity, "cadastrar/form", cadastrarComponent{
			CSRFToken: nosurf.Token(r),
			Form:      form,
			Usuario:   usuario,
			Token:     form.Token,
		})
		return
	}

	// Atualiza os dados do usuário, finalizando o cadastro.
	usuario.EmailVerificado = true
	if err := usuario.SetSenha(form.Senha); err != nil {
		app.serverError(w, r, err)
		return
	}

	tx, err := app.pool.Begin(ctx)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	defer tx.Rollback(ctx)

	store := app.store.WithTx(tx)

	// Salva os dados do usuário
	if err := store.UpdateUsuario(ctx, usuario); err != nil {
		app.serverError(w, r, err)
		return
	}

	// Remove os tokens
	if err := store.DeleteTokensUsuario(ctx, usuario.ID, database.EscopoSetup); err != nil {
		app.serverError(w, r, err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		app.serverError(w, r, err)
		return
	}

	w.Header().Set("HX-Redirect", "/entrar")
}

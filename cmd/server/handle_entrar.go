package main

import (
	"net/http"

	"github.com/automatiza-mg/fila/internal/validator"
	"github.com/justinas/nosurf"
)

type entrarPage struct {
	basePage
	CSRFToken string
	Form      entrarForm
}

func (app *application) handleEntrar(w http.ResponseWriter, r *http.Request) {
	app.servePage(w, r, http.StatusOK, "entrar.tmpl", entrarPage{
		basePage:  app.newBasePage(r, "Entrar"),
		CSRFToken: nosurf.Token(r),
	})
}

type entrarForm struct {
	CPF   string `form:"cpf"`
	Senha string `form:"senha"`

	validator.Validator `form:"-"`
}

func (app *application) handleEntrarPost(w http.ResponseWriter, r *http.Request) {
	var form entrarForm
	err := app.decodeForm(r, &form)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}
}

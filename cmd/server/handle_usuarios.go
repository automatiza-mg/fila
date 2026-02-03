package main

import (
	"net/http"

	"github.com/automatiza-mg/fila/internal/database"
)

// Retorna as opções de papeis disponíveis para usuários.
func (app *application) papelOptions() []option {
	return []option{
		{Label: "Selecione um papel"},
		{Label: "Subsecretário(a)", Value: database.PapelSubsecretario},
		{Label: "Gestor(a)", Value: database.PapelGestor},
		{Label: "Analista", Value: database.PapelAnalista},
	}
}

type option struct {
	Label    string
	Value    string
	Selected bool
}

type usuarioCriarPage struct {
	basePage
	PapelOptions []option
	Form         usuarioCriarForm
}

func (app *application) handleUsuarioCriar(w http.ResponseWriter, r *http.Request) {
	app.servePage(w, r, http.StatusOK, "gestor/usuarios/criar.tmpl", usuarioCriarPage{
		basePage:     app.newBasePage(r, "Criar Usuário"),
		PapelOptions: app.papelOptions(),
	})
}

type usuarioCriarForm struct {
	Nome  string `form:"nome"`
	CPF   string `form:"cpf"`
	Email string `form:"email"`
	Papel string `form:"papel"`
}

type usuarioCriarComponent struct {
	PapelOptions []option
	Form         usuarioCriarForm
}

func (app *application) handleUsuarioCriarPost(w http.ResponseWriter, r *http.Request) {
	var form usuarioCriarForm
	err := app.decodeForm(r, &form)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// Marca papel como selecionado.
	papelOptions := app.papelOptions()
	for i := range papelOptions {
		if papelOptions[i].Value == form.Papel {
			papelOptions[i].Selected = true
		}
	}

	app.serveComponent(w, r, http.StatusOK, "usuarios/criar-form", usuarioCriarComponent{
		Form:         form,
		PapelOptions: papelOptions,
	})
}

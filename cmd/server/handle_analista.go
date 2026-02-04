package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/validator"
	"github.com/justinas/nosurf"
)

// Retorna a lista de opções de unidades do SEI para Analistas.
func (app *application) unidadeOptions(ctx context.Context) ([]option, error) {
	unidades, err := app.fila.ListUnidadesAnalistas(ctx)
	if err != nil {
		return nil, err
	}
	options := []option{
		{Label: "Selecione uma unidade"},
	}
	for _, unidade := range unidades {
		options = append(options, option{
			Label: unidade.Sigla,
			Value: unidade.IdUnidade,
		})
	}
	return options, nil
}

func orgaoOptions() []option {
	return []option{
		{Label: "Selecione um órgão"},
		{Label: "SEPLAG", Value: "SEPLAG"},
		{Label: "SEE", Value: "SEE"},
	}
}

type analistaCriarPage struct {
	basePage
	CSRFToken      string
	UnidadeOptions []option
	OrgaoOptions   []option
	Usuario        *database.Usuario
}

func (app *application) handleAnalistaCriar(w http.ResponseWriter, r *http.Request) {
	usuarioID, err := app.intParam(r, "usuarioID")
	if err != nil || usuarioID < 1 {
		app.notFound(w, r)
		return
	}

	usuario, err := app.store.GetUsuario(r.Context(), usuarioID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			app.notFound(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	if !usuario.HasPapel(database.PapelAnalista) {
		app.serveErrorPage(w, r, http.StatusBadRequest, "Usuário não possui o papel correto para possuir dados de analista.")
		return
	}

	orgaoOptions := orgaoOptions()
	unidadeOptions, err := app.unidadeOptions(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.servePage(w, r, http.StatusOK, "gestor/usuarios/analista.tmpl", analistaCriarPage{
		basePage:       app.newBasePage(r, "Cadastrar Analista"),
		Usuario:        usuario,
		UnidadeOptions: unidadeOptions,
		OrgaoOptions:   orgaoOptions,
		CSRFToken:      nosurf.Token(r),
	})
}

type analitaCriarForm struct {
	Orgao        string `form:"orgao"`
	SEIUnidadeID string `form:"sei_unidade_id"`

	validator.Validator `form:"-"`
}

func (app *application) handleAnalistaCriarPost(w http.ResponseWriter, r *http.Request) {
	usuarioID, err := app.intParam(r, "usuarioID")
	if err != nil || usuarioID < 1 {
		app.notFound(w, r)
		return
	}

	usuario, err := app.store.GetUsuario(r.Context(), usuarioID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			app.notFound(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	if !usuario.HasPapel(database.PapelAnalista) {
		app.serveErrorPage(w, r, http.StatusBadRequest, "Usuário não possui o papel correto para possuir dados de analista.")
		return
	}

	var form analitaCriarForm
	err = app.decodeForm(r, &form)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	orgaoOptions := orgaoOptions()

	allowedOrgaos := make([]string, 0, len(orgaoOptions))
	for i := range orgaoOptions {

		if orgaoOptions[i].Value != "" {
			allowedOrgaos = append(allowedOrgaos, orgaoOptions[i].Value)
		}
	}

	unidadeOptions, err := app.unidadeOptions(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	allowedUnidades := make([]string, 0, len(unidadeOptions))
	for i := range unidadeOptions {
		if unidadeOptions[i].Value != "" {
			allowedUnidades = append(allowedUnidades, unidadeOptions[i].Value)
		}
	}

	form.Check(validator.PermittedValue(form.SEIUnidadeID, allowedUnidades...), "sei_unidade_id", "Deve ser um valor válido")
	form.Check(validator.PermittedValue(form.Orgao, allowedOrgaos...), "orgao", "Deve ser um valor válido")
	if !form.Valid() {
		app.serveComponent(w, r, http.StatusUnprocessableEntity, "usuarios/analista-form", analistaCriarPage{
			CSRFToken:      nosurf.Token(r),
			UnidadeOptions: unidadeOptions,
			OrgaoOptions:   orgaoOptions,
			Usuario:        usuario,
		})
		return
	}

	err = app.store.SaveAnalista(r.Context(), &database.Analista{
		UsuarioID:    usuario.ID,
		Orgao:        form.Orgao,
		SEIUnidadeID: form.SEIUnidadeID,
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/gestor/usuarios/%d", usuario.ID))
}

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/mail"
	"github.com/automatiza-mg/fila/internal/validator"
	"github.com/justinas/nosurf"
)

type usuarioIndexPage struct {
	basePage
	Usuarios []*database.Usuario
}

func (app *application) handleUsuarioIndex(w http.ResponseWriter, r *http.Request) {
	usuarios, err := app.store.ListUsuarios(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.servePage(w, r, http.StatusOK, "gestor/usuarios/index.tmpl", usuarioIndexPage{
		basePage: app.newBasePage(r, "Usuários"),
		Usuarios: usuarios,
	})
}

// Retorna as opções de papeis disponíveis para usuários, excluindo ADMIN.
func papelOptions() []option {
	return []option{
		{Label: "Selecione um papel"},
		{Label: "Subsecretário(a)", Value: database.PapelSubsecretario},
		{Label: "Gestor(a)", Value: database.PapelGestor},
		{Label: "Analista", Value: database.PapelAnalista},
	}
}

type usuarioCriarPage struct {
	basePage
	CSRFToken    string
	PapelOptions []option
	Form         usuarioCriarForm
}

func (app *application) handleUsuarioCriar(w http.ResponseWriter, r *http.Request) {
	app.servePage(w, r, http.StatusOK, "gestor/usuarios/criar.tmpl", usuarioCriarPage{
		basePage:     app.newBasePage(r, "Criar Usuário"),
		CSRFToken:    nosurf.Token(r),
		PapelOptions: papelOptions(),
	})
}

type usuarioCriarForm struct {
	Nome  string `form:"nome"`
	CPF   string `form:"cpf"`
	Email string `form:"email"`
	Papel string `form:"papel"`

	validator.Validator `form:"-"`
}

type usuarioCriarComponent struct {
	PapelOptions []option
	CSRFToken    string
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
	papelOptions := papelOptions()
	for i := range papelOptions {
		if papelOptions[i].Value == form.Papel {
			papelOptions[i].Selected = true
		}
	}

	// Papeis permitidos para validação
	papeisAllowed := make([]string, 0, len(papelOptions))
	for _, opt := range papelOptions {
		if opt.Value != "" {
			papeisAllowed = append(papeisAllowed, opt.Value)
		}
	}

	form.Check(validator.NotBlank(form.Nome), "nome", "Campo obrigatório")
	form.Check(validator.NotBlank(form.CPF), "cpf", "Campo obrigatório")
	form.Check(validator.Matches(form.CPF, validator.CpfRX), "cpf", "Deve ser um CPF válido")
	form.Check(validator.NotBlank(form.Email), "email", "Campo obrigatório")
	form.Check(validator.Matches(form.Email, validator.EmailRX), "email", "Deve ser um email válido")
	form.Check(validator.PermittedValue(form.Papel, papeisAllowed...), "papel", "Deve ser um papel válido")
	if !form.Valid() {
		app.serveComponent(w, r, http.StatusUnprocessableEntity, "usuarios/criar-form", usuarioCriarComponent{
			Form:         form,
			CSRFToken:    nosurf.Token(r),
			PapelOptions: papelOptions,
		})
		return
	}

	usuario := &database.Usuario{
		Nome:  form.Nome,
		CPF:   form.CPF,
		Email: form.Email,
	}
	usuario.SetPapel(form.Papel)

	ctx := r.Context()
	tx, err := app.pool.Begin(ctx)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	defer tx.Rollback(ctx)

	store := app.store.WithTx(tx)

	// Salva usuário no banco de dados
	err = store.SaveUsuario(r.Context(), usuario)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrUsuarioCPFTaken):
			form.SetFieldError("cpf", "CPF em uso por outro usuário")
			app.serveComponent(w, r, http.StatusUnprocessableEntity, "usuarios/criar-form", usuarioCriarComponent{
				Form:         form,
				CSRFToken:    nosurf.Token(r),
				PapelOptions: papelOptions,
			})
		case errors.Is(err, database.ErrUsuarioEmailTaken):
			form.SetFieldError("email", "Email em uso por outro usuário")
			app.serveComponent(w, r, http.StatusUnprocessableEntity, "usuarios/criar-form", usuarioCriarComponent{
				Form:         form,
				CSRFToken:    nosurf.Token(r),
				PapelOptions: papelOptions,
			})
		default:
			app.serverError(w, r, err)
		}
		return
	}

	token, err := store.CreateToken(ctx, usuario.ID, database.EscopoSetup, 72*time.Hour)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// TODO: Enviar o email através de um worker.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		email, err := mail.NewSetupEmail(usuario.Email, mail.SetupEmailParams{
			SetupURL: fmt.Sprintf("%s/cadastrar?token=%s", app.cfg.BaseURL, url.QueryEscape(token.Plaintext)),
		})
		if err != nil {
			app.logger.Error("Não foi possível gerar email", slog.String("err", err.Error()))
			return
		}

		if err := app.mail.Send(ctx, email); err != nil {
			app.logger.Error("Não foi possível enviar email", slog.String("err", err.Error()))
		}
	}()

	err = tx.Commit(ctx)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/gestor/usuarios/%d", usuario.ID))
}

type usuarioDetalhePage struct {
	basePage
	Usuario  *database.Usuario
	Analista *database.Analista
}

func (app *application) handleUsuarioDetalhe(w http.ResponseWriter, r *http.Request) {
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

	var analista *database.Analista
	if usuario.HasPapel(database.PapelAnalista) {
		analista, err = app.store.GetAnalista(r.Context(), usuario.ID)
		if err != nil && !errors.Is(err, database.ErrNotFound) {
			app.serverError(w, r, err)
			return
		}
	}

	app.servePage(w, r, http.StatusOK, "gestor/usuarios/detalhe.tmpl", usuarioDetalhePage{
		basePage: app.newBasePage(r, usuario.Nome),
		Usuario:  usuario,
		Analista: analista,
	})
}

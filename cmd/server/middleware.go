package main

import (
	"net/http"

	"github.com/justinas/nosurf"
)

func (app *application) csrfProtection(next http.Handler) http.Handler {
	h := nosurf.New(next)
	h.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.serveErrorPage(w, r, http.StatusBadRequest, "Token de proteção CSRF não informado")
	}))
	h.SetBaseCookie(http.Cookie{
		HttpOnly: true,
	})
	return h
}

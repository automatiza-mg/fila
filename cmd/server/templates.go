package main

import (
	"bytes"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
)

// Renderiza uma a página HTML informada.
//
// O diretório `pages` é adicionado ao caminho da página informada. Por exemplo: `entrar.tmpl` => `pages/entrar.tmpl`.
func (app *application) servePage(w http.ResponseWriter, r *http.Request, status int, page string, data any) {
	patterns := []string{"base.tmpl", "shared/*.tmpl", filepath.Join("pages", page)}

	tmpl := template.New("base.tmpl").Option("missingkey=zero")
	_, err := tmpl.ParseFS(app.views, patterns...)
	if err != nil {
		app.logger.Error(
			"Não foi possível construir página",
			slog.String("err", err.Error()),
			slog.String("uri", r.URL.RequestURI()),
		)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, data)
	if err != nil {
		app.logger.Error(
			"Não foi possível executar página",
			slog.String("err", err.Error()),
			slog.String("uri", r.URL.RequestURI()),
		)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

func (app *application) serveComponent(w http.ResponseWriter, r *http.Request, status int, name string, data any) {
	patterns := []string{"base.tmpl", "shared/*.tmpl"}

	pages := make([]string, 0)
	err := fs.WalkDir(app.views, "pages", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".tmpl") {
			pages = append(pages, path)
		}
		return nil
	})
	if err != nil {
		app.logger.Error("Não foi possível encontrar páginas", slog.String("err", err.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	patterns = append(patterns, pages...)

	tmpl := template.New("base.tmpl").Option("missingkey=zero")
	_, err = tmpl.ParseFS(app.views, patterns...)
	if err != nil {
		app.logger.Error(
			"Não foi possível construir template",
			slog.String("err", err.Error()),
			slog.String("uri", r.URL.RequestURI()),
		)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(buf, name, data)
	if err != nil {
		app.logger.Error(
			"Não foi possível executar template",
			slog.String("err", err.Error()),
			slog.String("uri", r.URL.RequestURI()),
		)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

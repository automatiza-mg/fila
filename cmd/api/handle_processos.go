package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/processos"
	"github.com/automatiza-mg/fila/internal/sei"
	"golang.org/x/sync/errgroup"
)

func (app *application) handleProcessosDebugPrompt(w http.ResponseWriter, r *http.Request) {
	processo := r.URL.Query().Get("processo")

	ctx := r.Context()

	// Consulta o processo na API do SEI.
	resp, err := app.sei.ConsultarProcedimento(ctx, processo)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Lista os documentos na página do link de acesso.
	docs, err := app.sei.ListarDocumentos(ctx, resp.Parametros.LinkAcesso)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	g := new(errgroup.Group)
	g.SetLimit(5)

	documentos := make([]processos.Documento, len(docs))

	// Consulta os dados dos documentos na API do SEI.
	for i, doc := range docs {
		g.Go(func() error {
			key := fmt.Sprintf("fila:sei:documentos:%s", doc.Numero)
			b, err := app.cache.Remember(ctx, key, 12*time.Hour, func() ([]byte, error) {
				d, err := app.sei.ConsultarDocumento(ctx, doc.Numero)
				if err != nil {
					return nil, err
				}
				return json.Marshal(d)
			})

			if err != nil {
				return err
			}

			var d *sei.ConsultarDocumentoResponse
			err = json.Unmarshal(b, &d)
			if err != nil {
				return err
			}

			tipo := d.Parametros.Serie.Nome
			if d.Parametros.Numero != "" {
				tipo += " " + d.Parametros.Numero
			}

			documento := processos.Documento{
				Tipo:   tipo,
				Numero: d.Parametros.DocumentoFormatado,
			}

			for _, ass := range d.Parametros.Assinaturas.Itens {
				documento.Assinaturas = append(documento.Assinaturas, processos.Assinatura{
					CPF:  ass.Sigla,
					Nome: ass.Nome,
				})
			}

			documentos[i] = documento
			return nil
		})
	}

	// Espera as consultas terminarem
	if err := g.Wait(); err != nil {
		app.serverError(w, r, err)
		return
	}

	usuarios, _, err := app.store.ListUsuarios(ctx, database.ListUsuariosParams{
		Papel: database.PapelAnalista,
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	analistas := make([]processos.Analista, len(usuarios))
	for i, usuario := range usuarios {
		cpf := strings.ReplaceAll(usuario.CPF, ".", "")
		cpf = strings.ReplaceAll(cpf, "-", "")

		analistas[i] = processos.Analista{
			ID:   usuario.ID,
			CPF:  cpf,
			Nome: usuario.Nome,
		}
	}

	prompt, err := processos.NewPrompt(documentos, analistas)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, prompt)
}

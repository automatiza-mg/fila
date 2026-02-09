package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type ErrorResponse struct {
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors,omitempty"`
}

func (app *application) writeError(w http.ResponseWriter, status int, msg string) {
	app.writeJSON(w, status, ErrorResponse{
		Message: msg,
	})
}

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Error("Erro interno do servidor", slog.String("err", err.Error()), slog.String("uri", r.URL.RequestURI()))
	app.writeJSON(w, http.StatusInternalServerError, ErrorResponse{
		Message: "O servidor encontrou um erro inesperado ao processar sua requisição",
	})
}

func (app *application) notFound(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusNotFound, ErrorResponse{
		Message: "O recurso solicitado não foi encontrado",
	})
}

func (app *application) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{
		Message: fmt.Sprintf("O método %q não é permitido para esse recurso", r.Method),
	})
}

func (app *application) validationFailed(w http.ResponseWriter, _ *http.Request, errors map[string]string) {
	app.writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
		Message: "A validação dos dados falhou",
		Errors:  errors,
	})
}

func (app *application) decodeError(w http.ResponseWriter, _ *http.Request, err error) {
	status := http.StatusBadRequest
	msg := "A requisição é inválida ou malformada"

	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError
	var maxBytesError *http.MaxBytesError

	if errors.As(err, &syntaxError) {
		msg = fmt.Sprintf("O corpo contém JSON inválido (na posição %d)", syntaxError.Offset)
	}
	if errors.As(err, &unmarshalTypeError) {
		if unmarshalTypeError.Field != "" {
			msg = fmt.Sprintf("O corpo contém tipo de JSON incorreto para o campo %q", unmarshalTypeError.Field)
		} else {
			msg = fmt.Sprintf("O corpo contém tipo de JSON incorreto (na posição %d)", unmarshalTypeError.Offset)
		}
	}
	if errors.As(err, &maxBytesError) {
		status = http.StatusRequestEntityTooLarge
		msg = fmt.Sprintf("O corpo da requisição excedeu o limite de %d bytes", maxBytesError.Limit)
	}
	if field, ok := strings.CutPrefix(err.Error(), "json: unknown field "); ok {
		msg = fmt.Sprintf("O corpo contém um campo desconhecido: %s", field)
	}
	if errors.Is(err, errMultipleJSONValues) {
		msg = "O corpo da requisição deve conter apenas um valor JSON"
	}
	if err == io.EOF {
		msg = "O corpo da requisição não pode estar vazio"
	}

	app.writeJSON(w, status, ErrorResponse{
		Message: msg,
	})
}

func (app *application) tokenError(w http.ResponseWriter, _ *http.Request) {
	app.writeJSON(w, http.StatusUnauthorized, ErrorResponse{
		Message: "O token informado é inválido ou expirou.",
	})
}

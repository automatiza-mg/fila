package main

import "net/http"

func (app *application) decodeForm(r *http.Request, v any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}
	return app.decoder.Decode(v, r.Form)
}

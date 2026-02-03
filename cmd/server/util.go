package main

import (
	"net/http"
	"strconv"
)

func (app *application) decodeForm(r *http.Request, v any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}
	return app.decoder.Decode(v, r.Form)
}

func (app *application) intParam(r *http.Request, key string) (int64, error) {
	v, err := strconv.ParseInt(r.PathValue(key), 10, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}

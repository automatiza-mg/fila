package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

const (
	maxBodySize = 5 << 20
)

func (app *application) writeJSON(w http.ResponseWriter, status int, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
}

func (app *application) decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func (app *application) intParam(r *http.Request, key string) (int64, error) {
	v, err := strconv.ParseInt(r.PathValue(key), 10, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}

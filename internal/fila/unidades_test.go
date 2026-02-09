package fila

import (
	"strings"
	"testing"
)

func TestListarUnidadesAnalista(t *testing.T) {
	t.Parallel()

	fila := newTestService(t)

	unidades, err := fila.ListUnidadesAnalistas(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, unidade := range unidades {
		if !strings.HasPrefix(unidade.Sigla, "SEPLAG/AP") {
			t.Fatalf("expected %q to have SEPLAG/AP prefix", unidade.Sigla)
		}
	}
}

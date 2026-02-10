package llm

import (
	"context"

	"github.com/automatiza-mg/fila/internal/processos"
)

type AnaliseProcesso struct {
	Aposentadoria bool   `json:"aposentadoria" jsonschema:"required,description=Indica se o processo é ou não um pedido de aposentadoria"`
	CPF           string `json:"cpf" jsonschema:"required,description=O CPF do requerente"`
	Judicial      bool   `json:"judicial" jsonschema:"required,description=Indica se houve pedido judicial para dar início ao processo"`
	Invalidez     bool   `json:"invalidez" jsonschema:"required,description=Indica se o requerente abriu o processo por invalidez"`
}

func (c *Client) AnalisarProcesso(ctx context.Context, p *processos.Processo) (*AnaliseProcesso, error) {
	return nil, nil
}

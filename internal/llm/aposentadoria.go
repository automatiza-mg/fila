package llm

import (
	"context"
	"encoding/json"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

type AnaliseAposentadoria struct {
	Aposentadoria    bool   `json:"aposentadoria" jsonschema:"required" jsonschema_description:"Indica se o processo é ou não um pedido de aposentadoria"`
	CPF              string `json:"cpf_requerente" jsonschema:"required" jsonschema_description:"O CPF do requerente da aposentadoria, sem pontos e traços"`
	DataRequerimento string `json:"data_requerimento" jsonschema:"required,format=date" jsonschema_description:"A data em que o requerimento foi enviado, no formato YYYY-MM-DD"`
	DataNascimento   string `json:"data_nascimento_requerente" jsonschema:"required,format=date" jsonschema_description:"A data de nascimento do requerente, no formato YYYY-MM-DD"`
	Judicial         bool   `json:"judicial" jsonschema:"required" jsonschema_description:"Indica se houve pedido judicial para dar início ao processo"`
	Invalidez        bool   `json:"invalidez" jsonschema:"required" jsonschema_description:"Indica se o requerente abriu o processo por invalidez"`
	CPFDiligencia    string `json:"cpf_responsavel_diligencia" jsonschema:"not_required" jsonschema_description:"O CPF do responsável pelo envio da diligência, se houver, sem pontos e traços"`
}

type Assinatura struct {
	Nome string
	CPF  string
}

type Documento struct {
	Tipo        string
	Data        string
	Conteudo    string
	Assinaturas []Assinatura
}

// AnalisarAposentadoria faz o uso de Inteligência Artificial para analisar
// uma lista de documentos para gerar um análise indicando os dados de
// aposentadoria de um processo.
func (c *Client) AnalisarAposentadoria(ctx context.Context, docs []Documento) (*AnaliseAposentadoria, error) {
	prompt, err := NewAposentadoriaPrompt(AposentadoriaPromptParams{
		Documentos: docs,
	})
	if err != nil {
		return nil, err
	}

	schema, err := GenerateMapSchema[AnaliseAposentadoria]()
	if err != nil {
		return nil, err
	}

	resp, err := c.openai.Responses.New(ctx, responses.ResponseNewParams{
		Model: openai.ChatModelGPT5_4,
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: prompt.Input(),
		},
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigParamOfJSONSchema("analise_aposentadoria", schema),
		},
	})
	if err != nil {
		return nil, err
	}

	var analise AnaliseAposentadoria
	err = json.Unmarshal([]byte(resp.Output[0].Content[0].Text), &analise)
	if err != nil {
		return nil, err
	}

	return &analise, nil
}

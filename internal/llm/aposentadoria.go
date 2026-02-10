package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"text/template"

	"github.com/automatiza-mg/fila/internal/aposentadoria"
	"github.com/automatiza-mg/fila/internal/processos"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

const promptAposentadoria = `
<tarefa>
Precisamos que sejam extraídas as seguintes informações do <documentos>: se este realmente é um processo de aposentadoria.
São exemplos de coisas que indicam que esse é um processo completo: requerimento de aposentadoria ou laudo de aposentadoria por invalidez ou aposentadoria compulsória,
contagem de tempo, cálculo de proventos, dados pessoais do servidor solicitante.
</tarefa>

<contexto>
Todo processo de aposentadoria deverá conter as seguintes informações:

1. Data de Nascimento do Requerente
2. CPF do Requerente

Caso encontre vários processos relativos a aposentadoria sem esse dado ou com vários nomes ao mesmo tempo,
deve se tratar de uma juntada de documento e não um processo de aposentadoria.

Você encontrará processos que contêm dados relativos a aposentadoria, mas que não são o processo como um todo,
por exemplo, somente documentos sobre cálculo de proventos, somente documentos com alguma retificação de dados etc,
o formato deve ser true para aposentadoria e false para o que não for aposentadoria;

A data do requerimento é a data em que o servidor solicita o processo de aposentadoria,
não confunda com a data de publicação; 

Se existe dentro do processo algum laudo médico indicando que se trata de aposentadoria por invalidez, 
retorne true caso exista e false caso não exista. Retorne sempre os dados mais recentes encontrados no processo.
Note que pequenas correções podem ocorrer ao longo do processo.
</contexto>

<documentos>
{{range .Documentos}}
<documento>
	Tipo: {{.Tipo}}
	Data: {{.Data}}
	Assinaturas:
		{{range .Assinaturas}}
		- {{.Nome}} ({{.CPF}})
		{{end}}
	Conteudo:
		{{.Conteudo}}
</documento>
{{end}}
</documentos>
`

// AnalisarAposentadoria faz o uso de Inteligência Artificial para analisar
// uma lista de documentos para gerar um análise indicando os dados de
// aposentadoria de um processo.
func (c *Client) AnalisarAposentadoria(ctx context.Context, docs []*processos.Documento) (*aposentadoria.Analise, error) {
	tmpl := template.Must(template.New("prompt").Parse(promptAposentadoria))
	buf := new(bytes.Buffer)
	err := tmpl.Execute(buf, map[string]any{
		"Documentos": docs,
	})
	if err != nil {
		return nil, err
	}

	schema, err := GenerateMapSchema[processos.AnaliseAposentadoria]()
	if err != nil {
		return nil, err
	}

	text := buf.String()

	resp, err := c.openai.Responses.New(ctx, responses.ResponseNewParams{
		Model: openai.ChatModelGPT5_2,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(text),
		},
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigParamOfJSONSchema("analise_aposentadoria", schema),
		},
	})
	if err != nil {
		return nil, err
	}

	var analise aposentadoria.Analise
	err = json.Unmarshal([]byte(resp.Output[0].Content[0].Text), &analise)
	if err != nil {
		return nil, err
	}

	return &analise, nil
}

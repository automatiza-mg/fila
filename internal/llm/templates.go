package llm

import (
	"bytes"
	"embed"
	"text/template"

	"github.com/openai/openai-go/v3/responses"
)

var (
	//go:embed prompts
	fs embed.FS

	aposentadoriaTmpl = template.Must(template.ParseFS(fs, "prompts/aposentadoria.tmpl"))
)

// Prompt representa um par de mensagens (system e user) renderizadas a partir
// de um template de prompt para ser enviado à API de Responses da OpenAI.
type Prompt struct {
	System string
	User   string
}

// Input converte o Prompt em uma lista de mensagens pronta para ser usada no
// campo Input de responses.ResponseNewParams.
func (p Prompt) Input() responses.ResponseInputParam {
	return responses.ResponseInputParam{
		responses.ResponseInputItemParamOfMessage(
			p.System,
			responses.EasyInputMessageRoleSystem,
		),
		responses.ResponseInputItemParamOfMessage(
			p.User,
			responses.EasyInputMessageRoleUser,
		),
	}
}

func executeTemplate(tmpl *template.Template, data any) (*Prompt, error) {
	systemBuf := new(bytes.Buffer)
	err := tmpl.ExecuteTemplate(systemBuf, "system", data)
	if err != nil {
		return nil, err
	}

	userBuf := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(userBuf, "user", data)
	if err != nil {
		return nil, err
	}

	return &Prompt{
		System: systemBuf.String(),
		User:   userBuf.String(),
	}, nil
}

// AposentadoriaPromptParams são os dados necessários para renderizar o
// prompt de análise de aposentadoria.
type AposentadoriaPromptParams struct {
	Documentos []Documento
}

// NewAposentadoriaPrompt renderiza o prompt de análise de aposentadoria a
// partir dos parâmetros informados.
func NewAposentadoriaPrompt(params AposentadoriaPromptParams) (*Prompt, error) {
	return executeTemplate(aposentadoriaTmpl, params)
}

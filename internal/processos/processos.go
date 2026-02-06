package processos

import (
	"bytes"
	"html/template"
)

type Assinatura struct {
	CPF  string
	Nome string
}

type Documento struct {
	Numero      string
	Tipo        string
	Assinaturas []Assinatura
}

type Analista struct {
	ID   int64
	CPF  string
	Nome string
}

const prompt = `
Lista de Documentos:
{{range .Documentos}}
<documento>
	Numero: {{.Numero}}
	Tipo: {{.Tipo}}
	{{- with .Assinaturas}}
	Assinaturas: 
	{{- range .}}
		- {{.Nome}} ({{.CPF}})
	{{- end}}
	{{- end}}
</documento>
{{end}}

Lista de Analistas:
{{range .Analistas}}
<analista>
	ID: {{.ID}}
	Nome: {{.Nome}}
	CPF: {{.CPF}}
</analista>
{{end}}
`

// NewPrompt cria um novo prompt de análise do documentos de um processo para IA.
func NewPrompt(docs []Documento, analistas []Analista) (string, error) {
	tmpl, err := template.New("").Parse(prompt)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, map[string]any{
		"Documentos": docs,
		"Analistas":  analistas,
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

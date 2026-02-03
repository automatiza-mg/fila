package mail

import (
	"bytes"
	"embed"
	"html/template"
)

var (
	//go:embed templates
	fs embed.FS

	setupTmpl = template.Must(template.ParseFS(fs, "templates/cadastro.tmpl"))
)

// Executa as três partes de um template de Email: subject, text e html.
func executeTemplate(tmpl *template.Template, to []string, data any) (*Email, error) {
	subjectBuf := new(bytes.Buffer)
	err := tmpl.ExecuteTemplate(subjectBuf, "subject", data)
	if err != nil {
		return nil, err
	}

	textBuf := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(textBuf, "text", data)
	if err != nil {
		return nil, err
	}

	htmlBuf := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBuf, "html", data)
	if err != nil {
		return nil, err
	}

	return &Email{
		To:      to,
		Subject: subjectBuf.String(),
		Text:    textBuf.String(),
	}, nil
}

type SetupEmailParams struct {
	// A URL de conclusão de cadastro do usuário.
	SetupURL string
}

// NewSetupEmail retorna um novo [Email] para o template `cadastro.tmpl`, possibilitando a conclusão de cadastro do usuário.
func NewSetupEmail(to string, params SetupEmailParams) (*Email, error) {
	return executeTemplate(setupTmpl, []string{to}, params)
}

package mail

import (
	"bytes"
	"embed"
	"html/template"
)

var (
	//go:embed templates
	fs embed.FS

	setupTmpl      = template.Must(template.ParseFS(fs, "templates/cadastro.tmpl"))
	resetSenhaTmpl = template.Must(template.ParseFS(fs, "templates/reset-senha.tmpl"))
	prioridadeTmpl = template.Must(template.ParseFS(fs, "templates/prioridade.tmpl"))
)

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
	SetupURL string
}

func NewSetupEmail(to string, params SetupEmailParams) (*Email, error) {
	return executeTemplate(setupTmpl, []string{to}, params)
}

type ResetSenhaEmailParams struct {
	ResetURL string
}

func NewResetSenhaEmail(to string, params ResetSenhaEmailParams) (*Email, error) {
	return executeTemplate(resetSenhaTmpl, []string{to}, params)
}

type PrioridadeEmailParams struct {
	NumeroProcesso string
	SolicitacaoURL string
	Justificativa  string
}

func NewPrioridadeEmail(to []string, params PrioridadeEmailParams) (*Email, error) {
	return executeTemplate(prioridadeTmpl, to, params)
}

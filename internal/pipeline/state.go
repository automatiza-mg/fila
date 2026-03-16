package pipeline

import (
	"github.com/google/uuid"
)

// State carrega os dados compartilhados entre as etapas do pipeline.
// Cada step lê e escreve neste estado conforme necessário.
type State struct {
	// ProcessoID é o identificador do processo sendo analisado.
	ProcessoID uuid.UUID

	// LinkAcesso é a URL de acesso externo do processo no SEI.
	LinkAcesso string

	// Status é o status de processamento atual do processo.
	Status string

	// DocumentosListados são os números dos documentos encontrados na página de acesso do SEI.
	DocumentosListados []string

	// DocumentosBuscados são os documentos novos buscados do SEI com conteúdo OCR.
	DocumentosBuscados []DocBuscado
}

// DocBuscado representa um documento SEI com os dados completos da API e o texto extraído via OCR.
type DocBuscado struct {
	Numero       string
	Conteudo     string
	ContentType  string
	Tipo         string
	Unidade      string
	LinkAcesso   string
	MetadadosAPI []byte
}

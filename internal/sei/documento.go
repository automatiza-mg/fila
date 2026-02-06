package sei

import (
	"context"
	"encoding/xml"
)

type ConsultarDocumentoRequest struct {
	XMLName                     xml.Name `xml:"Sei consultarDocumento"`
	SiglaSistema                string
	IdentificacaoServico        string
	ProtocoloDocumento          string
	IdUnidade                   int    `xml:",omitempty"`
	SinRetornarAndamentoGeracao string `xml:",omitempty"`
	SinRetornarAssinaturas      string `xml:",omitempty"`
	SinRetornarPublicacao       string `xml:",omitempty"`
	SinRetornarCampos           string `xml:",omitempty"`
}

type ConsultarDocumentoResponse struct {
	XMLName    xml.Name                 `xml:"Sei consultarDocumentoResponse"`
	Parametros RetornoConsultaDocumento `xml:"parametros"`
}

type UnidadeElaboradora struct {
	IdUnidade string
	Sigla     string
	Descricao string
}

type RetornoConsultaDocumento struct {
	IdProcedimento        string
	ProcedimentoFormatado string
	IdDocumento           string
	DocumentoFormatado    string
	NivelAcessoLocal      int
	NivelAcessoGlobal     int
	LinkAcesso            string
	Serie                 Serie
	Numero                string
	Data                  string
	Descricao             string
	UnidadeElaboradora    UnidadeElaboradora
	Assinaturas           Assinaturas
}

func (c *Client) ConsultarDocumento(ctx context.Context, protocolo string) (*ConsultarDocumentoResponse, error) {
	return doReq[ConsultarDocumentoRequest, ConsultarDocumentoResponse](ctx, c.http, c.cfg.URL, ConsultarDocumentoRequest{
		SiglaSistema:           c.cfg.SiglaSistema,
		IdentificacaoServico:   c.cfg.IdentificacaoServico,
		ProtocoloDocumento:     protocolo,
		SinRetornarAssinaturas: "S",
	})
}

package sei

import (
	"context"
	"encoding/xml"
)

type ListarUnidadesRequest struct {
	XMLName              xml.Name `xml:"Sei listarUnidades"`
	SiglaSistema         string
	IdentificacaoServico string
}

type Unidade struct {
	IdUnidade       string `xml:"IdUnidade" json:"id_unidade"`
	Sigla           string `xml:"Sigla,omitempty" json:"sigla"`
	Descricao       string `xml:"Descricao,omitempty" json:"descricao"`
	SinProtocolo    string `xml:"SinProtocolo,omitempty" json:"sin_protocolo"`
	SinArquivamento string `xml:"SinArquivamento,omitempty" json:"sin_arquivamento"`
	SinOuvidoria    string `xml:"SinOuvidoria,omitempty" json:"sin_ouvidoria"`
}

type ListarUnidadesResponse struct {
	XMLName    xml.Name            `xml:"Sei listarUnidadesResponse"`
	Parametros Parametros[Unidade] `xml:"parametros"`
}

func (c *Client) ListarUnidades(ctx context.Context) (*ListarUnidadesResponse, error) {
	return doReq[ListarUnidadesRequest, ListarUnidadesResponse](ctx, c.http, c.cfg.URL, ListarUnidadesRequest{
		SiglaSistema:         c.cfg.SiglaSistema,
		IdentificacaoServico: c.cfg.IdentificacaoServico,
	})
}

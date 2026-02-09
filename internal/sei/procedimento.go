package sei

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/charmap"
)

var ErrProcessoVazio = errors.New("processo vazio")

type ConsultarProcedimentoRequest struct {
	XMLName                               xml.Name `xml:"Sei consultarProcedimento"`
	SiglaSistema                          string
	IdentificacaoServico                  string
	IdUnidade                             int `xml:",omitempty"`
	ProtocoloProcedimento                 string
	SinRetornarAssuntos                   string
	SinRetornarInteressados               string
	SinRetornarObservacoes                string
	SinRetornarAndamentoGeracao           string
	SinRetornarAndamentoConclusao         string
	SinRetornarUltimoAndamento            string
	SinRetornarUnidadesProcedimentoAberto string
	SinRetornarProcedimentosRelacionados  string
}

type Procedimento struct {
	IdProcedimento        int    `json:"id_procedimento"`
	ProcedimentoFormatado string `json:"procedimento_formatado"`
	Especificacao         string `json:"especificacao"`
	DataAutuacao          string `json:"data_autuacao"`
	LinkAcesso            string `json:"link_acesso"`
}

type ConsultarProcedimentoResponse struct {
	XMLName    xml.Name     `xml:"Sei consultarProcedimentoResponse"`
	Parametros Procedimento `xml:"parametros"`
}

func (c *Client) ConsultarProcedimento(ctx context.Context, protocolo string) (*ConsultarProcedimentoResponse, error) {
	return doReq[ConsultarProcedimentoRequest, ConsultarProcedimentoResponse](ctx, c.http, c.cfg.URL, ConsultarProcedimentoRequest{
		SiglaSistema:          c.cfg.SiglaSistema,
		IdentificacaoServico:  c.cfg.IdentificacaoServico,
		ProtocoloProcedimento: protocolo,
	})
}

// DownloadProcedimento é uma extensão da API do SEI que permite o download do PDF de um processo.
func (c *Client) DownloadProcedimento(ctx context.Context, linkAcesso string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, linkAcesso, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get info: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	formData := make(url.Values)
	doc.Find("#frmProcessoAcessoExternoConsulta input[type='hidden']").Each(func(i int, s *goquery.Selection) {
		name, ok := s.Attr("name")
		if ok {
			val, _ := s.Attr("value")
			formData.Set(name, val)
		}
	})

	var listaIDs []string
	doc.Find("#tblDocumentos tr input[type='checkbox']").Each(func(i int, s *goquery.Selection) {
		if val, ok := s.Attr("value"); ok {
			listaIDs = append(listaIDs, val)
		}
	})

	if len(listaIDs) == 0 {
		return nil, ErrProcessoVazio
	}

	formData.Set("hdnInfraItensSelecionados", strings.Join(listaIDs, ","))
	formData.Set("hdnFlagGerar", "1")

	reqPost, err := http.NewRequestWithContext(ctx, http.MethodPost, linkAcesso, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}
	reqPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resPost, err := http.DefaultClient.Do(reqPost)
	if err != nil {
		return nil, err
	}

	contentType := resPost.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/pdf") {
		resPost.Body.Close()
		return nil, errors.New("invalid content-type")
	}

	return resPost.Body, nil
}

type LinhaDocumento struct {
	Numero  string `json:"numero"`
	Link    string `json:"link"`
	Tipo    string `json:"tipo"`
	Data    string `json:"data"`
	Unidade string `json:"unidade"`
}

// ListarDocumentos retorna a lista de documentos através de um WebScraping da página de acesso externo de um processo
// (link externo).
func (c *Client) ListarDocumentos(ctx context.Context, linkAcesso string) ([]LinhaDocumento, error) {
	res, err := c.http.Get(linkAcesso)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	baseURL := strings.TrimSuffix(c.cfg.URL, "/ws/SeiWS.php")

	dec := charmap.ISO8859_1.NewDecoder()
	rd := dec.Reader(res.Body)

	doc, err := goquery.NewDocumentFromReader(rd)
	if err != nil {
		return nil, err
	}

	documentos := make([]LinhaDocumento, 0)
	doc.Find("#tblDocumentos tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return
		}

		var documento LinhaDocumento
		s.Find("td").Each(func(i int, s *goquery.Selection) {
			if i == 0 {
				return
			}

			switch i {
			case 0:
				return
			case 1:
				link := s.Children().First()
				numero := link.Text()
				href, ok := link.Attr("href")
				if ok {
					documento.Numero = numero
					documento.Link = fmt.Sprintf("%s/%s", baseURL, href)
				}
			case 2:
				documento.Tipo = s.Text()
			case 3:
				documento.Data = s.Text()
			case 4:
				documento.Unidade = s.Text()
			default:
				return
			}
		})

		documentos = append(documentos, documento)
	})

	return documentos, nil
}

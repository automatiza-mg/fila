package sei

import "encoding/xml"

type Assinaturas struct {
	XMLName xml.Name     `xml:"Assinaturas" json:"-"`
	Itens   []Assinatura `xml:"item"`
}

type Assinatura struct {
	Nome        string
	CargoFuncao string
	DataHora    string
	IdUsuario   string
	IdOrigem    string
	IdOrgao     string
	Sigla       string
}

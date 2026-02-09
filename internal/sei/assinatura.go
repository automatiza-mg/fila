package sei

import "encoding/xml"

type Assinaturas struct {
	XMLName xml.Name     `xml:"Assinaturas" json:"-"`
	Itens   []Assinatura `xml:"item"`
}

type Assinatura struct {
	Nome        string `xml:"Nome" json:"nome"`
	CargoFuncao string `xml:"CargoFuncao" json:"cargo_funcao"`
	DataHora    string `xml:"DataHora" json:"data_hora"`
	IdUsuario   string `xml:"IdUsuario" json:"id_usuario"`
	IdOrigem    string `xml:"IdOrigem" json:"id_origem"`
	IdOrgao     string `xml:"IdOrgao" json:"id_orgao"`
	Sigla       string `xml:"Sigla" json:"sigla"`
}

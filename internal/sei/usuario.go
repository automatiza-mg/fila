package sei

type Usuario struct {
	IdUsuario string `xml:"IdUsuario" json:"id_usuario"`
	Sigla     string `xml:"Sigla" json:"sigla"`
	Nome      string `xml:"Nome" json:"nome"`
}

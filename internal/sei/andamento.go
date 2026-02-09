package sei

type Andamento struct {
	IdAndamento    string  `xml:"IdAndamento" json:"id_andamento,omitempty"`
	IdTarefa       string  `xml:"IdTarefa" json:"id_tarefa,omitempty"`
	IdTarefaModulo string  `xml:"IdTarefaModulo" json:"id_tarefa_modulo,omitempty"`
	Descricao      string  `xml:"Descricao" json:"descricao"`
	DataHora       string  `xml:"DataHora" json:"data_hora"`
	Unidade        Unidade `xml:"Unidade" json:"unidade"`
}

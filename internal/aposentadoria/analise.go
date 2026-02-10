package aposentadoria

type Analise struct {
	Aposentadoria    bool   `json:"aposentadoria" jsonschema:"required" jsonschema_description:"Indica se o processo é ou não um pedido de aposentadoria"`
	CPF              string `json:"cpf_requerente" jsonschema:"required" jsonschema_description:"O CPF do requerente da aposentadoria, sem pontos e traços"`
	DataRequerimento string `json:"data_requerimento" jsonschema:"required,format=date" jsonschema_description:"A data em que o requerimento foi enviado, no formato YYYY-MM-DD"`
	DataNascimento   string `json:"data_nascimento_requerente" jsonschema:"required,format=date" jsonschema_description:"A data de nascimento do requerente, no formato YYYY-MM-DD"`
	Judicial         bool   `json:"judicial" jsonschema:"required" jsonschema_description:"Indica se houve pedido judicial para dar início ao processo"`
	Invalidez        bool   `json:"invalidez" jsonschema:"required" jsonschema_description:"Indica se o requerente abriu o processo por invalidez"`
	CPFDiligencia    string `json:"cpf_responsavel_diligencia" jsonschema:"not_required" jsonschema_description:"O CPF do responsável pelo envio da diligência, se houver, sem pontos e traços"`
}

type Documento struct {
	Tipo        string
	Data        string
	Conteudo    string
	Assinaturas []string
}

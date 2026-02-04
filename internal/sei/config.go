package sei

type Config struct {
	URL                  string `env:"SEI_URL,notEmpty"`
	SiglaSistema         string `env:"SEI_SIGLA_SISTEMA,notEmpty"`
	IdentificacaoServico string `env:"SEI_IDENTIFICACAO_SERVICO,notEmpty" json:"-"`
}
